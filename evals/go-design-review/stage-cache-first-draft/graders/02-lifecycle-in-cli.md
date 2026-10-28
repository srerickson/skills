---
type: llm
focus: trace
weight: 1
---

The review identifies that stage-file persistence or the commit lifecycle (finalize, save, apply, remove the stage file) is handled in the run (CLI) package, and recommends that the stage package own it.
