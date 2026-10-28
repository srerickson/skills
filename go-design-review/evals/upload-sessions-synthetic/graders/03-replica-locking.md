---
type: llm
focus: trace
weight: 1
---

The review identifies that the in-process per-session mutex does not serialize operations across the three replicas sharing the NFS spool, so concurrent requests on different replicas can lose updates or race.
