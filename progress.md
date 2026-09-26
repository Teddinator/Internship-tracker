# Systems Engineering Progress

## Current status
**Current week:** 3  
**Current phase:** Linux and Operating Systems  
**Started:** 2026-09-21  
**Target weekly time:** ~10h  
**Primary project:** Internship-tracker

## How to use this file
Update at the end of every study week. Keep entries short but concrete.

## Current Week

### Week 1 — Linux anatomy

**Status:** Completed
**Time:** 5 hours

#### Done
- Ran the API directly on Fedora.
- Inspected the process with `ps`, `/proc` and `lsof`.
- Identified the HTTP and PostgreSQL sockets.
- Tested environment inheritance and permission denial.
- Wrote `docs/learning/linux-process.md`.
- Read How Linux Works, chapters 1–2.
- Can explain the main concepts without notes.

#### Key evidence
```text
PID/PPID: 10467/8828 User: theodor (UID 1000) HTTP: TCP *:8080 (LISTEN) PostgreSQL: TCP [::1]:42210->[::1]:5433 (ESTABLISHED) Open file descriptors: 9 /proc/1/environ: Permission denied, exit code 1
```

### Key takeaway
A Linux process has an identity, inherited environment, virtual address space and file descriptors that refer to open kernel resources.

### Week 2 — Title

**Status:** Completed
**Time:** 4 hours

#### Done
- [x] Main lab
- [x] Failure experiment
- [x] Learning note
- [ ] Remaining task

#### Key evidence
```text
accept4(7, ...) = 8
read(8, "GET /applications HTTP/1.1 ...", 4096) = 90
write(8, "HTTP/1.1 200 OK ...", 2352) = 2352
```

### Key takeaway
System calls are the interface processes use to request services from the Linux kernel.
File descriptors identify resources opened by a process, such as file and sockets. In the lab, I could follow an HTTP request by seeing the API accept a connection, read the request, communicate with PostgreSQL through another file descriptor, and write the HTTP response back to the client.

### Week 3 — Processes and signals

**Status:** In progress

#### Done
- [x] Tested SIGTERM, SIGINT and SIGKILL.
- [x] Tested SIGTERM during active requests of 5 and 12 seconds.
- [x] Observed parent/child processes and a zombie process.
- [x] Wrote `docs/learning/process-signals.md`.
- [ ] Read How Linux Works, chapter 5-6.

#### Key evidence
- 5-second request: `200 OK`; shutdown waited for the response.
- 12-second request: `context deadline exceeded` after 10 seconds; curl reported `Empty reply from server`.
- Zombie: `STAT Z` before `waitpid()`; no process entry afterward.

#### Key takeaway
SIGTERM and SIGINT use the API's graceful shutdown path. SIGKILL stops it immediately. Graceful shutdown waits for active requests only until its deadline expires.

## Completed Weeks

| Week | Status | Key takeaway | Main artifact |
|---|---|---|---|
| 1 | Completed | | |
| 2 | Completed | | |
| 3 | Not started | | |

## Skills Matrix
Scale:
- 0 = not started
- 1 = basic familiarity
- 2 = can use with guidance
- 3 = can use independently
- 4 = can explain trade-offs and debug
- 5 = deep practical understanding

| Area | Level | Evidence |
|---|---:|---|
| Linux processes | 2 | |
| Linux filesystems | 0 | |
| Linux networking | 0 | |
| TCP/IP | 0 | |
| DNS | 0 | |
| HTTP/TLS | 0 | |
| Go concurrency | 0 | |
| Go profiling | 0 | |
| PostgreSQL transactions | 0 | |
| PostgreSQL MVCC | 0 | |
| PostgreSQL indexes/planner | 0 | |
| Docker | 0 | |
| AWS networking | 0 | |
| AWS IAM | 0 | |
| Terraform | 0 | |
| CI/CD | 0 | |
| Observability | 0 | |
| Load testing | 0 | |
| Reliability engineering | 0 | |
| Messaging | 0 | |
| Idempotency | 0 | |
| Outbox pattern | 0 | |
| gRPC | 0 | |
| Distributed systems | 0 | |

## ADR Index
- ADR-001:
- ADR-002:
- ADR-003:

## Questions for future study
-

### Week N — Title

**Status:** In progress / Completed
**Time:** X hours

#### Done
- [x] Main lab
- [x] Failure experiment
- [x] Learning note
- [ ] Remaining task

#### Key evidence
```text
2–5 important measurements or outputs
```

### Key takeaway