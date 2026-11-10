[English](how-it-works.md) · **Español**

# Cómo funciona

SourceSight lee la demo con [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) y recorre la partida tick a tick. Para cada jugador vivo calcula posición, ángulos de vista, botones apretados y velocidad, y cada módulo acumula evidencia. Al final se aplican los umbrales y se arma el reporte. Todos los umbrales están en `pkg/core/constants.go`.

Una demo de 35 minutos a 128 tick (unos 270.000 ticks) tarda alrededor de 4 segundos.

## Autobunny (`pkg/modules/bhop.go`)

Un salto es *perfecto* si sale del piso dentro del primer tick de haber aterrizado: ahí `Jump()` se ejecuta antes que la fricción y no se pierde velocidad. Una persona con la rueda del mouse acierta entre 20% y 50%. Se marca con 75% o más (mínimo 10 saltos) o con 5 perfectos seguidos. Con 55% o más queda como sospechoso.

## Triggerbot (`pkg/modules/trigger.go`)

Se traza un rayo desde los ojos del jugador hacia donde mira y se intersecta con la caja de colisión de cada enemigo. Cuando el rayo entra en un enemigo empieza a contar y, al disparar, se registra el tiempo de reacción. Una persona promedia 150–250 ms con mucha dispersión. Se marca con al menos 4 muestras, promedio menor a 80 ms y desvío menor a 15 ms.

## Aimlock y spinbot (`pkg/modules/aim.go`)

Se mide cuánto giran los ángulos de vista entre un tick y el siguiente, corrigiendo el salto de 359° a 0°. Un giro de más de 45° en un tick es un *snap*. Un flick humano puede producir alguno suelto, por eso se marca solo cuando hay 5 o más dentro de un mismo segundo (y 10 en total). Así es como se ve un spinbot o un anti-aim: el yaw salta a un valor distinto en cada tick, normalmente con el pitch fijo en ±89°.

Se ignoran los primeros 1,5 s después de aparecer y los cambios a pitch exactamente 0, que son ángulos que impone el servidor al spawnear o teletransportar.

## Silent aim (`pkg/modules/aim.go`)

Cuando un disparo hace daño se compara hacia dónde miraba el tirador con la dirección real hacia la víctima. Disparando en movimiento la dispersión llega hasta unos 35°; más que eso, 3 veces, se marca.

## Wallhack (`pkg/modules/wallhack.go`)

La demo guarda qué enemigos tiene "vistos" cada jugador (el mismo dato que alimenta el radar). Se busca que la mira siga a un enemigo que **no** está visto, a menos de 4° de su pecho, durante al menos un segundo, recorriendo 20° o más. El recorrido es lo importante: dejar la mira quieta en un ángulo por donde va a salir alguien es buena colocación, no wallhack. Se marca con 3 seguimientos.

## Prefire (`pkg/modules/peek.go`)

Para cada pelea se mide cuánto tarda el jugador en hacer daño desde que el rival queda a la vista para él, usando el mismo dato de "visto" que el radar. Una persona necesita unos 200–400 ms para reaccionar; dañar en menos de 100 ms una y otra vez sugiere que ya sabía dónde iba a aparecer (wallhack) o un triggerbot. Se marca como **sospechoso** con 35% o más de peleas por debajo de 100 ms, mediana menor a 150 ms y al menos 15 peleas.

Se calibró con 13 demos reales: solo marcó a un jugador reportado como cheater. En la ficha del jugador se compara su prefire promedio de todas las demos con el de todos los demás, porque el patrón se ve mejor a lo largo de varias partidas.

## Macro (`pkg/modules/macro.go`)

Para pistolas semiautomáticas se mide el intervalo entre disparos. Un dedo tiene variación natural; un macro dispara con un ritmo casi constante. Se marca con un desvío menor a 1,5 ms entre disparos, medido sobre al menos 10.

## Kills cegado y por humo (`pkg/modules/flash_smoke.go`)

- **Cegado:** kills con el tirador totalmente cegado por una flash. Se marca con 4.
- **Humo:** kills cuyo trayecto atraviesa una granada de humo activa (esfera de 144 unidades) con pocas balas disparadas. Se marca con 5.

## Admins y chat (`pkg/modules/staff.go`, `pkg/modules/chat.go`)

Se reconocen comandos de castigo escritos en el chat (`!slay`, `/kick`, `sm_ban`, …) y los avisos `[SM] Admin: Slayed X.` del servidor. Si el admin castiga a quien lo mató en los 5 segundos anteriores, se marca como venganza.

En el chat se buscan insultos graves y binds repetidos.

## Qué no se detecta (todavía)

- **Ghosting por Discord:** que un compañero muerto cante posiciones por voz es legal, así que disparar donde estaba el asesino no prueba nada. El código está en `CheckGhosting` pero no se usa.
- Visibilidad real con la geometría del mapa: el wallhack usa el dato de "visto" del juego, que es una aproximación.

## Cómo ir a un incidente

Cada incidente trae `demo_gototick <tick>`, con unos segundos de margen antes de la jugada. Abrí la demo con `playdemo` y pegá el comando en la consola, o usá **Ver** en la app para que abra el juego directo en ese tick.
