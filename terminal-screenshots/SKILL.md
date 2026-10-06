---
name: terminal-screenshots
description: Capture high-resolution, pixel-perfect screenshots of terminal sessions, CLI commands, interactive TUI apps, and Opencode sessions as PNG, SVG, or WebP images. Use this skill whenever the user asks to take a screenshot, picture, capture, or render of the terminal, CLI output, a tmux pane, or wants to stage/capture an Opencode session.
---

# Terminal Screenshots

A general skill for generating clean, publication-ready terminal screenshots in headless environments using `tmux` buffer capture and Charmbracelet's `freeze` renderer.

Supported formats: **PNG**, **SVG** (scalable vector), and **WebP**.

---

## Tool: `capture-terminal`

The CLI tool is located at `scripts/capture_terminal.py` and is available globally as `capture-terminal` (with `capture-opencode` as an alias).

### 1. Capturing CLI Commands
Run any shell command and capture its stdout/stderr with full ANSI syntax highlighting and colors:

```bash
capture-terminal --command "git log --oneline --graph --decorate -n 10" -o ~/screenshots/git_log.png
```

Vector format:
```bash
capture-terminal --command "ls -lah --color=always /" -o ~/screenshots/ls.svg
```

### 2. Capturing an Active Tmux Pane
Capture what is currently visible on any existing tmux pane:

```bash
# Capture the active pane
capture-terminal --current -o ~/screenshots/active_pane.png

# Capture a specific pane target
capture-terminal --pane "my-session:0.0" -o ~/screenshots/pane.png

# Include scrollback history (e.g. last 100 lines)
capture-terminal --current --scrollback 100 -o ~/screenshots/history.png
```

### 3. Capturing Interactive TUI Apps
Launch an interactive console program (e.g., `htop`, `vim`, `lazygit`, `ranger`) headlessly, let it draw, snapshot it, and cleanly exit:

```bash
capture-terminal --app "htop" --delay 1.5 -o ~/screenshots/htop.png
```

### 4. Piped Input
Pipe raw ANSI text or existing tmux captures directly into `capture-terminal`:

```bash
tmux capture-pane -e -p | capture-terminal -o ~/screenshots/custom.png
```

---

## Opencode Sessions

The tool includes dedicated workflows for capturing the **Opencode** CLI agent.

### Staging a Prompt in a Fresh Session
To demonstrate or photograph a prompt typed into Opencode's input buffer without executing it:

```bash
capture-terminal --prompt "Your prompt text goes here" -o ~/screenshots/staged_prompt.png
```
* Spawns a temporary headless tmux session running Opencode.
* Waits for the Opencode TUI and greeting to initialize.
* Types the exact prompt into the input area without pressing `Enter`.
* Captures the full 24-bit truecolor interface with terminal window chrome.
* Automatically tears down the temporary session.

### Capturing the Active Opencode Conversation
To take a screenshot of the current Opencode conversation without capturing the command or the spinner:

```bash
capture-terminal --current --delay 2.0 -o ~/screenshots/conversation.png
```
The `--delay 2.0` gives Opencode time to finish generating, hide the tool execution widget, and return to an idle state before the capture runs.

### Tmux Instant Hotkey (`Ctrl+B` then `S`)
In `~/.tmux.conf`:
```tmux
bind-key S run-shell "mkdir -p ~/screenshots && tmux capture-pane -e -p | /home/coder/.pixi/bin/freeze --window -o ~/screenshots/opencode_$(date +%Y%m%d_%H%M%S).png && tmux display-message 'Screenshot saved to ~/screenshots/'"
```
* Press `Ctrl+B` followed by `S`.
* Completely invisible: no keystrokes or commands are sent to the terminal buffer.
* Saves immediately to `~/screenshots/`.

---

## Command Reference

| Flag | Description | Default |
|---|---|---|
| `-x`, `--command "<cmd>"` | Execute command and screenshot output | None |
| `--app "<cmd>"` | Run interactive TUI app headlessly and screenshot | None |
| `--current` | Capture active tmux pane | `True` (if no other mode set) |
| `--pane "<id>"` | Specific tmux pane identifier to capture | `$TMUX_PANE` |
| `--prompt "<text>"` | Type prompt in fresh Opencode session (implies `--opencode`) | None |
| `--opencode` | Launch a fresh Opencode session | `False` |
| `-o`, `--output "<path>"` | Destination file (`.png`, `.svg`, `.webp`) | `~/screenshots/term_<timestamp>.png` |
| `--delay <sec>` | Wait N seconds before taking snapshot | `0.0` |
| `--width <cols>` | Virtual terminal width | `120` |
| `--height <rows>` | Virtual terminal height | `34` |
| `--no-window` | Disable macOS-style window chrome buttons | `False` |
| `--scrollback <n>` | Lines of scrollback to include | `0` |

---

## References

- **freeze**: Charmbracelet, Inc. *freeze: Generate images of code and terminal output*. <https://github.com/charmbracelet/freeze>
- **tmux**: Nicholas Marriott et al. *tmux: A terminal multiplexer*. <https://github.com/tmux/tmux>

