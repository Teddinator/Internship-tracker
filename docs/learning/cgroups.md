
# Linux cgroups — Week 10

## Syfte
Förstå hur Linux cgroups v2 används för att mäta och begränsa CPU- och minnesanvändning hos en grupp processer.

## Cgroups och systemd
Cgroups är en funktion i Linux-kärnan som grupperar processer och hanterar deras sammanlagda resursanvändning.

Systemd använder cgroups för att organisera tjänster. Cgroups v2 använder en gemensam hierarki där flera controllers, exempelvis CPU och memory, kan hantera samma grupp.

En tjänsts cgroup kan undersökas i `/sys/fs/cgroup/system.slice/`.

Viktiga filer:
- `cgroup.procs`: processer i gruppen.
- `cpu.max`: CPU-kvot och period.
- `cpu.stat`: CPU-användning och throttling.
- `memory.current`: aktuell minnesanvändning.
- `memory.max`: hård minnesgräns.
- `memory.events`: minnesgräns- och OOM-händelser.

## CPU-experiment
Byggde ett CPU-intensivt Go-program och körde det via en tillfällig systemd-tjänst med och utan `CPUQuota=50%`.

| Mätning | Utan gräns | CPUQuota=50% |
|---|---:|---:|
| Genomsnittlig körningstid | 498 ms | 953 ms |
| Genomsnittlig CPU-tid | 502 ms | 498 ms |

Programmet utförde ungefär samma CPU-arbete, men tog cirka 1,9 gånger längre verklig tid med begränsningen.

### Verifiering med cpu.stat
`cpu.max` visade:

```text
50000 100000
```

Det innebär 50 000 mikrosekunder CPU-tid per 100 000 mikrosekunders period.

```text
nr_periods 607
nr_throttled 607
throttled_usec 30209398
```

Samtliga 607 observerade kvotperioder hade throttling.

## Minnesexperiment
Byggde ett Go-program som allokerade 10 MiB i taget och skrev till minnessidorna.

Körde det med:

```bash
sudo systemd-run \
  --unit=cgroup-memory-test \
  -p MemoryMax=50M \
  -p MemorySwapMax=0 \
  /usr/local/bin/cgroup-memory
```

Programmet rapporterade 40 MiB allokerat innan Linux dödade processen.

Systemd-loggar:

```text
The kernel OOM killer killed some processes in this unit.
Main process exited, code=killed, status=9/KILL
Failed with result 'oom-kill'.
```

Observerade även att `memory.current` ökade från cirka 11,6 till 42,2 MiB under ett nytt test.

`memory.events` visade inga OOM-händelser i de avläsningar som hann göras före processens avslut.

## Key takeaway
Cgroups begränsar och mäter resurser för hela processgrupper.

CPUQuota begränsar CPU-tid per period och kan leda till throttling, vilket ökar körningstiden utan att processen behöver avslutas.

MemoryMax sätter en hård minnesgräns. När Linux inte kan återvinna tillräckligt med minne kan OOM-kill avsluta en process.

Minnesgränsen gäller hela cgroupens redovisade minnesanvändning, inte bara programmets egna allokeringar.
