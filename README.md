# caddy-command

Caddy v2 module for running one-off commands at an endpoint.

## Installation

```
xcaddy build v2.0.0 \
    --with github.com/abiosoft/caddy-command
```

## Usage 

### Caddyfile
```
command [<matcher>] <command> [args...] {
    args        args...
    directory   directory
    timeout     timeout
    foreground
}
```
* **command** - the command to run
* **matcher** - [caddyfile matcher](https://caddyserver.com/docs/caddyfile/matchers) 
* **args...** - the command arguments
* **directory** - the directory to run the command from
* **timeout** - the timeout to terminate the command process. Default is 10s.
* **foreground** - if present, waits for the command to exit before responding to the http request.

#### Example

`command` can be the last action of a route block.

```
route /generate {
    ... # other directives e.g. for authentication
    command hugo generate --destination=/home/user/site/public
}
```

### API
```json
...
{
  "handle": [
    {
      "handler": "subroute",
      "routes": [
        {
          "handle": [
            {
              "command": "hugo",
              "args": [
                "--destination=/home/user/site/public"
              ],
              "foreground": true,
              "handler": "command",
              "timeout": "5s"
            }
          ]
        }
      ]
    }
  ]
}
```
