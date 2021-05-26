package command

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("exec", parseGlobalCaddyfileBlock)
	httpcaddyfile.RegisterHandlerDirective("exec", parseHandlerCaddyfileBlock)
}

func newCommandFromDispenser(d *caddyfile.Dispenser) (cmd Cmd, err error) {
	cmd.UnmarshalCaddyfile(d)
	return
}

// parseHandlerCaddyfileBlock configures the handler directive from Caddyfile.
// Syntax:
//
//   exec [<matcher>] [<command> [<args...>]] {
//       command     <text>
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       log         <log output module>
//       err_log     <log output module>
//       foreground
//       startup
//       shutdown
//   }
//
func parseHandlerCaddyfileBlock(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	cmd, err := newCommandFromDispenser(h.Dispenser)
	return Middleware{Cmd: cmd}, err
}

// parseGlobalCaddyfileBlock configures the "exec" global option from Caddyfile.
// Syntax:
//
//   exec [<command> [<args...>]] {
//       command     <text>...
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       log         <log output module>
//       err_log     <log output module>
//       foreground
//       startup
//       shutdown
//   }
//
func parseGlobalCaddyfileBlock(d *caddyfile.Dispenser, prev interface{}) (interface{}, error) {
	var exec App

	// decode the existing value and merge to it.
	if prev != nil {
		if app, ok := prev.(httpcaddyfile.App); ok {
			if err := json.Unmarshal(app.Value, &exec); err != nil {
				return nil, d.Errf("internal error: %v", err)
			}
		}
	}

	cmd, err := newCommandFromDispenser(d)
	if err != nil {
		return nil, err
	}

	// global block commands are not necessarily bound to a route,
	// should default to running at startup.
	if len(cmd.At) == 0 {
		cmd.At = append(cmd.At, "startup")
	}

	// append command to global exec app.
	exec.Commands = append(exec.Commands, cmd)

	// tell Caddyfile adapter that this is the JSON for an app
	return httpcaddyfile.App{
		Name:  "exec",
		Value: caddyconfig.JSON(exec, nil),
	}, nil
}

// UnmarshalCaddyfile configures the handler directive from Caddyfile.
// Syntax:
//
//   exec [<matcher>] [<command> [<args...>]] {
//       command     <text>
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       log         <log output module>
//       err_log     <log output module>
//       foreground
//       startup
//       shutdown
//   }
//
func (c *Cmd) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// consume "exec", then grab the command, if present.
	if d.NextArg() && d.NextArg() {
		c.Command = d.Val()
	}

	// everything else are args, if present.
	c.Args = d.RemainingArgs()

	// parse the next block
	return c.unmarshalBlock(d)
}

func (c *Cmd) unmarshalBlock(d *caddyfile.Dispenser) error {
	for d.NextBlock(0) {
		switch d.Val() {
		case "command":
			if c.Command != "" {
				return d.Err("command specified twice")
			}
			if !d.Args(&c.Command) {
				return d.ArgErr()
			}
			c.Args = d.RemainingArgs()
		case "args":
			if len(c.Args) > 0 {
				return d.Err("args specified twice")
			}
			c.Args = d.RemainingArgs()
		case "directory":
			if !d.Args(&c.Directory) {
				return d.ArgErr()
			}
		case "foreground":
			c.Foreground = true
		case "startup":
			c.At = append(c.At, "startup")
		case "shutdown":
			c.At = append(c.At, "shutdown")
		case "timeout":
			if !d.Args(&c.Timeout) {
				return d.ArgErr()
			}
		case "log":
			rawMessage, err := c.unmarshalLog(d)
			if err != nil {
				return err
			}
			c.StdWriterRaw = rawMessage
		case "err_log":
			rawMessage, err := c.unmarshalLog(d)
			if err != nil {
				return err
			}
			c.ErrWriterRaw = rawMessage
		default:
			return d.Errf("'%s' not expected", d.Val())
		}
	}

	return nil
}

func (c *Cmd) unmarshalLog(d *caddyfile.Dispenser) (json.RawMessage, error) {
	if !d.NextArg() {
		return nil, d.ArgErr()
	}
	moduleName := d.Val()

	// copied from caddy's source
	// TODO: raise the topic of log re-use by non-standard modules.
	var wo caddy.WriterOpener
	switch moduleName {
	case "stdout":
		wo = caddy.StdoutWriter{}
	case "stderr":
		wo = caddy.StderrWriter{}
	case "discard":
		wo = caddy.DiscardWriter{}
	default:
		modID := "caddy.logging.writers." + moduleName
		unm, err := caddyfile.UnmarshalModule(d, modID)
		if err != nil {
			return nil, err
		}
		var ok bool
		wo, ok = unm.(caddy.WriterOpener)
		if !ok {
			return nil, d.Errf("module %s (%T) is not a WriterOpener", modID, unm)
		}
	}
	return caddyconfig.JSONModuleObject(wo, "output", moduleName, nil), nil
}
