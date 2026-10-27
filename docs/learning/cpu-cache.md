# CPU-cache och memory locality – vecka 6

## Syfte

Undersöka varför samma data kan bearbetas olika snabbt beroende på i vilken ordning programmet läser minnet. Koppla resultaten till CPU-cache, minneshierarki och *spatial locality*.

## Genomförande

Jag skrev två Go-funktioner som summerar värdena i en `[]int64`:

- `sequentialSum` läser elementen i ordning.
- `randomSum` läser samma element via en separat slice med slumpmässigt blandade index.

Jag körde benchmarks på Fedora med en AMD Ryzen 7 5800 (8 kärnor) och varierade datamängden från 1 000 till 10 miljoner element.

```bash
go test ./labs/cache -run '^$' -bench=. -benchmem -count=3
```

Resultaten nedan är medelvärden från tre körningar per storlek.

## Resultat

| Antal element | Sekventiellt (ns/op) | Slumpmässigt (ns/op) | Slumpmässigt / sekventiellt |
|---|---:|---:|---:|
| 1 000 | 488 | 266 | 0,54× |
| 10 000 | 4 932 | 4 599 | 0,93× |
| 100 000 | 48 208 | 78 471 | 1,63× |
| 1 000 000 | 485 825 | 912 046 | 1,88× |
| 10 000 000 | 5 108 143 | 50 557 455 | 9,90× |

Alla benchmarks rapporterade `0 B/op` och `0 allocs/op` under den uppmätta operationen.

Vid 10 miljoner element tog den slumpmässiga metoden nästan **10 gånger så lång tid** som den sekventiella.

## Förklaring

Vid sekventiell läsning kommer programmet åt närliggande minnesadresser. CPU:n kan då utnyttja *spatial locality*, cache lines och förhandsinläsning. Slumpmässiga hopp över stora datamängder kan ge fler cache misses och kräva åtkomst till långsammare minnesnivåer.

Vid små datamängder var skillnaden liten eller omvänd. Resultaten visar alltså inte att sekventiell läsning alltid är snabbast, utan att åtkomstmönstret får stor betydelse för större mängder data.

## Begränsningar

- `randomSum` läser även en extra index-slice, vilket kostar arbete och minnesåtkomst. Vi har därför inte isolerat cacheeffekten helt.
- Cache misses har inte mätts direkt med exempelvis `perf`.
- Ett separat korrekthetstest som bekräftar att båda funktionerna ger samma summa återstår.

## Key takeaway

**Hur data läses spelar roll för prestandan, inte bara hur många operationer programmet utför.** Det är relevant för Go-algoritmer, datastrukturer och senare arbete med databaser och index.
