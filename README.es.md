<p align="center">
  <img src="cmd/sourcesight/winres/icon_128.png" width="88" alt="Ícono de SourceSight">
</p>

<h1 align="center">SourceSight</h1>

<p align="center">
  <a href="README.md"><img src="https://img.shields.io/badge/lang-English-2a2c30?style=for-the-badge" alt="English"></a>
  <a href="README.es.md"><img src="https://img.shields.io/badge/idioma-Español-e07b2a?style=for-the-badge" alt="Español"></a>
</p>

<p align="center">
  <a href="https://github.com/Hdauu/SourceSight/releases/latest"><img src="https://img.shields.io/github/v/release/Hdauu/SourceSight?label=descargar&color=e07b2a" alt="Descargar"></a>
  <img src="https://img.shields.io/badge/plataforma-Windows-2a2c30" alt="Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/licencia-MIT%20%2B%20Commons%20Clause-2a2c30" alt="Licencia"></a>
</p>

Una herramienta para el staff de servidores de CS:GO (Legacy). Abrís una demo (`.dem`) y marca a los jugadores cuyo comportamiento no parece humano: autobunny, triggerbot, spinbot, wallhack, macros y más. Cada incidente trae su tick exacto para ir a verlo en el juego con un click.

Funciona en Windows, todo en tu PC. No se sube nada a ningún lado.

> **Los resultados son pistas, no pruebas.** Cada detección explica qué midió y con qué números, y trae el comando para ir a verla. Mirá la demo antes de sancionar a alguien.

## Contenido

