package command

import (
	"log"

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
	var c moduleConfig

	// logic copied from RegisterHandlerDirective to customize.
	matcherSet, ok, err := h.MatcherToken()
	if err != nil {
		return nil, err
	}
	if ok {
		h.Dispenser.Delete()
		c.handler = true
	}
	h.Dispenser.Reset()

	// parse Caddyfile
	err = c.UnmarshalCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}

	// if there's a matcher, return as http handler
	if ok && matcherSet != nil {
		m := Middleware{moduleConfig: c}
		// matcherSet = caddy.ModuleMap{"path": json.RawMessage{}}
		return h.NewRoute(matcherSet, m), nil
	}

	// otherwise, non-http handler
	// let's handle this ourselves. We're trying to startup an App
	m := Middleware{moduleConfig: c}
	log.Printf("it is parsing... %+v\n", m)

	rawMsg := caddyconfig.JSON(Matcher{}, nil)
	matcherSet = caddy.ModuleMap{"execnomatch": rawMsg}
	return h.NewRoute(matcherSet, m), nil
	// configs = append(configs, c)
	// return nil, nil
}

// UnmarshalCaddyfile configures the global directive from Caddyfile.
// Syntax:
//
//   exec <matcher>|startup|shutdown [<command>] [args...] {
//       command     <text>
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       foreground
//   }
//
func (m *moduleConfig) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if m.handler {
		// command, if present.
		if d.Next() {
			if !d.Args(&m.Command) {
				return d.ArgErr()
			}
		}
		// everything else are args, if present.
		m.Args = d.RemainingArgs()
		return m.unmarshalBlock(d)
	}

	d.Next() // discard the directive

	// not an handler, matcher token missing
	// expect startup|shutdown

	if !d.Next() {
		return d.Err("one of startup|shutdown expected")
	}

	// the first argument must be startup|shutdown
	switch d.Val() {
	case "startup":
		m.Startup = true
	case "shutdown":
		m.Shutdown = true
	default:
		return d.Err("the first argument must be one of <matcher>|startup|shutdown")
	}
	if d.Args(&m.Command) {
		m.Args = d.RemainingArgs()
	}

	return m.unmarshalBlock(d)
}

func (m *moduleConfig) unmarshalBlock(d *caddyfile.Dispenser) error {
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
		case "timeout":
			if !d.Args(&m.Timeout) {
				return d.ArgErr()
			}
		}
	}

	return nil
}
