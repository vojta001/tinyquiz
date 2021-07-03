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

Model hosts all business logic, that is, all actions the user can make regardless of the protocol of invocation. Each exported method represents a logical action on its own and must depend on the model struct and its arguments only. This is crucial for testing.

Except for edge cases, each method shall accept context.Context and use it to pass cancellation signals database calls etc.

### pkg/rtcomm

To be as simple, as performant and as reliable as possible, realtime communication in tinyquiz follows a strict pattern. Users commit their actions by sending normal HTTP requests. After processing them, the server optionally sends some data to many clients through a previously established websocket.

The socket is never used to send data to the server.

Sockets are identified by their respective session only. If the user reloads the page (and thus establishes a new socket), both the old one and the new one are used until the old one times out and is garbage collected.

### Testing

Thanks to the strict separation of different parts of MVC, the model can be tested independently of the HTTP handlers. The model is tested against an in-memory SQLite database.

Brief end-to-end testing may be added later.

### Security

Tinyquiz can hardly be viewed as critical application or as containing sensitive information, therefore certain trade-offs were accepted.

There is no classical authentication; your identity is determined by the id contained in URL. This is usually viewed as bad practice, because the token gets saved in your browsing history, but in Tinyquiz, it becomes effectively worthless as soon as the quiz ends. On the other hand, it enables you to play multiple games in different browser tabs at the same time and to reaload the tabs anytime without loosing state - all while keeping the implementation very simple.

Games and quizzes aren't protected by any password, just the code. It is pretty secure though, the code is made by concatenating a sequential part (to prevent collisions) and a random part (to make it difficult to guess).

Unlike the previous part, no trade-offs were accepted in the server security. Go is a GCed language doing its best to prevent memory corruption bugs. All database queries are assembled by passing the user supplied input separately thus preventing SQL injection. HTML output is handled by the well tested `html/template` standard library which automatically context-aware escapes included content thus preventing XSS.

## Building and running

The reference way to build the app is in the `flake.nix`. For non-Nix users, building shall be pretty trivial though. Just obtain a new enough version of Go (see `go.mod`) and build the runnable package of your choice (usually `go build ./cmd/web`). You may have to run `go generate ./...` to build the ORM files etc.

Tinyquiz requires Postgresql, though other database systems might be added later thanks to ent. Postgresql configuration is currently hardcoded in the binary.
