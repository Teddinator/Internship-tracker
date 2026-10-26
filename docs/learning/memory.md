# Linux Memory – Week 5

## Syfte

Syftet med labben var att förstå hur Linux hanterar virtuellt och fysiskt minne, samt hur man undersöker minnesanvändning i ett Go-program.

## Genomförande

Jag skapade ett Go-program som allokerade 200 MiB med `make([]byte, size)`.

Jag mätte processens virtuella minne (VSZ) och residenta minne (RSS) före allokeringen, efter allokeringen och efter att programmet skrivit till varje minnessida.

Jag använde `ps`, `/proc/<pid>/status`, `pmap` och `/proc/<pid>/maps` för att undersöka processen.

## Resultat

| Mätning | VSZ | RSS |
|---|---:|---:|
| Före allokering | ~1198 MiB | ~2,3 MiB |
| Efter allokering | ~1456 MiB | ~3,1 MiB |
| Efter skrivning | ~1456 MiB | ~203,1 MiB |

Efter allokeringen ökade VSZ kraftigt medan RSS knappt förändrades. När programmet skrev till minnessidorna ökade RSS med ungefär 200 MiB.

### Page faults

Programmet skrev en byte var 4096:e byte, motsvarande en skrivning per 4 KiB minnessida.

**Förväntade minor page faults:** 51 200  
**Observerad ökning:** 51 202  
**Major page faults:** 0

Detta stämmer med att Linux tilldelade fysiska minnessidor vid första skrivningen, genom demand paging.

## Heap profiling

Jag använde Go:s `runtime/pprof` för att samla in en heap-profil.

```bash
go tool pprof -top heap.prof
```

Resultatet visade:

```text
Type: inuse_space
Total: 201.50 MB

flat: 200 MB
flat%: 99.25%
function: main.main
```

Nästan all profilerad heap-användning kom från `main.main`, där jag allokerade en slice på 200 MiB.

## Viktigaste lärdomar

- Virtuellt minne är inte samma sak som fysiskt RAM.
- VSZ visar processens virtuella minnesstorlek medan RSS visar hur mycket minne som är resident i RAM.
- Linux kan tilldela fysiska minnessidor först när de används, genom demand paging.
- En minor page fault är inte nödvändigtvis ett fel, utan kan vara en normal del av minneshanteringen.
- Go:s garbage collector kan återvinna heap-minne utan att minnet omedelbart återlämnas till Linux.
- `pprof` kan identifiera vilka funktioner som ansvarar för heap-allokeringar.

## Praktisk användning

Vid felsökning av hög minnesanvändning i ett Go-API kan jag använda Linux-verktyg för att mäta processens RAM-användning och `pprof` för att undersöka vilka delar av Go-koden som använder heap-minne.

Det hjälper mig att identifiera möjliga minnesproblem utifrån mätningar istället för att gissa.