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

Herramienta para staff de servidores de CS:GO: abrís una demo (`.dem`) y te marca qué jugadores tienen comportamientos que no encajan con un humano —autobunny, triggerbot, spinbot, wallhack, macros— con el tick exacto de cada incidente para ir a verlo en el juego.

Corre en Windows, todo local. No sube nada a ningún lado.

## Qué detecta

| Detección | Qué mira | Cuándo marca |
|---|---|---|
| Autobunny | Saltos en el primer tick de contacto con el piso | 75% o más de perfs, o 5 perfs seguidos |
| Triggerbot | Tiempo entre tener un enemigo en la mira y disparar | Promedio menor a 80 ms con desvío menor a 15 ms |
| Aimlock / spinbot | Giros de más de 45° en un solo tick | 5 o más dentro de un segundo, y al menos 10 en la partida |
| Silent aim | Diferencia entre hacia dónde mira y dónde pega | 3 impactos con más de 35° de diferencia |
| Wallhack | Seguir con la mira a un enemigo que no tiene a la vista | 1 s siguiéndolo con 20° de recorrido, 3 veces |
| Macro | Ritmo de disparo de pistolas demasiado regular | Desvío menor a 1,5 ms en 10 disparos |
| Kills cegado / por humo | Kills con flash total o a través de humo sin visión previa | 4 cegado o 5 por humo |
| Admins | `!slay`, `sm_kick`, etc., y avisos `[SM]` del servidor | Castigar a quien lo mató hace menos de 5 s se marca como venganza |
| Chat | Insultos graves, spam de binds | — |

Los resultados son **indicios, no pruebas**. Cada incidente trae su comando `demo_gototick` para que lo confirmes mirando la demo antes de sancionar.

## Instalar

1. Bajá **`SourceSight-Setup-vX.Y.Z.exe`** desde la [última versión](https://github.com/Hdauu/SourceSight/releases/latest).
2. Ejecutalo. Se instala solo para tu usuario, sin pedir permisos de administrador.
3. Puede aparecer *"Windows protegió su PC"* porque el instalador no está firmado. Hacé click en **Más información → Ejecutar de todas formas**.

¿Preferís no instalar? En la misma versión está el `portable.zip`: descomprimilo y abrí `SourceSight.exe`.

Necesita el runtime de WebView2, que ya viene con Windows 10 y 11.

**Actualizaciones:** al abrirse, la app consulta si hay una versión nueva en GitHub y ofrece instalarla con un click. El ejecutable descargado se verifica contra el SHA-256 que publica GitHub.

## Uso

- Soltá una o varias demos (o una carpeta) sobre la ventana, o usá **Abrir demos**. El instalador también agrega *Abrir con → SourceSight* para los `.dem`.
- Cada demo analizada queda en el **historial**; volver a abrirla es instantáneo.
- Click en un jugador para ver sus métricas, los incidentes por ronda y su historial en todas las demos que analizaste.
- **Exportar evidencia** genera un HTML imprimible con los incidentes y el SHA-256 de la demo, para pasarle a otro admin.
- **Ver** abre CS:GO directo en el tick del incidente (configurá la carpeta del juego en Configuración).

Los datos se guardan en `%APPDATA%\SourceSight`, las exportaciones en `Documentos\SourceSight` y un registro de diagnóstico en `%APPDATA%\SourceSight\sourcesight.log`.

## Bot de Discord

1. Creá una aplicación y un bot en el [portal de desarrolladores de Discord](https://discord.com/developers/applications) y activá *Message Content Intent*.
2. Ejecutá `SourceSightBot.exe` una vez: crea `bot_config.json` al lado del ejecutable. Pegá el token ahí. **No compartas ni subas ese archivo.**
3. Volvé a ejecutarlo. Subí una demo al canal y responde con el reporte y un botón por cada jugador marcado.

## Compilar

Requiere Go 1.27 o superior.

```
go build -ldflags "-H windowsgui" -o SourceSight.exe ./cmd/sourcesight
go build -o SourceSightCLI.exe ./cmd/sourcesight
go build -o SourceSightBot.exe ./cmd/bot
go test ./...
```

Sin `-X sourcesight/pkg/version.Version=vX.Y.Z` en `-ldflags` la versión queda como `dev` y no busca actualizaciones.

`third_party/go-webview2` es una copia de [jchv/go-webview2](https://github.com/jchv/go-webview2) (MIT) con un parche chico para recibir la ruta real de los archivos que se sueltan en la ventana. El detalle está en `third_party/go-webview2/SOURCESIGHT_PATCH.md`.

El ícono se genera con `go run ./tools/icongen`. En los releases se embebe en los `.exe` con [go-winres](https://github.com/tc-hib/go-winres), y el instalador se arma con [Inno Setup](https://jrsoftware.org/isinfo.php) desde `installer/SourceSight.iss`.

## Publicar una versión

```
git tag v2.1.0
git push origin v2.1.0
```

GitHub Actions compila los ejecutables, el instalador y el zip, y publica la versión. Las copias instaladas la ven la próxima vez que se abren.

## Cómo funciona

Ver [docs/como-funciona.md](docs/como-funciona.md).

## Licencia

[MIT con la Commons Clause](LICENSE): podés usar, modificar y compartir SourceSight libremente, incluso en servidores que cobran VIP o skins. Lo que **no se puede es venderlo** ni cobrar por un producto o servicio basado en él.
