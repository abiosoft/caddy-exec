# caddy-exec

Caddy v2 module for running one-off commands.

## Installation

```
xcaddy build \
    --with github.com/abiosoft/caddy-exec
```

## Usage

Commands can be configured to be triggered globally during startup/shutdown or by via a route.

They can also be configured to run in the background or foreground and to be terminated after a timeout.

:warning: startup commands running on foreground will prevent Caddy from starting if they exit with an error.

### Caddyfile

```
exec [<matcher>] [<command> [<args...>]] {
    command     <command> [<args...>]
    args        <args...>
    directory   <directory>
    timeout     <timeout>
    log         <log output module>
    err_log     <log output module>
    foreground
    startup
    shutdown
}
```

- **matcher** - [Caddyfile matcher](https://caddyserver.com/docs/caddyfile/matchers). When set, this command runs when there is an http request at the current route or the specified matcher. You may leverage other matchers to protect the endpoint.
- **command** - command to run
- **args...** - command arguments
- **directory** - directory to run the command from
- **timeout** - timeout to terminate the command's process. Default is `10s`. A timeout of `0` runs indefinitely.
- **log** - [Caddy log output module](https://caddyserver.com/docs/caddyfile/directives/log#output-modules) for standard output log. Defaults to `stderr`.
- **err_log** - [Caddy log output module](https://caddyserver.com/docs/caddyfile/directives/log#output-modules) for standard error log. Defaults to the value of `log` (standard output log).
- **foreground** - if present, runs the command in the foreground. For commands at http endpoints, the command will exit before the http request is responded to.
- **startup** - if present, run the command at startup. Ignored in routes.
- **shutdown** - if present, run the command at shutdown. Ignored in routes.

#### Example

`exec` can run at start via the [global](https://caddyserver.com/docs/caddyfile/options) directive.

```
{
  exec hugo generate --destination=/home/user/site/public {
      timeout 0 # don't timeout
  }
}
```

`exec` can be the last action of a route block.

```
route /update {
    ... # other directives e.g. for authentication
    exec git pull origin master {
        log file /var/logs/hugo.log
    }
}
```

### API/JSON

As a top level app for `startup` and `shutdown` commands.

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
          // [optional] log output module config for standard output. Default is `stderr` module.
          "log": {
            "output": "file",
            "filename": "/var/logs/hugo.log"
          },
          // [optional] log output module config for standard error. Default is the value of `log`.
          "err_log": {
            "output": "stderr"
          }
        }
      ]
    }
  }
}

```

As an handler within a route.

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
          // command arguments it's also possible to use 
          // caddy variables like {http.request.uuid}
          "args": ["pull", "origin", "master", "# {http.request.uuid}"],

          // [optional] directory to run the command from. Default is the current directory.
          "directory": "/home/user/site/public",
          // [optional] if the command should run on the foreground. Default is false.
          "foreground": true,
          // [optional] timeout to terminate the command's process. Default is 10s.
          "timeout": "5s",
          // [optional] log output module config for standard output. Default is `stderr` module.
          "log": {
            "output": "file",
            "filename": "/var/logs/hugo.log"
          },
          // [optional] log output module config for standard error. Default is the value of `log`.
          "err_log": {
            "output": "stderr"
          }
        }
      ],
      "match": [
        {
          "path": ["/update"]
        }
      ]
    }
  ]
}
```

## Dynamic Configuration

Caddy supports dynamic zero-downtime configuration reloads and it is possible to modify `exec`'s configurations at runtime.

`exec` intelligently determines when Caddy is starting and shutting down. i.e. startup and shutdown commands do not get triggered during configuration reload, only during Caddy's actual startup and shutdown.

## License

Apache 2
