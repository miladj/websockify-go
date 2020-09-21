websockify-go is simple implementation of [websockify](https://github.com/novnc/websockify) in golang.

It uses [Gorilla WebSocket](https://github.com/gorilla/websocket) (thanks for great work)

```
websockify [options] [source_addr]:source_port target_addr:target_port
```

```
options:
-cert string
        SSL certificate file
  -h    Print Help
  -key string
        SSL key file
  -run-once
        handle a single WebSocket connection and exit
  -v    Verbose
  -web string
        Serve files from DIR.
```
