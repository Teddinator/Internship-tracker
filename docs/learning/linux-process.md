# Linux-processen för Internship Tracker
## Syfte
Målet med laborationen var att köra Internship Tracker direkt på Fedora och undersöka hur Linux ser API-processen.
PostgreSQL kördes fortfarande i Docker, men själva Go-servern kördes som vanlig process på datorn.
## Start av API
Jag startade först PostgreSQL:
`docker compose up -d db`
Sedan laddade jag upp variablerna från .env.local i mitt shell:
```bash
    set -a
    source .env.local
    set +a
```
Till sist byggde och startade jag API:t:
```bash
go build -o bin/internship-tracker ./cmd/api
./bin/internship-tracker
```
API:t började lyssna på port `:8080`.
## PID, PPID, och användare
Efter den senaste starten hade API-processen PID `10467`. PID identifierar den aktuella processen och ändras när programmet startas om.
PPID var `8828`. PPID identifierar föräldrarprocessen, vilket i detta fall var shellen som startade API:t.

Processen kördes som användaren `theodor` med UID och GID `1000`. Det innebär att API får samma vanliga filrättigheter som min användare
och inte automatiskt har root behörigheter.

När servern vänta på requests hade processen tillståndet `S`, vilket betyder sleeping. Det är normalt för en server : den väntar på nätverksaktivitet i stället för att använda CPU hela tiden.
## Information i `/proc`
Linux exponerar information om varje process under `/proc/<PID>`.
Jag undersöäkte bland annat:
* `status` för PID, PPID, UID, minne, trådar och processtillstånd.
* `exe` för den körbara binären
* `cwd` för processens arbetskatalog
* `cmdline` för startkommandot
* `environ` för processens miljövariabler
* `fd/` för öppna file descriptors
* `maps` för processens virtuella minnesområden
`/proc/10467/exe` pekade på:
/home/theodor/Go Project/src/Internship-tracker/bin/internship-tracker
Processens arbetskatalog var repositoryts rot, och startkommandot var:
./bin/internship-tracker
## File descriptors
En file descriptor är ett processlokalt heltal som refererar till en öppen resurs som hanteras av Linuxkärnan. Resursen kan exempelvis var en fil, terminal eller nätverkssocket.
API-processen hade nio öppna descriptors
|  FD | Resurs       | Betydelse                                  |
| --: | ------------ | ------------------------------------------ |
| `0` | `/dev/pts/1` | Standard input från terminalen             |
| `1` | `/dev/pts/1` | Standard output till terminalen            |
| `2` | `/dev/pts/1` | Standard error till terminalen             |
| `4` | TCP-socket   | Anslutning till PostgreSQL                 |
| `5` | `eventpoll`  | Go-runtimens mekanism för att vänta på I/O |
| `6` | `eventfd`    | Intern signalering                         |
| `7` | TCP-socket   | HTTP-serverns lyssnande socket             | 
FD-numret är inte själva resursen. Det är ett handtag som processen använder för att hänvisa till resursen i kärnan.
## HTTP- och PostgreSQL-socketar
med `lsof -nP -p 10467` såg jag: 
TCP [::1]:42210->[::1]:5433 (ESTABLISHED) TCP *:8080 (LISTEN)

`Listen` visar att API väntade på inkommande TCP-anslutningare på port `8080`.

`ESTABLISHED` visar att API hade en upprättad TCP-anslutning till PostgreSQL. API använde den tillfälliga lokala porten `42210`
och anslöt sig till PostgreSQL via värdport `5433`.

När processen startades om ändrades PID, den tillfälliga klientporten och socketarnas interna ID. Resursernas funktion var däremot densamma.
## Miljövariabler
API kräver miljövariabeln DATABASE_URL. När den inte var satt avslutades programmet med:

*DATABASE_URL is not set*

När jag använde `set -a` och `source .env.local` blev variablerna exporterade från skalet.
API-processen ärvde sedan en kopia av dessa variabler när den startades.

Jag verifierade därefter att `DATABASE_URL` fanns i `/proc/<PID>/environ`.
En redan startad process får inte automatiskt nya värden  om miljön senare ändras i skalet. Processen måste startas om för att ärva den nya värdena.
## Virtuellt minne
`VmSize` beskriver processensvirtuella adressutrymme, medan `VmRSS` ungefär visar hur mycket av processens minne som för tillfället finns i fysiskt RAM.

Processen använder virtuella adresser. MMU och sidtabeller används för att översätta dessa till fysiska minnessidor när en giltig mapping finns.

I `/proc/<PID>/maps` såg jag bland annat:
* körbar programkod med rättigheterna r-xp
* skrivskyddad data med `r--p`
* skrivbar data med `rw-p`
* heap och stack
* reserverade områden utan åtkomst, markerade `---p`
* bibliotek som `libc.so.6`
* kernel-assisterade områden som `[vdso]` 
Ett stort virtuellt adressutrymme betyder därför inte att en process använder så mycket fysisk RAM.
## Behörighets experiment
Jag kunde läsa API-processens miljö eftersom mitt shell och API-processen kördes som samma användare.

När jag försökte läsa `/proc/1/environ` som användaren `theodor` fick jag:

<Permission denied> och exit kod <1>

Det visar att informationen i `/proc` fortfarande skyddas av linux behörighetsmodell.
Att processinformation presenteras som filer innebär inte alla användare får läsa den.
## Det som överraskade mig
Det som överraskade mig mest var att file descriptors inte bara används för vanliga filer.
Linux använder samma grundmodell även för terminaler, socketar och kernelresurser.

Jag såg också att processen hade cirka 1,6 GiB virtuellt adressutrymme, men bara cirka 11 MiB resident i fysiskt RAM.
Storleken på det virtuella adressutrymmet behöver inte motsvara datorns totala RAM.

## Sammanfattning
En linux-process har ett PID, en föräldraprocess(PPID) och en användaridentitet(UID). Den ärver exporterade
miljövariabler när den startas, använder file descriptors som handtag till öpnna resurser och arbetar i ett eget virtuellt adressutrymme.

Genom `/proc` och `lsof` kunde jag se att Internship Tracker körde rätt binär, väntade på HTTP-trafik på port `8080` och
hade en etablerad anslutning till PostgreSQL på port `5433`.
