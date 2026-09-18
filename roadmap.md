# Systems Engineering Roadmap — 60 Weeks

## Goal

Build from backend engineering into infrastructure, production systems, observability, distributed applications, and eventually distributed systems.

Primary lab project: `Internship-tracker`

Expected time: ~10 hours/week.

Recommended weekly split:
- 2h documentation/theory
- 2h book reading
- 4h implementation
- 2h debugging, experiments, and notes

The goal is not to collect tools. The goal is to understand how systems behave, fail, recover, and scale.

## Phase 1 — Linux and Operating Systems (Weeks 1–10)

### Week 1 — Linux anatomy
Read: processes, `/proc`, users/permissions, file descriptors.
Implement: run Internship Tracker without Docker; inspect `ps`, `/proc/<pid>/status`, `fd/`, `maps`, `environ`, and `lsof -p`.
Definition of Done: explain PID, PPID, file descriptors, environment, and what resources the API process has open.
Book: How Linux Works, ch. 1–2.

### Week 2 — Syscalls
Read: user space/kernel space, `open`, `read`, `write`, `accept`, `connect`.
Implement: run API under `strace -f`; send one HTTP request; identify socket creation, accept, DB network I/O, and stdout/stderr writes.
Definition of Done: explain what a syscall is and point to concrete syscalls caused by one API request.
Book: How Linux Works, ch. 3.

### Week 3 — Processes and signals
Read: process lifecycle, fork/exec, signals.
Implement: send SIGTERM, SIGINT, SIGKILL; test graceful shutdown with a slow in-flight request.
Definition of Done: explain graceful shutdown vs forced termination.
Book: How Linux Works, ch. 5–6.

### Week 4 — CPU and scheduling
Read: threads, scheduling, context switching, load average.
Implement: CPU-bound Go lab with 1/2/4/20/100 goroutines; measure with `top`, `pidstat`, `time`; test `GOMAXPROCS`.
Definition of Done: explain why more goroutines do not always increase throughput.

### Week 5 — Memory
Read: virtual memory, RSS/VSZ, stack/heap, page faults.
Implement: Go program that allocates memory; inspect `/proc/<pid>/maps`, `pmap`, `vmstat`; capture basic pprof heap profile.
Book: OSTEP processes/address spaces.

### Week 6 — Filesystems
Read: inodes, directory entries, page cache, fsync, permissions.
Implement: `stat`, `df`, `du`, `lsblk`; hard/symlinks; Go `Write()` + `Sync()` latency comparison.
Book: How Linux Works, ch. 4.

### Week 7 — I/O
Read: blocking/buffered I/O, disk I/O.
Implement: many PostgreSQL inserts; compare single inserts vs batched transaction; observe with `iostat`/`iotop`.
Definition of Done: distinguish CPU-bound vs I/O-bound workloads.

### Week 8 — systemd
Read: units, services, restart policies, journald.
Implement: `internship-tracker.service`, non-root user, EnvironmentFile, restart policy, journald logs.
Book: How Linux Works, ch. 6–7.

### Week 9 — cgroups
Read: CPU/memory resource control and containers.
Implement: enforce CPU/memory limits; intentionally exceed a limit; inspect cgroup data.
Definition of Done: explain how resource limits are enforced.

### Week 10 — OS checkpoint
Write `docs/learning/request-lifecycle.md` explaining browser → socket → kernel → Go runtime → handler → DB socket → PostgreSQL → response.

## Phase 2 — Networking (Weeks 11–18)

### Week 11 — IP and routing
Read: IPv4, CIDR, routing, ARP.
Implement: `ip addr`, `ip route`, `ip neigh`; create two network namespaces with veth pair.
DoD: explain packet routing between namespaces.

### Week 12 — TCP
Read: handshake, seq/ack, retransmission, teardown.
Implement: TCP client/server in Go; capture SYN/SYN-ACK/ACK and FIN/RST.
Book: Computer Networking: A Top-Down Approach.

### Week 13 — TCP failure
Read: timeout, connection refused/reset, packet loss, jitter.
Implement: use `tc netem` where available; observe behavior in Go client.
DoD: explain timeout vs refused vs reset.

### Week 14 — DNS
Read: recursive resolution, A/AAAA/CNAME/MX, TTL.
Implement: `dig +trace`; inspect container DNS and `/etc/resolv.conf`.

### Week 15 — HTTP
Read: request/response semantics, keep-alive, headers/status.
Implement: `curl -v`; observe reuse; build a tiny HTTP parser over `net.Conn`.

### Week 16 — TLS
Read: certs, CAs, handshake, asymmetric/symmetric crypto.
Implement: Caddy/Nginx in front of API; inspect with `openssl s_client`.

### Week 17 — Reverse proxy and load balancing
Implement: run two API instances behind Nginx/HAProxy; add temporary instance ID logging; kill one during traffic.

### Week 18 — Docker networking
Implement: explicit frontend/backend Docker networks; API on both, Postgres only backend; prove frontend-only container cannot reach DB.

## Phase 3 — Go Systems and Backend Engineering (Weeks 19–26)

### Week 19 — Goroutines
Worker-pool lab; benchmark 1/5/20/100 workers and document saturation.

### Week 20 — Synchronization
Create a race and fix using mutex, atomic, and channel ownership. Use `go test -race`.

### Week 21 — Deadlocks
Create a two-lock deadlock; fix with lock ordering.

### Week 22 — Context
Create a slow endpoint; short client timeout; prove cancellation reaches downstream work/DB.