- [Instalar](#instalar)
- [Uso rápido](#uso-rápido)
- [Cómo se analiza una demo](#cómo-se-analiza-una-demo)
- [Cómo está organizada la app](#cómo-está-organizada-la-app)
- [El reporte de una demo](#el-reporte-de-una-demo)
- [Niveles](#niveles)
- [Detecciones](#detecciones)
  - [Autobunny](#autobunny) · [Triggerbot](#triggerbot) · [Aimlock y spinbot](#aimlock-y-spinbot) · [Silent aim](#silent-aim) · [Wallhack](#wallhack) · [Prefire](#prefire) · [Macro](#macro) · [Kills cegado](#kills-cegado) · [Kills a través del humo](#kills-a-través-del-humo) · [AFK](#afk) · [Teamflash](#teamflash)
- [Dónde pega: la figura del cuerpo](#dónde-pega-la-figura-del-cuerpo)
- [Ping y lag](#ping-y-lag)
- [Observaciones automáticas](#observaciones-automáticas)
- [Registro de SHIFT](#registro-de-shift)
- [Mapa](#mapa)
- [Admins y chat](#admins-y-chat)
- [Historial y demos repetidas](#historial-y-demos-repetidas)
- [Jugadores: promedios entre demos](#jugadores-promedios-entre-demos)
- [Confirmar un cheater](#confirmar-un-cheater)
- [Ver en juego](#ver-en-juego)
- [Archivo automático de demos](#archivo-automático-de-demos)
- [Exportar evidencia](#exportar-evidencia)
- [Bot de Discord](#bot-de-discord)
- [Actualizaciones](#actualizaciones)
- [Dónde se guardan los datos](#dónde-se-guardan-los-datos)
- [Qué no se detecta](#qué-no-se-detecta)
- [Compilar y probar](#compilar-y-probar)
- [Publicar una versión](#publicar-una-versión)
- [Licencia](#licencia)

## Instalar

1. Bajá **`SourceSight-Setup-vX.Y.Z.exe`** desde la [última versión](https://github.com/Hdauu/SourceSight/releases/latest).
2. Ejecutalo. Se instala solo para tu usuario, sin pedir permisos de administrador.
3. Puede aparecer *"Windows protegió su PC"* porque el instalador no está firmado. Hacé click en **Más información → Ejecutar de todas formas**.

¿Solo necesitás el bot de Discord? En la misma versión está `SourceSightBot.exe` suelto, sin instalar nada. Ver [Bot de Discord](#bot-de-discord).

Necesita el runtime de WebView2, que ya viene con Windows 10 y 11.

## Uso rápido

1. Soltá una o varias demos (o una carpeta entera) sobre la ventana, o usá **Abrir demos**. El instalador también agrega *Abrir con → SourceSight* para los `.dem`.
2. Con una demo se abre su reporte. Con varias, ves el progreso de cada una y al final un resumen; click en cualquiera para abrirla.
3. Click en un jugador para ver sus métricas, dónde pega, sus incidentes por ronda y el botón **Ver** de cada uno.
4. En la pestaña **Jugadores** ves a cada jugador sumado en todas tus demos y comparado con el resto.

Para usar **Ver**, configurá una vez la carpeta de CS:GO en **Ajustes** (abajo a la izquierda, en el menú).

## Cómo se analiza una demo

SourceSight lee la demo con [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) v4 (la última que soporta CS:GO; la v5 es solo para CS2) y recorre la partida tick a tick. En cada tick, para cada jugador vivo, toma su posición, hacia dónde mira, qué botones aprieta, su velocidad y qué enemigos tiene a la vista. Cada detector acumula evidencia mientras avanza la partida. Al final se aplican los umbrales y se arma el reporte.

Una demo de 35 minutos a 128 tick (unos 270.000 ticks) tarda entre 4 y 8 segundos.

**Varias demos a la vez.** Si abrís varias, se analizan en paralelo:

- Usa un análisis por cada hilo del procesador menos uno, que queda libre para que la ventana no se trabe.
- Cada análisis usa hasta unos 300 MB de memoria (medido: una demo de 207 MB ocupó unos 210 MB). Por eso nunca usa más de la mitad de la RAM libre: con poca memoria analiza menos a la vez.
- Lo pesado (leer y analizar) va en paralelo. Guardar en el historial, preguntar por repetidas y actualizar la base de reincidentes va **de a una**, en el orden en que terminan. Así, si soltás dos copias de la misma partida juntas, la segunda igual se detecta como repetida.
- La pantalla muestra cada demo con su estado (en cola, verificando, analizando, guardando, lista, ya estaba, repetida, error, cancelada), su barra, el resultado y cuánto tardó. Arriba hay una barra total con el tiempo que falta. **Cancelar** frena de verdad los análisis en curso.
- Al terminar queda un resumen ("2 con cheat · 1 sospechosa · 5 limpias…").

Medido con 12 demos reales en un procesador de 12 hilos: 62 s una por una, 11 s en paralelo. Los reportes salen idénticos de las dos formas, y hay un test que lo comprueba.

## Cómo está organizada la app

A la izquierda, un menú con las secciones:

- **Inicio:** lo que conviene revisar primero. Las últimas demos con cheat o sospechosos, los totales del historial, el estado del archivo automático y el aviso si hay demos para re-analizar.
- **Demos:** el historial a la izquierda (buscar, filtrar por nivel o categoría) y el reporte de la demo elegida a la derecha.
- **Jugadores:** la lista de jugadores de todas tus demos y la ficha de cada uno. Sin jugador elegido muestra "Todas las demos", con los totales y cómo pegan en promedio.
- **Archivo:** el archivo automático de demos del servidor: estado, configuración y la lista de demos que guardó.
- **Ajustes:** carpeta del juego, segundos antes de la jugada, teclas de incidente, distancia del registro de SHIFT, umbrales y versión.

Arriba, cuando hay una demo abierta: su mapa, archivo y servidor, **Re-analizar** y **Exportar** (imagen PNG o resumen para Discord). **Abrir demos** está siempre. Las demos se pueden soltar en cualquier parte de la ventana.

## El reporte de una demo

Arriba, un resumen (jugadores, cheats, duración, reincidentes). Abajo, pestañas:

- **Jugadores:** una fila por jugador con su resultado, perfs, reacción, prefire, % a la cabeza, snaps, casos de SHIFT y ping. Se ordena por cualquier columna (por defecto, del peor resultado al mejor) y se filtra por nivel o por nombre. Click en un jugador para ver su detalle. Los nombres hechos de caracteres invisibles se muestran como "(nombre invisible)".
- **Incidentes:** cada incidente con su ronda, tipo, jugador, comando y botón **Ver**.
- **Mapa:** ver [Mapa](#mapa).
- **SHIFT:** ver [Registro de SHIFT](#registro-de-shift).
- **Chat y admins:** los comandos de admins y el chat.

La app recuerda la última pestaña que usaste.

## Niveles

Cada jugador queda con un nivel. El de la demo es el peor de sus jugadores.

| Nivel | Qué significa |
|---|---|
| **Cheat detectado** | Superó un umbral fuerte: autobunny, triggerbot, aimlock/spinbot, silent aim, wallhack o macro |
| **Sospechoso** | Indicios más débiles que conviene mirar: bhop alto o rachas de perfs, algunos giros bruscos, prefire, kills cegado o a través de humo |
| **Conducta** | No es cheat pero le importa al staff: AFK, flashes al propio equipo |
| **Limpio** | No se encontró nada |
| **Confirmado** | Lo marca un admin después de ver la demo, nunca la app. Queda guardado por SteamID (ver [Confirmar un cheater](#confirmar-un-cheater)) |

En el historial se puede filtrar por nivel y por categoría de detección.

## Detecciones

Todos los umbrales están en `pkg/core/constants.go` y las reglas que deciden el nivel en `pkg/core/state.go` y `pkg/core/level.go`. Los tiempos que se definen en ticks están pensados para 64 tick y se escalan al tickrate real de la demo, así que la misma regla vale en demos de 64 y de 128.

Cada detección explica: qué busca, qué datos usa, cómo calcula, cuándo marca y por qué ese número, qué la puede engañar y cómo verificarla.

### Autobunny

`pkg/modules/bhop.go` · nivel **cheat** (AUTOBUNNY) o **sospechoso** (SOSPECHOSO-BHOP)

**Qué busca.** Scripts que saltan solos en el instante exacto en que el jugador toca el piso, para encadenar saltos sin perder velocidad.

**Datos.** Si el jugador está en el piso (flag `FL_ONGROUND`), si aprieta saltar (`IN_JUMP`), escaleras, agua y velocidad, en cada tick.

**Cálculo.**
- Cuando aterriza, se anota el tick. Cuando despega apretando saltar, se cuentan los ticks que estuvo en el piso.
- Si despegó con 0 o 1 tick en el piso, es un salto **perfecto**: ahí el salto se ejecuta antes que la fricción y no se pierde velocidad.
- Un despegue después de más de 15 ticks en el piso no cuenta: es caminar y saltar, no bunny hop. Escaleras y agua no cuentan.
- Se guarda el total de saltos, los perfectos, el porcentaje (*perfs*) y la racha más larga de perfectos seguidos.

**Cuándo marca.**
- **AUTOBUNNY (cheat):** 75 % de perfectos o más, sobre al menos 10 saltos.
- **SOSPECHOSO-BHOP:** con 30 saltos o más, 55 % de perfectos o una racha de 5 seguidos.

**Por qué.** Una persona con la rueda del mouse acierta entre 20 % y 50 % de perfectos. Pasar el 75 % sostenido no se logra a mano. Una racha de 5 seguidos es muy improbable, pero un buen bhopper la puede lograr alguna vez, por eso sola es solo sospechosa. Con menos de 30 saltos la muestra es chica para decir "sospechoso".

**Límites.** Un servidor con autobunny propio (plugin de bhop) hace que todos salten perfecto. En esos servidores la detección no sirve.

**Cómo verificar.** En la demo, con `host_timescale 0.25`, mirá cómo encadena los saltos. Con script, el jugador no pierde ni gana velocidad de forma irregular y nunca "falla" un salto. Cada racha de 5+ aparece como incidente con su tick.

### Triggerbot

`pkg/modules/trigger.go` · nivel **cheat** (TRIGGERBOT)

**Qué busca.** Programas que disparan solos en cuanto la mira pasa por encima de un enemigo.

**Datos.** Hacia dónde mira el jugador, la posición de sus ojos y la caja de colisión de cada enemigo vivo, en cada tick, y el momento en que aprieta disparar.

**Cálculo.**
- Se traza un rayo desde los ojos del jugador hacia donde mira y se intersecta con la caja de cada enemigo.
- Cuando el rayo entra en un enemigo empieza a contar. Al apretar disparar se registra el tiempo transcurrido (*reacción*), entre 0 y 1,5 s.
- Si la mira sale del enemigo antes del disparo, la medición se descarta.
- Las reacciones que caen durante un pico de lag se descartan (ver [Ping y lag](#ping-y-lag)).

**Cuándo marca.** Con al menos 4 reacciones medidas, promedio menor a **80 ms** y desvío menor a **15 ms**.

**Por qué.** Una persona tarda en promedio entre 150 y 250 ms, y con mucha variación de una vez a otra. Un triggerbot reacciona siempre rápido y siempre igual. Por eso se piden las dos cosas: rapidez y regularidad.

**Límites.** Mide contra la caja de colisión, no contra la forma exacta del cuerpo. Un jugador que ya está disparando cuando el rival cruza su mira (spray) no genera una reacción nueva.

**Cómo verificar.** Mirá los incidentes en cámara lenta: con triggerbot el disparo sale justo cuando la mira toca al rival, incluso cuando el rival apenas asoma.

### Aimlock y spinbot

`pkg/modules/aim.go` · nivel **cheat** (AIMLOCK) o **sospechoso** (SOSPECHOSO-AIM)

**Qué busca.** Mira que salta de golpe al objetivo (aimlock) o que gira sin parar (spinbot / anti-aim).

**Datos.** Los ángulos de vista (pitch y yaw) de cada tick.

**Cálculo.**
- Se mide cuánto giró la vista entre un tick y el siguiente, corrigiendo el salto de 359° a 0°.
- Un giro de más de **45° en un solo tick** es un *snap* (a 128 tick, un tick son 7,8 ms).
- Se ignoran los primeros 3 segundos después de aparecer y los cambios a un pitch exactamente 0: son ángulos que impone el servidor al spawnear o teletransportar.
- **No cuentan los giros medidos durante un corte de conexión.** Con lag, los comandos del jugador llegan amontonados y el giro de varios ticks aparece de golpe en uno. Se descarta el snap si en ese tick: el servidor procesó más de un comando del jugador (se ve en su *tickbase*, que avanza uno por comando), el jugador se movió a más de 1200 u/s (un warpeo; nadie corre tan rápido, ni con bhop), o tuvo un pico de ping. Esos giros se cuentan aparte ("Giros con lag") y, si son 5 o más, se avisa en las observaciones. Un ping alto pero estable no genera snaps y no se descuenta nada.

**Cuándo marca.** Solo si los snaps se agrupan: al menos 5 dentro de un mismo segundo y 10 o más en la partida.
- 50 snaps o más: **AIMLOCK (cheat)**.
- De 10 a 49: **SOSPECHOSO-AIM**.

**Por qué.** Un flick humano rápido puede producir algún snap suelto, pero nunca 5 en un segundo. Un spinbot produce cientos: el yaw salta a un valor distinto en cada tick, normalmente con el pitch fijo en ±89°. Entre 10 y 49 puede ser un cheat activado un momento o flicks raros, y conviene mirarlo.

**Límites.** Un aimlock "suave" (smooth aimbot), que gira rápido pero repartido en varios ticks, no supera los 45° por tick y no se detecta así.

**Cómo verificar.** Ir al tick del incidente y mirar en primera persona: el spinbot se ve como la vista girando sin control. Desde afuera se ve al modelo dando vueltas o mirando al piso.

### Silent aim

`pkg/modules/aim.go` · nivel **cheat** (SILENT AIM)

**Qué busca.** Cheats que mandan al servidor un ángulo de disparo distinto del que se ve en pantalla: la bala pega aunque la mira no apunte al rival.

**Datos.** Hacia dónde miraba el tirador al disparar, la posición de sus ojos y la posición de la víctima cuando recibe el daño.

**Cálculo.** Cuando un disparo hace daño (hasta 16 ticks después del tiro), se compara la dirección en la que miraba el tirador con la dirección real desde sus ojos hasta el pecho de la víctima.

**Cuándo marca.** Si la diferencia supera **35°**, 3 veces en la partida.

**Por qué.** Disparando en movimiento o saltando con SMGs y escopetas, la dispersión puede llegar a unos 35°. Más que eso no lo explica la imprecisión del arma: la bala fue en otra dirección que la vista.

**Límites.** Las balas que atraviesan paredes y los rebotes de daño en zonas raras pueden dar diferencias grandes de vez en cuando. Por eso se piden 3.

**Cómo verificar.** En el tick del incidente, en primera persona: el rival recibe el daño sin estar cerca de la mira.

### Wallhack

`pkg/modules/wallhack.go` · nivel **cheat** (WALLHACK)

**Qué busca.** Jugadores que siguen con la mira a un enemigo que está detrás de una pared.

**Datos.** La demo guarda qué enemigos tiene "vistos" cada jugador (`m_bSpottedByMask`, el mismo dato que alimenta el radar). Además, hacia dónde mira cada uno, las posiciones y las velocidades.

**Cálculo.**
- En cada tick se busca un enemigo **no visto** a menos de 4° de la mira (apuntando al pecho), entre 150 y 3000 unidades de distancia.
- Si ninguno de los dos se mueve, no cuenta: mira quieta sobre alguien quieto es colocación de mira.
- Si la mira sigue a ese mismo enemigo durante **1 segundo**, recorriendo **20° o más**, es un seguimiento.

**Cuándo marca.** Con 3 seguimientos en la partida.

**Por qué.** El recorrido es lo importante: dejar la mira quieta en un ángulo por donde va a salir alguien es buena colocación, no wallhack. Seguir 20° a alguien que no ves durante un segundo es otra cosa. Menos de 150 unidades se descarta porque de cerca se oye y se sigue fácil, y más de 3000 porque a esa distancia cualquier mira pasa cerca de alguien.

**Límites.** El dato de "visto" es una aproximación del juego, no la geometría exacta del mapa. Seguir por el sonido de los pasos a alguien detrás de una pared fina es legítimo y puede parecerse. Por eso hacen falta 3 casos.

**Cómo verificar.** Mirá cada seguimiento en primera persona con `spec_show_xray 1`, o con el radar: la mira acompaña al rival a través de la pared.

### Prefire

`pkg/modules/peek.go` · nivel **sospechoso** (PREFIRE)

**Qué busca.** Jugadores que dañan al rival apenas aparece, más rápido de lo que una persona puede reaccionar, una y otra vez. Sugiere que ya sabía dónde iba a aparecer (wallhack) o un triggerbot.

**Datos.** El mismo dato de "visto" que el radar, y el daño de cada bala.

**Cálculo.**
- Para cada pelea, se mide desde que el rival queda a la vista del jugador hasta la primera bala que lo daña (máximo 2 s; más tarde ya no es reacción a esa aparición).
- Una medición de menos de **100 ms** es un *prefire*.
- Las mediciones durante un pico de lag se descartan (ver [Ping y lag](#ping-y-lag)).

**Cuándo marca.** Con al menos 15 peleas medidas, **35 % o más** de prefires y una mediana menor a **150 ms**.

**Por qué.** Una persona necesita unos 200 a 400 ms para ver y reaccionar. Un prefire suelto pasa (pre-disparar un ángulo conocido es una técnica legal), pero no en un tercio de las peleas. Se calibró con 13 demos reales: solo marcó a un jugador reportado como cheater.

**Límites.** Pre-disparar ángulos comunes es legítimo y en algunos mapas hay muchos. Por eso es solo "sospechoso" y pide muchas peleas.

**Cómo verificar.** En la ficha del jugador se compara su % de prefire de todas las demos contra el de todos los demás: el patrón se ve mejor a lo largo de varias partidas. Mirá los incidentes: ¿dispara antes de que el rival termine de asomar, en lugares que no son ángulos típicos?

### Macro

`pkg/modules/macro.go` · nivel **cheat** (MACRO)

**Qué busca.** Programas o mouses con macro que hacen clic por el jugador con una pistola semiautomática para disparar tan rápido como permite el arma.

**Datos.** Cada vez que el jugador aprieta disparar (el flanco de subida de `IN_ATTACK`) y qué arma tiene en la mano.

**Cálculo.**
- Solo con pistolas que disparan una vez por clic: Glock, USP-S, P2000, P250, Desert Eagle, Dual Berettas, Five-SeveN, Tec-9 y R8. El arma se reconoce por su tipo. La CZ75 no entra porque es automática.
- Se mide el tiempo entre clics seguidos. Una pausa de más de 500 ms o cambiar a otra arma empieza una ráfaga nueva, para que la pausa entre peleas no se mezcle.
- Se mira el desvío de los últimos 10 intervalos de la ráfaga.

**Cuándo marca.** Si esos 10 intervalos tienen un desvío menor a **1,5 ms**. Hay un incidente por ráfaga, como máximo uno cada 5 s.

**Por qué.** Un dedo tiene variación natural: entre clic y clic cambia en varios milisegundos. Un macro hace clic con un ritmo casi constante. A 128 tick un tick son 7,8 ms, así que un desvío menor a 1,5 ms en 10 intervalos significa que prácticamente todos midieron exactamente lo mismo.

**Límites.** Disparar con la rueda del mouse (`bind mwheeldown +attack`) en ruedas de giro libre puede ser muy regular. Si aparece un MACRO, revisá si el jugador usa la rueda.

**Cómo verificar.** Mirá la ráfaga en cámara lenta y escuchá el ritmo de los disparos: con macro suena como un metrónomo.

### Kills cegado

`pkg/modules/flash_smoke.go` · nivel **sospechoso** (ANTI-FLASH)

**Qué busca.** Cheats que eliminan el efecto de las flashes.

**Datos.** En el momento de cada kill con arma de fuego, cuánto le queda de ceguera al asesino.

**Cálculo.** Una kill con más de 2,5 s de ceguera restante (pantalla totalmente blanca) cuenta como kill cegado.

**Cuándo marca.** Con 4 kills cegado en la partida.

**Por qué.** Totalmente cegado se puede matar de casualidad (spray hacia donde estaba el rival), pero no 4 veces.

**Límites.** Las flashes de costado o de espalda ciegan menos de lo que el juego reporta en algunos casos.

**Cómo verificar.** Ver la kill en primera persona: con anti-flash el jugador apunta con precisión con la pantalla blanca.

### Kills a través del humo

`pkg/modules/flash_smoke.go` · nivel **sospechoso** (SMOKE-VISION)

**Qué busca.** Cheats que dejan ver a través del humo.

**Datos.** Las granadas de humo activas (posición, inicio y fin), la posición de los ojos del asesino y de la víctima, y las balas que disparó el asesino en la ronda.

**Cálculo.**
- Se traza el segmento entre los ojos del asesino y la víctima. Solo cuenta un humo que esté **entre** los dos (una esfera de 144 unidades de radio): uno detrás de la víctima no tapa nada.
- El humo tiene que existir en ese momento: un humo dura 18 s desde que se abre y desaparece al terminar la ronda. El juego borra los humos al cambiar de ronda sin dejar registro en la demo, por eso se usa la duración y no el aviso de fin del humo.
- Si además el asesino disparó 3 balas o menos en la ronda, es una kill a través del humo con pocas balas: no fue spray a ciegas.

**Cuándo marca.** Con 5 kills así en la partida.

**Por qué.** Disparar a través del humo hacia donde se escuchan pasos es legítimo, pero suele requerir muchas balas. Matar 5 veces con 3 balas o menos indica que veía al rival.

**Límites.** El humo se modela como una esfera, pero el real se deforma con el mapa. Un rival que se ve por el borde del humo puede contar.

**Cómo verificar.** En la kill, en primera persona: ¿el rival era visible? ¿El jugador apuntó directo a él?

### AFK

`pkg/modules/staff.go` · nivel **conducta** (AFK/GRIEF)

**Qué busca.** Jugadores que no juegan.

**Cálculo.** Al final de cada ronda se mide cuánto se movió el jugador desde el final de la ronda anterior. Menos de 50 unidades es una ronda quieta.

**Cuándo marca.** Con 3 rondas quietas seguidas.

**Límites.** Un jugador que cuida siempre el mismo lugar y muere rápido puede moverse poco. Por eso es conducta y no cheat.

### Teamflash

`pkg/modules/staff.go` · nivel **conducta** (TEAMFLASH)

**Qué busca.** Jugadores que ciegan a su propio equipo.

**Cálculo.** Se suman los segundos de ceguera que cada flash del jugador causó a sus compañeros.

**Cuándo marca.** Con 35 segundos o más en la partida.

**Por qué.** Cegar a un compañero alguna vez pasa jugando. 35 segundos es sistemático.

## Dónde pega: la figura del cuerpo

`pkg/core/hits.go`, `pkg/gui/ui/js/16-body.js`

Cada bala que daña a un enemigo trae la zona del cuerpo donde pegó (*hit group*): cabeza, cuello, pecho, estómago, cada brazo y cada pierna. SourceSight guarda, por jugador:

- tiros disparados, impactos y daño causado en cada zona;
- lo mismo separado por tipo de arma: rifles, snipers (AWP, SSG 08, autos), pistolas, SMG, escopetas y ametralladoras;
- kills y kills con headshot.

Solo cuentan las balas a enemigos: cuchillo, granadas, fuego y zeus no cuentan.

**La figura.** Es una silueta lisa y minimalista, con cada zona como una pieza aparte. Cada zona se pinta según cuánto le pega el jugador, en una escala de calor de gris (nada) a naranja y amarillo claro (mucho), y al lado tiene su %. La escala es fija (el 45 % de los impactos ya es el color máximo) y tiene su barra de referencia debajo, así que dos figuras lado a lado se comparan a simple vista. Los brazos y piernas son los del soldado, como si lo tuvieras de frente: su brazo izquierdo queda a tu derecha.

**Modos.**
- **Jugador:** sus impactos en esa demo (detalle del jugador) o en todas sus demos (ficha del jugador).
- **Contra qué se compara, nunca contándolo a él:** en el detalle de una demo, **el resto de esa partida** (mismo mapa, mismos rivales); en la ficha del jugador, **el resto de los jugadores** del historial; en "Todas las demos", el promedio de todos.
- **Comparación:** las dos figuras lado a lado.

Se puede filtrar por tipo de arma. Debajo hay una tabla con impactos, daño, % y la diferencia contra el promedio en puntos. Una diferencia en la cabeza de 10 puntos o más sale en amarillo, solo como dato. Con menos de 10 impactos avisa que los porcentajes todavía no dicen mucho.

**Precisión.** Es impactos dividido tiros, **sin escopetas**: cada perdigón que pega cuenta como un impacto, y con escopeta la precisión pasaría el 100 % (en demos reales la Nova dio 141 %). Para la figura los perdigones sí cuentan, porque cada uno cae en una zona.

**Qué es normal.** En CS:GO, entre el 10 y el 25 % de los impactos de un jugador común van a la cabeza, y la mayoría al pecho y al estómago. Un jugador que pega mucho más a la cabeza que el resto **no es prueba de nada**: los jugadores buenos pegan más a la cabeza. Es un dato para mirar junto con lo demás.

## Ping y lag

`pkg/core/ping.go`

La demo trae el ping de cada jugador (el del scoreboard, que el servidor actualiza más o menos una vez por segundo). SourceSight lo mide una vez por segundo **con el jugador vivo**, para no contar los picos de cuando se conecta, carga el mapa o está de espectador.

**¿Un ping alto cambia las detecciones? No.** Las reacciones (prefire, triggerbot) se miden entre dos cosas que llegan al servidor por la conexión del mismo jugador: lo que vio y cuándo disparó. Un ping estable retrasa las dos por igual y se cancela. Un jugador con ping alto no "parece más rápido"; si reacciona tarde, es porque reaccionó tarde.

**Lo que sí engaña es un pico de lag.** Durante un pico, los comandos del jugador llegan amontonados y una reacción normal puede parecer imposiblemente rápida. Por eso, si el ping está **50 ms o más por encima del habitual** de ese jugador, las reacciones medidas en ese momento se descartan y se cuentan aparte ("Reacciones con lag: N descartadas"). El "habitual" es un promedio móvil que no sigue a los picos, para que un pico largo no pase a ser lo normal.

En 25 demos reales, los picos ocupan el 0,09 % del tiempo con los jugadores vivos, y no se descartó ninguna de 17.538 reacciones: el filtro está para el caso raro, no cambia los resultados normales.

El ping promedio y el pico de cada jugador se muestran en su detalle y en su ficha. El ping solo nunca marca a nadie.

## Observaciones automáticas

`pkg/core/notes.go`

Cuando un número de un jugador se sale de lo normal, la app lo escribe en una frase dentro de un recuadro **"Para mirar"**. Por ejemplo:

> El 46% de sus impactos fueron a la cabeza (51 de 111); el resto de la partida: 22%.

**Son solo informativas: nunca cambian el nivel ni las detecciones.** Sirven para saber qué mirar en la demo.

**Contra qué se compara.** El jugador nunca se compara consigo mismo:
- en una demo, contra **el resto de la partida**;
- en la ficha del jugador (todas sus demos), contra **el resto de los jugadores del historial**.

**Reglas.** Cada una pide datos suficientes del jugador y de la comparación (la comparación necesita 3 veces el mínimo del jugador), así que una buena ronda suelta no genera frases.

| Qué | Mínimo | Cuándo aparece |
|---|---|---|
| Impactos a la cabeza | 20 impactos | 30 % o más, y 15 puntos o más por encima de la comparación |
| Precisión (sin escopetas) | 80 tiros | 40 % o más, y 1,6 veces la de la comparación |
| Kills con headshot | 8 kills | 75 % o más, y 25 puntos o más por encima |
| Tiempo entre mira y disparo | 8 reacciones | 60 % o menos que el de la comparación |
| Daño apenas aparece el rival | 10 peleas | 20 % o más de prefires sin llegar al 35 % de la detección, y 2,5 veces la comparación |
| Ping alto | — | 120 ms de promedio o más |
| Reacciones descartadas por lag | — | Si se descartó alguna |

En 25 demos reales aparecen observaciones en 24 de 339 jugadores (7 %).

Se muestran en el detalle del jugador, en su ficha y en el mensaje del bot de Discord.

## Registro de SHIFT

`pkg/core/shift.go`, `pkg/gui/ui/js/17-shift.js`

**Qué es.** Una lista de cada momento en que un jugador caminó con shift (`+speed`, caminar sin hacer ruido) teniendo cerca un enemigo que no veía. Caminar en silencio es como uno se acerca sin ser escuchado, así que hacerlo justo al lado de un enemigo que "no sabe" que está ahí vale la pena mirarlo. Pero también es juego completamente normal, por eso el registro **no cambia el nivel ni las detecciones**: es una lista para revisar, sin otro criterio que la distancia y la visibilidad.

**Datos.** En cada tick: si el jugador tiene apretado shift (el botón `IN_SPEED`), la distancia a cada enemigo vivo y si lo está viendo (el mismo dato de "visto" que usa el radar).

**Cálculo.**
- Un caso empieza cuando el jugador camina con shift y hay al menos un enemigo **que no ve** a **1000 unidades o menos** (más o menos la distancia a la que se escuchan los pasos corriendo).
- Mientras ese enemigo (o cualquier otro que no ve) siga cerca, todo va al **mismo caso**: cada vez que vuelve a apretar shift se suma como una apretada más, no como un caso nuevo.
- El caso se cierra cuando pasa 1 segundo sin ningún enemigo cerca, cuando el jugador muere o cuando termina la ronda.
- No cuenta el calentamiento ni el congelamiento del inicio de cada ronda.
- Si la demo no trae el dato de visibilidad para un enemigo, se lo toma como visto: nunca agrega un caso por falta de datos.

**Qué muestra cada caso.** Ronda, jugador, los enemigos cerca que no veía, cuántas veces apretó shift, cuánto tiempo caminó con shift, lo más cerca que estuvo un enemigo, el comando `demo_gototick` y el botón **Ver**.

**Dónde está.**
- En el reporte de la demo, la sección **"Registro de SHIFT"**, con un filtro por jugador (ordenado por cantidad de casos). Muestra 25 y se pueden pedir más.
- En el detalle de un jugador, un recuadro con cuántos casos tuvo y el botón **Ver en el registro**, que abre la lista filtrada por él.
- El bot de Discord muestra un resumen: cuántos casos hubo y los jugadores con más.

**Distancia.** Se cambia en **Ajustes** (500, 750, 1000, 1500 o 2000 unidades) y en el bot con `shiftRadius` en `bot_config.json`. Se aplica a las demos que analices o re-analices después de cambiarla.

**Qué esperar.** En partidas normales hay muchos casos: en 25 demos reales salieron de decenas a unos 300 por partida, y hasta unos 45 por jugador. Lo que conviene mirar es el jugador que camina con shift **directo hacia** un enemigo que no ve, una y otra vez, o que se frena justo detrás de una pared donde hay alguien.

**Límites.** El dato de "visto" es la aproximación del juego, no la geometría exacta. Y el registro no distingue si el jugador camina porque escuchó al rival, porque es la jugada del equipo o porque sabe dónde está: eso solo se ve mirando la demo.

## Mapa

`pkg/radar`, `pkg/core/mapdata.go`, `pkg/gui/ui/js/18-map.js`

**Qué es.** La partida dibujada sobre el radar del mapa, en el reporte de cada demo (sección **"Mapa"**).

**De dónde sale el radar.** De **tu CS:GO instalado**: la imagen del radar y su archivo de posición están en `csgo\resource\overviews`. SourceSight no trae imágenes de Valve: las lee de tu juego, las convierte a PNG una sola vez y las guarda en `%APPDATA%\SourceSight\radar`. Por eso hace falta configurar la carpeta de CS:GO en **Ajustes** (si no, intenta encontrarla sola). Los mapas con varios pisos (como Nuke) tienen un selector **Arriba / Abajo**, y cada punto va al piso que corresponde según su altura.

**Qué se guarda en el análisis.**
- Cada kill, con la posición exacta del que mata y del que muere, el arma y si fue headshot.
- El recorrido de cada jugador: una muestra de su posición por segundo mientras está vivo.
- No cuenta el calentamiento. Un partido completo suma unos 300 KB al reporte (unos 100 KB comprimido).

**Capas** (se prenden y apagan por separado):
- **Kills y muertes:** un círculo del color del equipo donde estaba el que mató, una cruz donde murió la víctima y una línea entre los dos. Con un jugador elegido, sus muertes van en rojo.
- **Recorrido:** por dónde se movió. Con un jugador elegido, su recorrido en blanco; con todos, los de todos más tenues. La línea se corta al morir, al cambiar de ronda o de piso.
- **Mapa de calor:** dónde pasaron más tiempo, con la misma escala de colores que la figura del cuerpo.
- **Incidentes:** un rombo donde estaba el jugador en cada incidente (rojo para triggerbot, silent aim y wallhack; amarillo para el resto).
- **SHIFT:** un triángulo donde empezó cada caso del registro de SHIFT, con una línea punteada hasta cada enemigo que no veía.

**Filtros.** Por jugador y por ronda. Desde el detalle de un jugador, **Ver en el mapa** abre el mapa filtrado por él con su recorrido.

**Click en un punto.** Muestra qué pasó (quién mató a quién, con qué arma y a qué distancia; o el incidente; o el caso de SHIFT) y el botón **Ver** para abrirlo en el juego.

**Límites.** Las posiciones de los incidentes y de los casos de SHIFT salen de la muestra más cercana del recorrido, así que pueden estar corridas hasta medio segundo de movimiento. Las de las kills son exactas.

## Admins y chat

`pkg/modules/staff.go`, `pkg/modules/chat.go`

**Comandos de admin.** Se reconocen comandos de castigo escritos en el chat (`!slay`, `/kick`, `sm_ban`…) y los avisos del servidor de SourceMod (`[SM] Admin: Slayed X.`). Se registran: slay, slap, kick, ban, burn, freeze, beacon, mute, gag, silence, timebomb, firebomb, freezebomb y drug.

**Venganza.** Si un admin castiga (slay, slap, kick, ban, burn, timebomb, firebomb) al jugador que lo mató en los **5 segundos** anteriores, se marca como posible abuso por venganza. SourceMod busca al objetivo por parte del nombre, y SourceSight hace lo mismo para encontrarlo.

**Chat.** Cada mensaje queda en el reporte con su ronda y su tick. Se marcan:
- **Toxicidad:** insultos graves y amenazas (lista en castellano, inglés y portugués).
- **Ghosting por chat:** un jugador muerto que escribe posiciones en el chat general ("están en b", "planted"…).
- **Binds de spam:** el mismo mensaje 3 veces seguidas, cada una a menos de 1 segundo de la anterior.

El chat no cambia el nivel del jugador: es información para el staff.

## Historial y demos repetidas

`pkg/history`

Cada demo analizada queda en el historial, y volver a abrirla es instantáneo (se abre el reporte guardado). Se guardan las últimas 300. Se puede buscar por demo, mapa o jugador, y filtrar por nivel y por categoría.

**Repetidas.** Para que ninguna partida cuente doble en los promedios:
- **Mismo archivo:** se compara el SHA-256 del archivo. Si es una copia exacta con otro nombre o en otra carpeta, pregunta si abrir la guardada, guardar igual o saltear.
- **Misma partida:** se compara una huella de la partida (mapa, duración, rondas y SteamIDs). Además, si coinciden el mapa y las rondas, la duración difiere menos de 2 % y comparten 80 % de los jugadores, también se considera la misma (otra grabación o una copia recortada). Pregunta si reemplazar la anterior, guardar las dos o descartar.

**Re-analizar.** Cuando una versión nueva cambia los detectores, la app avisa qué demos se analizaron con una versión anterior y ofrece re-analizarlas todas.

## Jugadores: promedios entre demos

`pkg/history/players.go`

La pestaña **Jugadores** suma a cada jugador (por SteamID) en todas tus demos:

- en cuántas demos aparece, en cuántas quedó marcado y por qué;
- todos los nombres que usó;
- totales de incidentes, kills cegado, por humo y macro;
- **su promedio contra el resto de los jugadores** (todos los demás, sin contarlo a él, para que la comparación no se corra hacia sus propios números): perfs de bhop, reacción, prefire, impactos a la cabeza, ping, snaps, silent aim y wallhack por demo. En rojo, lo que está muy por encima;
- la figura de dónde pega, comparada con el promedio;
- las observaciones de sus demos sumadas;
- la lista de sus demos, con click para abrir cada una.

**Búsqueda.** Por nombre, SteamID64, link de perfil de Steam, `STEAM_0:X:Y` o `[U:1:N]`.

La vista **Todas las demos** muestra los totales del historial y cómo pegan en promedio todos los jugadores.

## Confirmar un cheater

En el detalle o la ficha de un jugador, **Confirmar cheater** lo marca como confirmado por el staff, con una nota opcional. La app nunca confirma sola. Queda guardado por SteamID y se ve en todas sus demos, y se puede quitar.

## Ver en juego

`pkg/launcher`

El botón **Ver** de cada incidente abre CS:GO en ese momento de la demo:

1. Si la demo no está dentro de la carpeta `csgo` del juego, crea un acceso (*junction*) a la carpeta de la demo dentro de `csgo`. No copia nada, funciona entre discos distintos y no necesita permisos de administrador.
2. Si CS:GO ya está abierto y lo abrió SourceSight, le manda la demo por su consola remota (`-netconport 2121`). Si es la misma demo que ya se está viendo, solo salta al tick.
3. Si no está abierto, lo inicia a través de Steam (`steam.exe -applaunch <id>`, con el App ID que lee de los archivos de Steam). Espera al menú principal, carga la demo con `playdemo` y, cuando termina de cargar el mapa, salta con `demo_gototick`.

Salta unos segundos antes de la jugada (configurable en **Ajustes**, por defecto 2,5 s), convertidos a ticks según el tickrate de la demo.

Si CS:GO está abierto pero lo abriste vos desde Steam, no tiene la consola remota: cerralo una vez y tocá Ver de nuevo, o agregá `-netconport 2121` a sus opciones de lanzamiento.

Cada incidente también trae su comando `demo_gototick` para copiar y pegar a mano.


**Cámara en primera persona.** Cuando el momento es de un jugador (un incidente, un caso de SHIFT, una kill del mapa, un mensaje de chat), después de saltar al tick la app le manda al juego `spec_player_by_accountid` con el número de cuenta del jugador (sale de su SteamID) y `spec_mode 4`: la demo queda mirando desde sus ojos. En la consola del juego aparece una línea `SourceSight: …` con qué incidente es.

**Incidente anterior / siguiente.** Con el juego en primer plano, **Ctrl + Alt + ←** y **Ctrl + Alt + →** saltan al incidente anterior o siguiente, también en primera persona. Recorren la lista del último **Ver** que tocaste: los incidentes de ese jugador, o todos los de la demo si el momento no era de un jugador. Son teclas globales de Windows que recibe SourceSight y reenvía al juego por la consola remota: no se toca ningún bind ni archivo del juego. Se pueden cambiar en **Ajustes** (Ctrl + Shift + flechas, Ctrl + Alt + Re Pág / Av Pág, Alt + Re Pág / Av Pág) o desactivar. No se usan teclas F porque CS:GO ya usa varias (F10 abre el cartel para salir).

**Seguridad.** Los nombres de los jugadores los elige cualquiera, así que antes de mostrarlos en la consola del juego se les sacan `;`, comillas y barras: un nombre no puede colar comandos.

## Archivo automático de demos

`pkg/archive`

**Para qué.** Los servidores borran sus demos a los pocos días (3y33, a los 3). El archivo automático revisa la página de demos, analiza las nuevas y **guarda solo las que tengan algún jugador marcado como cheat o sospechoso**. Así las demos que importan quedan para siempre, aunque el servidor las borre.

**Cómo funciona.**
1. Cada cierto tiempo (30 minutos a 4 horas, por defecto 1 hora) lee la página de demos (por defecto `https://csgo.3y33.com.ar/demos`) y sus páginas siguientes, hasta encontrar una página entera que ya revisó.
2. Se queda con las demos nuevas que pasen los filtros: **servidor** (#1 a #8), **modo** (Competitivo, AutoMIX, Retake, Casual, DeathMatch), **mapa**, y saltea las de menos de 5 MB (calentamientos, servidores vacíos).
3. Las baja **de a una**, empezando por las más viejas (las más cerca de borrarse), directo a la carpeta elegida, y comprueba que llegaron enteras.
4. Las analiza igual que cualquier demo: quedan en el historial y suman a los promedios de los jugadores. Si la partida ya estaba en el historial, no se cuenta dos veces.
5. Si algún jugador quedó como cheat o sospechoso, **el archivo se queda** en la carpeta. Si no, se borra en el momento: el disco no se llena.
6. Recuerda qué demos ya revisó, para no bajarlas de nuevo.

**Dónde corre.** En la app (mientras está abierta) y en el bot (24/7), según dónde lo actives. Si los dos están en la misma PC, un archivo de bloqueo evita que bajen la misma demo a la vez. Es una opción de cada PC y viene apagada: conviene activarla **en una sola**, la del bot, para no bajar lo mismo varias veces.

**Datos de internet.** Bajar no ocupa disco, pero sí consume internet: si se bajara todo serían unos 15 GB por día en 3y33. Los filtros sirven para eso; por ejemplo, solo Competitivo baja mucho menos.

**Configuración.**
- **App:** la sección **Archivo** del menú: carpeta, servidores, modos, mapas, cada cuánto revisar, página de demos y el botón **Revisar ahora**, con el resultado de la última revisión.
- **Bot:** en `bot_config.json`, `archive` (`enabled`, `dir`, `servers`, `modes`, `maps`, `minMB`, `everyMin`, `url`) y `archiveChannelId`, el canal donde avisa cada demo guardada (si falta, el primero de `channelIds`).

**Límites.** Lee la página como la ve un navegador: si el sitio cambia su diseño, hay que adaptar el lector (`ParsePage`). Solo funciona con páginas que dejen bajar las demos sin iniciar sesión.

## Exportar evidencia

- **Exportar evidencia** (en el detalle del jugador) genera un HTML imprimible con los datos del jugador, sus incidentes con tick y comando, y el SHA-256 de la demo, para que otro admin compruebe que es el mismo archivo.
- **Exportar PNG** genera una imagen del reporte de la demo, la misma que manda el bot.

Las exportaciones se guardan en `Documentos\SourceSight`.

## Bot de Discord

`cmd/bot`

1. Creá una aplicación y un bot en el [portal de desarrolladores de Discord](https://discord.com/developers/applications) y activá *Message Content Intent*.
2. Ejecutá `SourceSightBot.exe` una vez: crea `bot_config.json` al lado del ejecutable. Pegá el token ahí. **No compartas ni subas ese archivo.**
3. Volvé a ejecutarlo. Subí una demo al canal y responde con el reporte.

**Qué responde.** Un mensaje con el resultado (con cheat, sospechosos o limpio), los jugadores marcados y su nivel, las observaciones, los primeros incidentes con su comando `demo_gototick`, la imagen del reporte y un botón por cada jugador marcado para ver su detalle.

**Configuración (`bot_config.json`).**

| Campo | Qué hace | Por defecto |
|---|---|---|
| `token` | Token del bot | — |
| `prefix` | Comando de ayuda | `!sourcesight` |
| `channelIds` | Canales donde escucha (vacío = todos) | todos |
| `maxDemoMB` | Tamaño máximo de demo | 250 |
| `maxParallel` | Demos analizadas a la vez | 2 |
| `shiftRadius` | Distancia del registro de SHIFT | 1000 |
| `archive` | Archivo automático (ver [Archivo automático de demos](#archivo-automático-de-demos)) | apagado |
| `archiveChannelId` | Canal donde avisa las demos guardadas | el primero de `channelIds` |

Si llegan más demos que `maxParallel`, las demás esperan su turno y el bot avisa cuántas hay antes. El bot comparte el historial y la base de reincidentes con la app si corren en la misma PC y usuario, y no cuenta dos veces una partida que ya estaba.

## Actualizaciones

Al abrirse, la app consulta si hay una versión nueva en GitHub y ofrece instalarla con un click. Baja el instalador, lo verifica contra el SHA-256 que publica GitHub, actualiza la app, la línea de comandos y el bot, y vuelve a abrir SourceSight. Solo se ofrece en copias instaladas con el instalador.

## Dónde se guardan los datos

Todo queda en tu PC:

| Qué | Dónde |
|---|---|
| Historial, reportes, reincidentes, confirmados y configuración | `%APPDATA%\SourceSight` |
| Exportaciones (evidencia, PNG) | `Documentos\SourceSight` |
| Registro de diagnóstico | `%APPDATA%\SourceSight\sourcesight.log` |

La única conexión a internet es la consulta de actualizaciones a GitHub (y Discord, si usás el bot).

## Qué no se detecta

- **Aimbot suave** (smooth aim), que gira rápido pero repartido en varios ticks: no supera el umbral de snaps.
- **Ghosting por Discord:** que un compañero muerto cante posiciones por voz es legal en muchos servidores, así que disparar donde estaba el asesino no prueba nada. El código existe (`CheckGhosting`) pero no se usa.
- **Visibilidad exacta:** wallhack y prefire usan el dato de "visto" del juego, que es una aproximación, no la geometría real del mapa.
- **Backtrack y no-recoil** todavía no se analizan.

## Compilar y probar

Requiere Go 1.27 o superior.

```
go build -ldflags "-H windowsgui" -o SourceSight.exe ./cmd/sourcesight
go build -o SourceSightCLI.exe ./cmd/sourcesight
go build -o SourceSightBot.exe ./cmd/bot
go test ./...
```

Sin `-X sourcesight/pkg/version.Version=vX.Y.Z` en `-ldflags` la versión queda como `dev` y no busca actualizaciones.

**Tests.**
- `go test ./...` corre todo sin demos reales. Los tests usan una carpeta de datos temporal (`SOURCESIGHT_DATA_DIR`) y se niegan a correr si no queda apuntando ahí, así nunca tocan tu historial.
- Con demos reales: `SOURCESIGHT_TEST_DEMOS=<carpeta con demos> go test ./pkg/parser ./pkg/gui -run "Parallel|Batch"` comprueba que analizar en paralelo da lo mismo que una por una, y prueba el motor completo con una demo repetida y cancelando a mitad de camino.
- El CI corre además `go test -race` en Linux para detectar estado compartido entre análisis.

**Estructura.**

| Carpeta | Qué hay |
|---|---|
| `cmd/sourcesight` | La app (ventana) y la línea de comandos |
| `cmd/bot` | El bot de Discord |
| `pkg/parser` | Lectura de la demo y recorrido tick a tick |
| `pkg/modules` | Un archivo por detector |
| `pkg/core` | Umbrales, niveles, impactos, ping, observaciones |
| `pkg/history`, `pkg/db` | Historial, promedios por jugador y reincidentes |
| `pkg/gui` | La ventana; la página está en `pkg/gui/ui` (HTML, CSS y JS separados) |
| `pkg/launcher` | Ver en juego |
| `pkg/radar` | Radares del juego instalado para el mapa |
| `pkg/archive` | Archivo automático de demos del servidor |

`third_party/go-webview2` es una copia de [jchv/go-webview2](https://github.com/jchv/go-webview2) (MIT) con un parche chico para recibir la ruta real de los archivos que se sueltan en la ventana. El detalle está en `third_party/go-webview2/SOURCESIGHT_PATCH.md`.

El ícono se genera con `go run ./tools/icongen`. En los releases se embebe en los `.exe` con [go-winres](https://github.com/tc-hib/go-winres), y el instalador se arma con [Inno Setup](https://jrsoftware.org/isinfo.php) desde `installer/SourceSight.iss`.

## Publicar una versión

```
git tag vX.Y.Z
git push origin vX.Y.Z
```

GitHub Actions arma el instalador y el bot suelto y publica la versión. Las copias instaladas la ven la próxima vez que se abren.

## Licencia

[MIT con la Commons Clause](LICENSE): podés usar, modificar y compartir SourceSight libremente, incluso en servidores que cobran VIP o skins. Lo que **no se puede es venderlo** ni cobrar por un producto o servicio basado en él.
