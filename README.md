# caddy-exec

[WIP] 

Caddy v2 module for running one-off commands.

## Installation

```
xcaddy build v2.0.0 \
    --with github.com/abiosoft/caddy-exec
```

## Usage 

### Caddyfile
```
exec [<matcher>] [<at>] <command> [args...] {
    args        args...
    directory   directory
    timeout     timeout
    foreground
}
```
* **matcher** - [caddyfile matcher](https://caddyserver.com/docs/caddyfile/matchers). When set, this command runs when there is an http request at the current route or the specified matcher. You may leverage [request matchers](https://caddyserver.com/docs/caddyfile/matchers) to protect the endpoint.
* **at** - when to run the command. Must be one of `startup` or `shutdown`. This disables http endpoint and only run the command at `at`. Only one of `at` or `matcher` may be used.
* **command** - the command to run
* **args...** - the command arguments
* **directory** - the directory to run the command from
* **timeout** - the timeout to terminate the command process. Default is 10s.
* **foreground** - if present, runs the command in the foreground. Beware, the failure of startup command running in the foreground can prevent Caddy from starting. For commands at http endpoints, the command will exit before the http request is responded to.

#### Example

`exec` can be the last action of a route block.

```
route /generate {
    ... # other directives e.g. for authentication
    exec hugo generate --destination=/home/user/site/public
}
```

Note that Caddy prevents non-standard directives from being used globally in the Caddyfile except when defined with [order](https://caddyserver.com/docs/caddyfile/options) or scoped to a [route](https://caddyserver.com/docs/caddyfile/directives/route). 
route is recommended for `exec`.

### API

`exec` is somewhat unique in that it can be configured in two ways via the API. The Caddyfile config above abstracts this from the user but the API gives more control.

1. As a top level app for `startup` and `shutdown` commands.

```json
{
  "apps": {
    "http": { ... },
    "exec": {
      "commands": [
        {
          "command": "hugo",
          "args": [
            "generate",
            "--destination=/home/user/site/public"
          ],
          "at": "startup"
        }
      ]
    }
  }
}

```

2. As an handler within a route for commands that get triggered by an http endpoint.

```json

{
...
  "routes": [
    {
      "handle": [
        {
          "handler": "exec",
          "command": "git",
          "args": ["pull", "origin", "master"],
          "directory": "/home/user/site/public",
          "foreground": true,
          "timeout": "5s"
        }
      ],
      "match": [
        {
          "path": ["/generate"]
        }
      ]
    }
  ]
}
```

