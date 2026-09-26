# Processer och signaler

## Signaler till API:t

SIGTERM och SIGINT fångades av API:t och körde samma shutdown-kod. Servern väntade på aktiva requests och stängde sedan databaskopplingen.

SIGKILL avslutade processen direkt. Shellen skrev `Killed`, men API:t hann inte skriva några shutdown-loggar.

## Pågående request

Jag testade en tillfällig `/lab/slow`-endpoint och skickade SIGTERM medan den kördes:

- 5 sekunders arbete: klienten fick `200 OK` och `done`. Servern väntade tills requesten var klar.
- 12 sekunders arbete: efter 10 sekunders shutdown-väntan loggades `context deadline exceeded`. Klienten fick `Empty reply from server`.

Det var `shutdownCtx` på 10 sekunder som begränsade väntan. Loggen `HTTP server stopped` skrevs även när shutdown misslyckades.

## Fork, exec och zombies

`fork()` skapar en barnprocess med ett nytt PID. `exec()` byter programmet som processen kör men behåller dess PID. I labben hade `sleep` mitt shell som förälder och körde `/usr/bin/sleep`.

Ett avslutat barn visades som `Z` (zombie) tills föräldern hämtade dess avslutsstatus med `waitpid()`. Därefter försvann processposten.