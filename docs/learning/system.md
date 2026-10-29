
# Linux systemd — Week 9

## Syfte
Förstå hur systemd startar, stoppar och övervakar Linux-tjänster, samt hur användare, behörigheter och loggar påverkar en backendapplikation.

## systemd och service units
Systemd är Fedoras init-system (PID 1) och ansvarar bland annat för att hantera tjänster.

En unit beskriver en resurs eller uppgift som systemd kan hantera. En service unit beskriver hur en tjänst ska startas och övervakas.

Skapade `internship-tracker.service` för att köra Go-API:t direkt på Fedora medan PostgreSQL fortsatte köras i Docker.

Viktiga inställningar:
- `ExecStart`: sökväg till Go-binären.
- `User`: Linux-användaren som kör processen.
- `EnvironmentFile`: tjänstens miljövariabler.
- `Restart=on-failure`: automatisk omstart vid fel.
- `RestartSec=3`: tre sekunders väntetid före omstart.
- `WantedBy=multi-user.target`: kopplar tjänsten till normal systemuppstart när den aktiveras.

## Felsökning och SELinux
Använde `systemctl status`, `journalctl`, `namei` och `ausearch` för att felsöka tjänsten.

Problem som identifierades:
1. SELinux blockerade läsning av `.env.local` i `/home`.
2. SELinux blockerade exekvering av Go-binären med typen `user_home_t`.
3. `EnvironmentFile` expanderade inte `${POSTGRES_USER}` och `${POSTGRES_DB}` i `DATABASE_URL`.

Lösningar:
- Flyttade tjänstens miljökonfiguration till `/etc/sysconfig/internship-tracker`.
- Installerade Go-binären i `/usr/local/bin/internship-tracker`.
- Använde en färdig `DATABASE_URL` utan Bash-variabelexpansion.

Behöll SELinux i `Enforcing`.

## Processhantering och restart
Testade två sätt att stoppa API:t.

### Kontrollerat stopp
Körde `systemctl stop`.

Systemd skickade SIGTERM och Go-servern genomförde graceful shutdown:
- Väntade på aktiva HTTP-requests.
- Stoppade HTTP-servern.
- Stängde databasanslutningen.

Tjänsten startades inte om eftersom stoppet var avsiktligt.

### Simulerad krasch
Skickade SIGKILL till huvudprocessen.

Observationer:
- Ursprunglig PID: 201193.
- Processen avslutades med `status=9/KILL`.
- Systemd registrerade `Result: signal`.
- Efter cirka tre sekunder startades en ny process med PID 202062.

Detta verifierade `Restart=on-failure`.

## Användare och cgroups
Undersökte tjänsten med `systemctl show` och `ps`.

Observationer:
- API-processens PPID var 1 (systemd).
- Tjänsten tillhörde `/system.slice/internship-tracker.service`.
- Cgroups används för att gruppera, övervaka och begränsa processers resurser.

Bytte från användaren `theodor` till en dedikerad systemanvändare, `internship-api`.

Verifierade:
- `User=internship-api`.
- API:t svarade med HTTP 200 på `/health`.
- Tjänsteanvändaren fick `Permission denied` vid åtkomst till `/home/theodor`.

Aktiverade även autostart med `systemctl enable`. Resultatet blev `enabled`.

## Key takeaway
Systemd hanterar tjänsters livscykel, kan upptäcka processkrascher och automatiskt starta om tjänster.

`systemctl` används för att styra tjänster och `journalctl` för felsökning. SELinux och vanliga filbehörigheter är separata säkerhetskontroller.

En dedikerad icke-root-användare begränsar tjänstens rättigheter. Cgroups hjälper systemd att hantera processer som tillhör samma tjänst.
