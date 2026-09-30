# CPU and scheduling

## Syfte

Målet med laborationen var att förstå hur goroutines, OS-trådar och CPU-kärnor hänger ihop och varför fler goroutines inte alltid gör ett CPU-bound program snabbare.

## Experiment

Jag skapade ett CPU-bound Go-program med totalt 5 miljarder iterationer och delade samma mängd arbete mellan flera goroutines.

Min dator har:

- 8 fysiska CPU-kärnor
- 16 logiska CPU
- 2 hardware threads per kärna

Jag testade även olika värden för GOMAXPROCS.

## Resultat

| GOMAXPROCS | Real   | User   | CPU   |
|------------|--------|--------|-------|
| 1          | 1.23 s | 1.22 s | 99%   |
| 2          | 0.62 s | 1.23 s | 198%  |
| 4          | 0.31 s | 1.23 s | 388%  |
| 8          | 0.16 s | 1.24 s | 772%  |
| 16         | 0.17 s | 2.54 s | 1479% |

Med GOMAXPROCESS=8 visade även top och pidstat ungefär 800% CPU-anvädning.

## Vad jag lärde mig

En go routine är inte samma sak som en OS-tråd. Go-runtime schemalägger gorutines på OS-trådar och Linux schemalägger OS-trådarna på logiska CPU.

GOMAXPROCS begränsar hur många OS-trådar som samtidigt kan exekvera GO-kod.

Programmet skalade nästan linjärt från 1 till 8. Att gå från 8 till 16 gav däremot ingen förbättring för denna workload. Datorn har 8fysiska kärnor och de 16 logiska CPU delar resurser genom SMT.

Fler goroutines skapar därför inte mer fysisk CPU-kapacitet

`user` mäter sammanlagd CPu-tid. Flera CPU kan arbeta parallellt, vilket gör att `user` kan vara mycket större än `real`.

## Key takeaway

Concurrency gör det möjligt att strukturera flera samtidiga arbeten, men faktisk CPU-parallellism begränsas av hårdvaran. För CPU-bound arbete måste graden av parallelllism mätas istället för att anta att fler goroutines ger bättre prestanda.