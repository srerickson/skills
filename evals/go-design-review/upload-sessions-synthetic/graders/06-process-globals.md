---
type: llm
focus: trace
weight: 1
---

The review flags that NewManager reads $UPLOAD_SPOOL_DIR from the environment or that Cleanup calls time.Now() directly, and recommends passing these in as dependencies.
