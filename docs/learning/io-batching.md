
# PostgreSQL I/O and Batching — Week 8

## Syfte
Förstå hur batching påverkar PostgreSQL:s skrivprestanda och hur WAL, commits och disk-I/O hänger ihop.

## PostgreSQL och WAL
PostgreSQL använder shared buffers för databassidor i RAM och WAL (Write-Ahead Log) för att kunna återställa ändringar efter en krasch.

Vid en vanlig durable commit måste relevant WAL synkroniseras till beständig lagring. Själva databassidorna kan skrivas till SSD:n senare, exempelvis vid checkpoints.

Index kräver underhåll vid INSERT, medan constraints kontrollerar att datan följer databasens regler.

## Benchmark — Go och PostgreSQL
Skapade tabellen `io_benchmark` och jämförde tre strategier för att spara 1 000 rader:

1. Individual INSERTs: 1 000 statements och 1 000 commits.
2. Multi-row INSERTs: 10 statements med 100 rader vardera.
3. Transaction batches: 1 000 statements grupperade i 10 transaktioner.

Använde Go med `database/sql` och pgx.

### Resultat

| Metod | Tid | Throughput |
|---|---:|---:|
| Individual INSERTs | 1069,85 ms | 935 rows/s |
| Multi-row INSERTs | 15,08 ms | 66299 rows/s |
| Transaction batches | 86,13 ms | 11610 rows/s |

Multi-row INSERTs var cirka 71 gånger snabbare än individuella INSERTs. Transaktionsbatcher var cirka 12 gånger snabbare.

## Disk-I/O med iostat
Använde `lsblk` för att identifiera Fedoras NVMe-enhet (`nvme1n1`).

Observerade diskaktivitet med:

`iostat -d -y -k nvme1n1 2 5`

Högsta observerade skrivhastighet var 64818 kB/s med 5322,5 I/O-operationer per sekund.

Mätningen visar aktivitet på hela SSD:n och kan inte isolera PostgreSQL:s skrivningar.

## Failure experiment — Rollback
Startade en transaktion och körde två INSERTs.

Den första lyckades, men den andra bröt mot `NOT NULL` eftersom `status` var NULL.

Efter `ROLLBACK` visade `SELECT COUNT(*)` att tabellen innehöll 0 rader.

Detta demonstrerade atomicity: ändringarna inom transaktionen rullades tillbaka tillsammans.

## Slutsats
Batching kan förbättra throughput genom färre SQL-anrop och commits.

Multi-row INSERTs var snabbast i experimentet eftersom de minskade både antalet statements och commits. Transaktionsbatcher minskade antalet commits men behöll 1 000 individuella statements.

CPU-bound arbete begränsas främst av beräkning, medan I/O-bound arbete begränsas av väntan på exempelvis disk eller nätverk. Benchmarken isolerade inte exakt vilken resurs som var flaskhalsen.

## Key takeaway
Att gruppera databasoperationer kan minska overhead kraftigt. WAL gör commits beständiga utan att alla ändrade databassidor måste skrivas direkt till SSD:n.
