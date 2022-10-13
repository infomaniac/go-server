> Disclaimer: This library is a work in progress and I will expand/enhance as I see fit for my current projects and needs. Use at your own peril. 
# go-server

This library is intended to ease the use of running HTTP Servers for your go services. It abstracts away all the nitty gritty stuff like handling shutdown signals, and lets you focus on implementing your specific `http.Handler`s.

Use the `PORT` environment variable to set the tcp port the server will listen to.  (When using e.g. GCP Cloud Run, this is set by the runtime environment.)

# Usage
```go
s := stdserver.New()
err := s.Run(myHandler, myHealthzHandler, myDebugHandler)
```

On a clean shutdown, the `Run` function will return `nil` – if something doesn't work as expected, you will receive the corresponding `error`. Typical cases for this are, if the listener cannot be started due to another process already listening on the same port.

## Handlers
### handler
Here is where your application logic lives. Implement whatever is needed for your service run as expected. No magic here.

### healthzHandler
You can define a custom health check that will be called when `/healthz` is requested. This is typically used to monitor your service in dynamic environments like Kubernetes or Serverless cloud services.

If you don't have the need for a custom `healthz` handler, you can pass `nil` and the default handler will be used, returning "ok", as long as the server is running. You can use this to check the status of dependencies (e.g. database, external service, etc.) and adjust the response accordingly. Meaning, if the database your service relies upon is not reachable, your service may or may not be functioning correctly.

### debugHandler
Here you have the optional chance to serve custom debug information about your application. Provide whatever information you wish to expose on the `/debug` path.

## Debug Mode – `pprof` and `expvar`
If you choose to enable "debug mode" by setting the debug field to `true` before starting the server.
```go
s := stdserver.New()
s.Debug = true
```
The server will then expose
- `/debug/vars/` for `expvar` handler
- `/debug/pprof/` for `pprof` handler

You can use this to expose any data using [`expvar.Publish`][1].

# Logging


# Stats
The server tracks 
- how many requests have been served since starting the server 
- how many client connections are currently open

per default.  
Both these values are exposed via `/debug/vars` when the debug mode is enabled.

# Issues, bugs, etc.
If you find a bug or have a feature request, please file an issue on the Github project and I am happy to look into it.

[1]:	https://pkg.go.dev/expvar@go1.19.2#Publish