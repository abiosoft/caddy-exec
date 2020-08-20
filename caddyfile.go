package command

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

// parseHandlerCaddyfile unmarshals tokens from h into a new Middleware.
func parseHandlerCaddyfile(h httpcaddyfile.Helper) ([]httpcaddyfile.ConfigValue, error) {
	if !h.Next() {
		return nil, h.ArgErr()
	}
	var c Cmd

	// logic copied from RegisterHandlerDirective to customize.
	matcherSet, ok, err := h.MatcherToken()
	if err != nil {
		return nil, err
	}
	if ok {
		h.Dispenser.Delete()
	}
	h.Dispenser.Reset()

	// parse Caddyfile
	err = c.UnmarshalCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}

	// if there's a matcher, return as http handler
	if c.isRoute() {
		m := Middleware{Cmd: c}
		return h.NewRoute(matcherSet, m), nil
	}

	// otherwise, non-http handler
	// wrap with a NoOpMatcher to prevent http requests.
	m := Middleware{Cmd: c}

	rawMsg := caddyconfig.JSON(NoOpMatcher{}, nil)
	matcherSet = caddy.ModuleMap{
		NoOpMatcher{}.CaddyModule().ID.Name(): rawMsg,
	}

	return h.NewRoute(matcherSet, m), nil

}

// UnmarshalCaddyfile configures the global directive from Caddyfile.
// Syntax:
//
//   exec [<matcher>] [<command>] [args...] {
//       command     <text>
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       foreground
//       startup
//       shutdown
//   }
//
func (m *Cmd) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// command, if present.
	if d.Next() {
		if !d.Args(&m.Command) {
			return d.ArgErr()
		}
	}
	// everything else are args, if present.
	m.Args = d.RemainingArgs()

	// parse the next block
	return m.unmarshalBlock(d)
}

func (m *Cmd) unmarshalBlock(d *caddyfile.Dispenser) error {
	for d.NextBlock(0) {
		switch d.Val() {
		case "command":
			if m.Command != "" {
				return d.Err("command specified twice")
			}
			if !d.Args(&m.Command) {
				return d.ArgErr()
			}
		case "args":
			if len(m.Args) > 0 {
				return d.Err("args specified twice")
			}
			m.Args = d.RemainingArgs()
		case "directory":
			if !d.Args(&m.Directory) {
				return d.ArgErr()
			}
		case "foreground":
			m.Foreground = true
		case "startup":
			m.At = append(m.At, "startup")
		case "shutdown":
			m.At = append(m.At, "shutdown")
		case "timeout":
			if !d.Args(&m.Timeout) {
				return d.ArgErr()
			}
		}
	}

	return nil
}
