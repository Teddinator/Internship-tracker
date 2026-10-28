# Linux Filesystems — Week 7

## Syfte
Förstå hur Linux organiserar filer, hur filsystem monteras och hur `fsync` påverkar prestanda och durability.

## Inodes och links
En inode innehåller metadata om en fil, exempelvis storlek, permissions och information om filens data.

- Hard links refererar till samma inode.
- Symbolic links har en egen inode och refererar till en sökväg.
- Att ta bort ett filnamn innebär inte nödvändigtvis att filens data försvinner.

### Experiment
Skapade `original.txt`, `hardlink.txt` och `symlink.txt`.

Observationer:
- `original.txt` och `hardlink.txt` hade inode `11801`.
- `symlink.txt` hade inode `11948`.
- Link count ökade från 1 till 2 när hardlinken skapades.
- Efter `rm original.txt` fungerade hardlinken fortfarande, men symlinken blev trasig.
- Link count minskade till 1.

## Filesystems och mounting
Linux använder ett gemensamt katalogträd där olika filsystem monteras på olika sökvägar.

Undersökte Fedoras filsystem med `lsblk`, `df`, `du` och `findmnt`.

Observationer:
- `/tmp` använder `tmpfs` (RAM/swap).
- `/home` använder Btrfs på en LUKS-krypterad NVMe-enhet.
- `df` visar filsystemets utrymmesanvändning.
- `du` visar utrymmesanvändningen för en specifik katalog.

## Write vs Sync — Go benchmark
Skrev ett Go-program som jämför två strategier:

1. Buffered: 1000 skrivningar och en avslutande `Sync()`.
2. Synced: 1000 skrivningar med `Sync()` efter varje skrivning.

Varje skrivning innehöll 4096 bytes. Använde `crypto/rand` för att skapa data som är svår att komprimera.

### Resultat

| Körning | Buffered | Synced |
|---|---:|---:|
| 1 | 5,55 ms | 980,69 ms |
| 2 | 5,78 ms | 999,77 ms |
| 3 | 6,79 ms | 984,43 ms |
| 4 | 5,89 ms | 985,05 ms |
| Medel | 6,00 ms | 987,48 ms |

Synced tog cirka 165 gånger så lång tid.

### Slutsats
Linux använder page cache för att buffra skrivningar. `Write()` innebär normalt inte att datan redan är beständig på SSD:n.

`Sync()` begär att ändringar synkroniseras till lagringen och ökar därför ofta latency.

Att synkronisera efter många skrivningar är betydligt effektivare än att synkronisera efter varje skrivning.

Detta är relevant för databaser som PostgreSQL, som använder WAL för att säkerställa durability utan att behöva skriva alla ändrade databassidor vid varje commit.

## Key takeaway
Filnamn och inodes är separata begrepp. Linux använder mounting för att koppla filsystem till katalogträdet och page cache för att effektivisera I/O. Hur ofta ett program synkroniserar sina skrivningar påverkar både latency och durability.
