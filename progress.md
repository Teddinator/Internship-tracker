# Systems Engineering Progress

## Current status
**Current week:** 7  
**Current phase:** Linux and Operating Systems  
**Started:** 2026-09-21  
**Target weekly time:** ~10h  
**Primary project:** Internship-tracker

## How to use this file
Update at the end of every study week. Keep entries short but concrete.

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
- [X] Read How Linux Works, chapter 5-6.

#### Key evidence
- 5-second request: `200 OK`; shutdown waited for the response.
- 12-second request: `context deadline exceeded` after 10 seconds; curl reported `Empty reply from server`.
- Zombie: `STAT Z` before `waitpid()`; no process entry afterward.

#### Key takeaway
SIGTERM and SIGINT use the API's graceful shutdown path. SIGKILL stops it immediately. Graceful shutdown waits for active requests only until its deadline expires.

### Week 4 — CPU and scheduling

**Status:** Completed  
**Time:** ~4 hours

#### Done
- [x] Built a CPU-bound Go lab.
- [x] Tested 1, 2, 4, 20 and 100 goroutines.
- [x] Tested `GOMAXPROCS` from 1 to 16.
- [x] Measured CPU usage with `time`, `top` and `pidstat`.
- [x] Wrote `docs/learning/cpu-scheduling.md`.
- [X] Read How Linux Works, chapter 8.1–8.5.

#### Key evidence
- `GOMAXPROCS=1` → real 1.23s, CPU ~100%
- `GOMAXPROCS=8` → real 0.16s, CPU ~800%
- `GOMAXPROCS=16` → real 0.17s, CPU ~1480%

### Key takeaway
Goroutines are scheduled by the Go runtime onto OS threads, which Linux schedules onto logical CPUs. More goroutines can increase parallelism, but CPU-bound performance is ultimately limited by available hardware.

### Week 5 — Memory

**Status:** Completed
**Time:** 3,5 hours

#### Done
- [x] Main lab
- [x] Failure experiment
- [x] Learning note
- [ ] Remaining task

#### Key evidence
| Mätning | VSZ | RSS |
|---|---:|---:|
| Före allokering | ~1198 MiB | ~2,3 MiB |
| Efter allokering | ~1456 MiB | ~3,1 MiB |
| Efter skrivning | ~1456 MiB | ~203,1 MiB |

**Förväntade minor page faults:** 51 200  
**Observerad ökning:** 51 202  
**Major page faults:** 0

```text
Type: inuse_space
Total: 201.50 MB

flat: 200 MB
flat%: 99.25%
function: main.main
```

### Key takeaway
Att Varje process har eget minnesutrymme, större än det har på ram och kan få använda ram genom kerneln som allokerar minne från applikationen till ram genom page faults.
Om det inte redan finns i ram. Virtuellt utrymme har heap och stack, heap är objekt och andra objekt som kan behövs för programmet. Allokeras manuellt eller av garbage collector. Stack är sekventiell och hanterar bara funktioner och lokala variabler som behövs för att köra programmet.

### Week 6 — Computer architecture and CPU-cache
Status: Completed
Time: 2.5h
Done
- [x] Läste utvalda delar av CS:APP kapitel 6 om minneshierarki, SRAM/DRAM och minnesåtkomst.
- [x] Byggde Go-benchmarks för sekventiell och slumpmässig läsning av en slice.
- [x] Testade fem datastorlekar med go test -bench=. -benchmem -count=3.
- [x] Skrev docs/learning/cpu-cache.md.
- [X] Verifiera med ett test att båda funktionerna ger samma summa.
- [ ] Bonus: undersök cache misses med perf om möjligt.
#### Key evidence
Antal element	Sekventiellt	Slumpmässigt
1 000	488 ns/op	266 ns/op
10 000	4 932 ns/op	4 599 ns/op
100 000	48 208 ns/op	78 471 ns/op
1 000 000	485 825 ns/op	912 046 ns/op
10 000 000	5 108 143 ns/op	50 557 455 ns/op


### Key takeaway
Sekventiell minnesåtkomst utnyttjar spatial locality och CPU-cache effektivt. För 10 miljoner element tog slumpmässig indexering nästan 10 gånger så lång tid. En del av skillnaden kan också bero på den extra index-slicen, så resultatet är inte ett direkt mått på cache misses.

### Week 7 — Filesystems

**Status:** Completed  
**Time:** 3

#### Done
- [x] Read How Linux Works, chapter 4 (selected sections).
- [x] Investigated inodes, hard links and symbolic links.
- [x] Tested broken symlink after deleting original file.
- [x] Inspected filesystems using stat, lsblk, df, du and findmnt.
- [x] Built Go benchmark comparing Write and Sync.
- [x] Wrote docs/learning/filesystems.md.

#### Key evidence

- Hard links shared inode 11801; symlink used inode 11948.
- Removing original.txt reduced link count from 2 to 1.
- /tmp used tmpfs; /home used Btrfs.
- Buffered: ~6 ms (1000 writes + 1 sync).
- Synced: ~987 ms (1000 writes + 1000 sync).
- Synced was ~165x slower.

### Key takeaway
Linux separates filenames from file identity through inodes. Filesystems are mounted into a shared directory tree. Page cache makes writes efficient, while fsync provides stronger durability guarantees at the cost of latency.


## Completed Weeks

| Week | Status | Key takeaway | Main artifact |
|---|---|---|---|
| 1 | Completed | | |
| 2 | Completed | | |
| 3 | Completed | | |
| 4 | Completed | | |
| 5 | Completed | | |
| 6 | Completed | | |
| 7 | Not started | | |

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