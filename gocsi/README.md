GoCSI is a Go-based framework
that enables developers to build stand-alone CSI endpoints and reuse the
same code to create Go shared objects (plug-ins). The "gocsi" framework
can load the Go plug-ins and communicate with them using gRPC via
in-memory, piped-based network connections.

This framework can be used to provide centralized throttling,
authentication, logging, etc. -- all the while still encouraging CSI
endpoints to be built as stand-along programs.

The "gocsi" library looks at the environment variable `CSI_PLUGINS` for
a CSV list of shared object file paths. These files are expected to be
Go shared object files and will be loaded as plug-ins. A "gocsi" plug-in
should export the following symbol:

```go
Endpoints map[string]func() interface{}
```

The `Endpoints` symbol is a map of endpoint providers to the functions
used to construct them. The constructed object should adhere to the
following interface:

```go
// Endpoint is a gRPC server that provides the CSI Controller,
// Identity, and Node services.
type Endpoint interface {
    Init(ctx context.Context) error
    Serve(ctx context.Context, li net.Listener) error
    Shutdown(ctx context.Context) error
}
```

A Go plug-in can provide multiple endpoint providers. The "gocsi"
library will serve the Endpoints with a listener backed by
fully-duplexed, in memory connection courtesy of `net.Pipe()`.
