# Copia parcheada de jchv/go-webview2

Base: `github.com/jchv/go-webview2 v0.0.0-20260205173254-56598839c808` (licencia MIT, ver `LICENSE`).

Único cambio: `pkg/edge/filedrop.go` + un gancho en `Chromium.MessageReceived` (`pkg/edge/chromium.go`)
para leer la ruta real de los archivos arrastrados a la ventana vía
`chrome.webview.postMessageWithAdditionalObjects`. El resto es idéntico al original.

`edge.DebugLog` permite registrar cada paso del arrastre (la app lo escribe en
`%APPDATA%\SourceSight\sourcesight.log`). Los IID son los publicados en `WebView2.h`;
los de la especificación de diseño en WebView2Feedback eran borradores y no funcionan.