### Week 23 — Graceful shutdown under load
Use existing shutdown implementation; run many slow requests; SIGTERM; measure completed vs failed.

### Week 24 — Profiling
Use benchmarks and pprof CPU/heap; find and improve one real bottleneck with before/after evidence.

### Week 25 — Architecture refactor
Choose one domain, preferably applications. Target: Handler → Service → Store → PostgreSQL.

### Week 26 — Testing
Unit-test Service with fake Store; integration-test PostgresStore against real PostgreSQL; include an error/concurrency path.

## Phase 4 — PostgreSQL Deep Dive (Weeks 27–36)

### Week 27 — Transactions
Map current idempotent-create transaction; inject failure mid-flow; prove rollback.

### Week 28 — Isolation
Two psql sessions; compare Read Committed vs Serializable; trigger serialization failure and handle retry in a lab.

### Week 29 — MVCC
Long-running transaction + updates from another session; inspect snapshots/dead tuples/VACUUM stats.

### Week 30 — Locks
Use `pg_locks`; create blocking row lock and deadlock; fix with lock ordering.

### Week 31 — Large dataset generator
Generate 500k–1M realistic applications; measure import time, DB size, baseline latency.

### Week 32 — Indexes
Benchmark four real filters with `EXPLAIN (ANALYZE, BUFFERS)` before/after indexing; document when index is not used.

### Week 33 — Query planner
Find a poor row estimate; run ANALYZE; change distribution/statistics and compare.

### Week 34 — Pagination
Implement offset pagination; benchmark shallow/deep pages; implement keyset pagination on `(created_at,id)`.

### Week 35 — Connection pooling
Constrain DB max connections; load test; measure pool waits; tune max open/idle/lifetime.

### Week 36 — Idempotency deep dive
Run 50 concurrent same-key requests; same key/different payload; simulate commit-success/response-loss retry; define endpoint invariants.

## Phase 5 — AWS and Terraform (Weeks 37–46)

### Week 37 — IAM
Users, roles, trust vs permission policies, least privilege. Build a restricted deployment role.

### Week 38 — VPC
Design and manually build: one VPC, 2 public + 2 private subnets, 2 AZs, routing, IGW, NAT where needed.

### Week 39 — Security Groups
App SG + DB SG; only app SG can reach PostgreSQL; verify allowed/denied paths.

### Week 40 — RDS
Private PostgreSQL; run migrations; app connects; laptop cannot connect directly.

### Week 41 — ECR + compute
Push production image to ECR and run same image on AWS compute.

### Week 42 — ALB + replicas
Two app replicas behind ALB; kill one; observe health checks and continued service.

### Week 43 — Terraform fundamentals
Create `infra/`; define VPC/subnets/routes/SGs; add fmt/validate to CI.

### Week 44 — Terraform modules
Use sensible modules such as network/database/app; avoid over-abstraction.

### Week 45 — State and drift
Remote state; create deliberate manual drift; inspect `terraform plan`.

### Week 46 — CI/CD to AWS
Pipeline: test → build image → ECR → deploy. Use GitHub OIDC. Test rollback.

## Phase 6 — Observability and Reliability (Weeks 47–52)

### Week 47 — Structured logging
Add request ID, method, path, status, duration, useful error context; avoid secrets.
Optional reading: DDIA Batch Processing.

### Week 48 — Metrics
Instrument request count/error count/latency histogram/DB pool metrics; calculate p50/p95/p99.
Optional reading: DDIA Stream Processing.

### Week 49 — Tracing
OpenTelemetry across HTTP → handler → service → PostgreSQL; inject latency and find it via trace.

### Week 50 — SLI/SLO
Define learning SLOs for availability and latency; build measurements; intentionally violate them.

### Week 51 — Load test
Ramp concurrency; collect CPU, memory, DB pool, latency, error rate; identify first bottleneck with evidence.

### Week 52 — Failure testing
Test API replica death, DB unavailable, DB pool exhaustion, injected latency, SIGTERM under load. Create a failure matrix.

## Phase 7 — Distributed Application Patterns (Weeks 53–60)

### Week 53 — Worker architecture
Add `cmd/worker`; use a natural background job such as follow-up reminder/notification.

### Week 54 — SQS
Produce work and consume in worker; kill worker mid-processing and observe redelivery.

### Week 55 — Idempotent consumer
Implement deduplication; deliver same message repeatedly; prove one business side effect.

### Week 56 — Retry/backoff/DLQ
Fake dependency fails; add backoff; route poison message to DLQ; create replay command.

### Week 57 — Transactional outbox
Write domain change + outbox event in one DB transaction; separate publisher; simulate crash after commit before publish.

### Week 58 — gRPC
Create a small Notification Service; define protobuf; implement unary RPC + health endpoint.

### Week 59 — Deadlines and retries
Use short client deadline and intentionally slow server; observe deadline exceeded and server cancellation; test retries and explain idempotency.

### Week 60 — Capstone
Document full system: LB → replicas → Postgres → outbox → queue → workers → gRPC service → observability.
For every boundary: timeout? retry? idempotent? backpressure? consistency? failure mode?
Write 5 ADRs and 1 postmortem.

## Optional continuation — Weeks 61+
- MapReduce
- replication
- consistency models
- logical clocks
- leader election
- Raft
- replicated KV store
- partition/failure injection
- MIT 6.5840 labs

## Weekly workflow
1. Read the week's goals.
2. Create/update issue/task list.
3. Do theory first.
4. Implement.
5. Break the system intentionally.
6. Record evidence.
7. Update `progress.md`.

If one week needs two calendar weeks, take two. Depth matters more than schedule.
