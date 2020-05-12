# caddy-exec

Caddy v2 module for running one-off commands. 

## Installation

```
xcaddy build v2.0.0 \
    --with github.com/abiosoft/caddy-exec
```

## Usage 

Commands can be configured to be triggered by an http endpoint or during startup and shutdown.

They can also be configured to run in the background or foreground and to be terminated after a timeout. 

Beware, startup commands running on foreground will prevent Caddy from starting if they exit with an error.

### Caddyfile
```
exec [<matcher>] [<command> [args...]] {
    command     command
    args        args...
    directory   directory
    timeout     timeout
    foreground
    startup
    shutdown
}
```
* **matcher** - [caddyfile matcher](https://caddyserver.com/docs/caddyfile/matchers). When set, this command runs when there is an http request at the current route or the specified matcher. You may leverage [request matchers](https://caddyserver.com/docs/caddyfile/matchers) to protect the endpoint.
* **command** - command to run
* **args...** - command arguments
* **directory** - directory to run the command from
* **timeout** - timeout to terminate the command's process. Default is 10s.
* **foreground** - if present, runs the command in the foreground. For commands at http endpoints, the command will exit before the http request is responded to.
* **startup** - if present, run the command at startup. Disables http endpoint.
* **shutdown** - if present, run the command at shutdown. Disables http endpoint.

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

### API/JSON

`exec` is somewhat unique in that it can be configured in two ways with JSON. Configuring with Caddyfile abstracts this from the user but the API gives more control.

1. As a top level app for `startup` and `shutdown` commands.

```jsonc
{
  "apps": {
    "http": { ... },
    // app configuration
    "exec": {
      // list of commands
      "commands": [
        // command configuration
        {
          // command to execute
          "command": "hugo",
          // [optional] command arguments
          "args": [
            "generate",
            "--destination=/home/user/site/public"
          ],
          // when to run the command, can include 'startup' or 'shutdown'
          "at": ["startup"],

          // [optional] directory to run the command from. Default is the current directory.
          "directory": "",
          // [optional] if the command should run on the foreground. Default is false.
          "foreground": false,
          // [optional] timeout to terminate the command's process. Default is 10s.
          "timeout": "10s",
        }
      ]
    }
  }
}

```

2. As an handler within a route for commands that get triggered by an http endpoint.

```jsonc

{
  ...
  "routes": [
    {
      "handle": [
        // exec configuration for an endpoint route
        {
          // required to inform caddy the handler is `exec`
          "handler": "exec",
          // command to execute
          "command": "git",
          // command arguments
          "args": ["pull", "origin", "master"],

          // [optional] directory to run the command from. Default is the current directory.
          "directory": "/home/user/site/public",
          // [optional] if the command should run on the foreground. Default is false.
          "foreground": true,
          // [optional] timeout to terminate the command's process. Default is 10s.
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
## Dynamic Configuration

Caddy supports dynamic zero-downtime configuration reloads and it is possible to modify `exec`'s configurations at runtime.

`exec` intelligently determines when Caddy is starting and shutting down. i.e. startup and shutdown commands do not get triggered during configuration reload, only during Caddy's startup and shutdown.

Therefore, you are recommended to dynamically configure only commands triggered by http endpoints for more predictable behaviour.

## License

Apache 2
