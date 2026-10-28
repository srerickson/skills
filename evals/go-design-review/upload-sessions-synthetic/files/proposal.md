# Proposal: resumable uploads for the archive web server

Users upload large files (up to 50 GB) to a collection over flaky
connections. Uploads must survive a dropped connection or a server restart:
the client sends numbered parts, asks which parts the server has, and tells
the server when it's done. The server runs as three replicas behind a load
balancer; the spool and collection directories are on a shared NFS mount.

## HTTP API

- `POST /collections/{c}/uploads` with `{"name": "...", "size": N,
  "sha256": "..."}` returns `{"id": "..."}`.
- `GET /collections/{c}/uploads/{id}` returns the parts received so far.
- `PUT /collections/{c}/uploads/{id}/parts/{n}` with the part bytes.
- `PATCH /collections/{c}/uploads/{id}` with `{"name": "..."}` renames the
  upload before it's finished.
- `POST /collections/{c}/uploads/{id}/finish`.
- `DELETE /collections/{c}/uploads/{id}` aborts.

## Package `internal/upload`

```go
type Part struct {
	N      int
	Size   int64
	SHA256 string
}

// Session is an upload in progress.
type Session struct {
	ID         string
	Collection string
	Name       string
	Size       int64
	SHA256     string
	Parts      map[int]Part
	Created    time.Time
}

// Store persists sessions and part files.
type Store interface {
	Load(ctx context.Context, id string) (*Session, error)
	Save(ctx context.Context, s *Session) error
	Remove(ctx context.Context, id string) error
	WritePart(ctx context.Context, id string, n int, r io.Reader) (Part, error)
	OpenPart(ctx context.Context, id string, n int) (io.ReadCloser, error)
	RemoveParts(ctx context.Context, id string) error
	List(ctx context.Context) ([]*Session, error)
}

// Manager implements the upload lifecycle. It is safe for concurrent use:
// a per-session mutex serializes operations on the same session.
type Manager struct {
	store       Store
	collections string
	mu          sync.Mutex
	locks       map[string]*sync.Mutex
}

// NewManager returns a Manager using the spool directory in
// $UPLOAD_SPOOL_DIR and the collections root dir.
func NewManager(dir string) (*Manager, error)

func (m *Manager) Start(ctx context.Context, collection, name string, size int64, sha string) (*Session, error)

// Session returns the session for id, for reading or updating with Save.
func (m *Manager) Session(ctx context.Context, id string) (*Session, error)
func (m *Manager) Save(ctx context.Context, s *Session) error

// PutPart stores part n: it loads the session, writes the part with
// WritePart, adds it to Parts, and saves the session.
func (m *Manager) PutPart(ctx context.Context, id string, n int, r io.Reader) error

// Finish assembles the parts into a temp file in the collection, checks
// the size and SHA-256, renames the temp file to its final name, then
// removes the session and its parts. Errors describe what went wrong, e.g.
// "missing part 3", "checksum mismatch", or "file exists".
func (m *Manager) Finish(ctx context.Context, id string) error

// Abort removes the session and its parts. It refuses (returns an error)
// if the final file already exists, so a finished upload isn't disturbed.
func (m *Manager) Abort(ctx context.Context, id string) error

// Cleanup removes sessions whose Created time is more than 7 days before
// time.Now(). It's called hourly from a goroutine started in main.
func (m *Manager) Cleanup(ctx context.Context) error
```

`NewFileStore(dir string) Store` is the only `Store` implementation. It
saves each session as `<id>.json` (written to a temp file, then renamed)
and parts as `<id>.<n>.part`.

## Handlers (`internal/server`)

```go
func (h *Handler) rename(w http.ResponseWriter, r *http.Request) {
	s, err := h.uploads.Session(ctx, id)
	// ...
	s.Name = body.Name
	h.uploads.Save(ctx, s)
}

func (h *Handler) finish(w http.ResponseWriter, r *http.Request) {
	err := h.uploads.Finish(ctx, id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusCreated)
	case strings.Contains(err.Error(), "missing part"),
		strings.Contains(err.Error(), "checksum"):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case strings.Contains(err.Error(), "exists"):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

The other handlers call the matching `Manager` method.
