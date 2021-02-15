# tinyquiz

Tinyquiz is an opinionated minimalistic webapp for running quizzes similar to Kahoot. It aims to provide a pleasant user experience without compromising their privacy or requiring them to have a supercomputer to process multiple layers of unnecessary javascript libraries.

The following is a dev documentation, for those willing to contribute or just curious how it all works. User documentation is a work in progress.

## Overall architecture

Tinyquiz implements the well known MVC (Model View Controller) heavily inspired by the brilliant Alex Edwards' book [Let's Go](https://lets-go.alexedwards.net).

### Directory structure overview

```
├── cmd # each cmd's subdirectory is a runnable package (i. e. cli interface, web server, cron task binary…)
│   └── web
│       └── handlers.go
├── pkg # groups reusable code between different cmd's packages
│   ├── model # business logic especialy related to database
│   │   ├── ent # generated code by the ent ORM
│   │   │   └── schema # the schema description used to generate ent files
│   │   └── model.go # the bussiness logic itself
│   └── rtcomm # realtime communication
├── ui # the view component of MVC
│   ├── html # HTML templates
│   │   ├── *.layout.tmpl.html # reusable layouts
│   │   ├── *.page.tmpl.html # particular pages
│   └── static
└── flake.nix # Nix build instructions useful for both CI and development
```

### cmd/web

Tinyquiz is a rather small projects thus its http handlers (Controller in MVC) tend to be small as well. They are all in `./cmd/web/handlers.go` as methods of the `app` struct, which is one of the ways to do dependency injection in Go.

Although not a requirement, most of them just parse data from the request, call a method of the model and signalize its result, possibly sending back some data.

Few of them send back HTML. That's accomplished by filling in an intermediary struct and passing it to a template to render it.

### pkg/model

Model hosts all business logic, that is, all actions the user can make regardless of the protocol of invocation. Each exported method represent a logical action on its own and must depend on the model struct and its arguments only. This is crucial for testing.

Except for edge cases, each method shall accept context.Context and use it to pass cancellation signals database calls etc.

### pkg/rtcomm

To be as simple, as performant and as reliable as possible, realtime communication in tinyquiz follows a strict pattern. Users commit their actions by sending normal HTTP requests. After processing them, the server optionally sends some data to many clients through a previously established websocket.

The socket is never used to send data to the server.

Sockets are identified by their respective session only. If the user reloads the page (and thus establishes a new socket), both the old one and the new one are used until the old one times out and is garbage collected.

### Testing

Thanks to the strict separation of different parts of MVC, the model can be tested independently of the HTTP handlers. The model is tested against an in-memory SQLite database.

Brief end-to-end testing may be added later.

## Building and running

*Go 1.16 is expected to be released in February 2021. It will bring static files embedding into the binary. Until it happens, the binary has to be run from the rot of the project to be able to find the `ui` directory.*

The reference way to build the app is in the `flake.nix`. For non-Nix users, building shall be pretty trivial though. Just obtain a new enough version of Go (see `go.mod`) and build the runnable package of your choice (usually `go build ./cmd/web`).

Tinyquiz requires Postgresql, though other database systems might be added later thanks to ent. Postgresql configuration is currently hardcoded in the binary.
