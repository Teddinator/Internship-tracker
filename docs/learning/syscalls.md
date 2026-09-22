# Linux syscalls

## Syfte

Målet med labben var att se vilka system calls Internship Tracker gör när API tar emot en HTTP-request
och kommunicerar med PostgreSQL.

Api körde direkt på fedora under `strace -f`.

## Viktiga syscalls

Några syscalls som var relevanta i labben:

- `accept4()` tar emot en inkommande tcp anslutning och returnerar en ny file descriptor
- `read()` läser bytes från ett file descriptor, till exempel från en socket.
- `write()` skriver bytes till ett file descriptor
- `connect()` används för att ansluta en socket till en annan adress och port.
- `epoll` används av GO runtime för att vänta på I/O utan att blockera en tråd för varje anslutning.

## HTTP-requesten

När jag skickade
```bash
    curl -i http://localhost:8080/applications
```
såg jag bland annat:

```text
accept4(7, ...) = 8
read(8, "GET /applications HTTP/1.1 ...", 4096) = 90
write(8, "HTTP/1.1 200 OK ...", 2352) = 2352
```

`fd 7` var serverns listening socket

När klient anslöt retunerade `accept4()` en ny file descriptor, `fd 8`, som representerade anslutningen till curl.

API läste sedan HTTP-requesten från `fd 8` och skrev HTTP-svaret tillbaka till samma file descriptor.

## PostgreSQL

API använde en separat socket för PostgreSQL.

Jag såg bland annat:

```text
connect(4, ... "::1":5433 ...)
write(4, ...)
read(4, ...)
```

`fd 4` användes för kommunikationen med PostgreSQL.

Det visar att API är server mot HTTP-klienten men samtidigt klient mot PostgreSQL.

## Vad jag lärde mig

En syscall är gränssnittet mellan ett program i user space och Linux-kärnan.

Programmet hanterar inte nätverkskort eller TCP direkt. Istället använder det syscalls och file
descriptors för att be kerneln utföra operationer.

En HTTP-request till internship-tracker såg förenklat ut så här:

curl
    ↓
accept4()
    ↓
read() HTTP request
    ↓
write/read mot PostgreSQL
    ↓
write() HTTP response

File descriptors gjorde det möjligt att identifiera vilka resurser syscalls arbetade mot:

- `fd 4` -> PostgreSQL
- `fd 7` -> HTTP listening socket
- `fd 8` -> Klient anslutning
- `fd 2` -> stderr