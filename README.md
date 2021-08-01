# disstate

[![PkgGoDev](https://pkg.go.dev/badge/github.com/mavolin/disstate)](https://pkg.go.dev/github.com/mavolin/disstate)

Disstate is an alternative state, with a more advanced event system as well as a different approach to sharding.
The API is the same as the one of arikawa's `State`, only adding handlers and gateway commands work differently.

## Changes

There are four major changes to the event system of arikawa:

1.  Handlers take a `*state.State` as first argument.
2. There is support for middlewares both on a global, and a per-handler level
3. All events have new types, that contain a `Base` which is a key-value store, that allows you to pass information from your middleware to the handler.
4. Integrated error and panic handling.
Handlers and middlewares can optionally have an `error` return type.
If a handler or middleware returns an error, it will be given to the event system's error handler. 
If a middleware returns an error, all other middlewares and the handler won't be called.
Similarly, if a middleware or handler panics, the panic will be recovered and handled by the panic handler.
