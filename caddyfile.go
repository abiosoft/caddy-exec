package command

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

// parseHandlerCaddyfile unmarshals tokens from h into a new Middleware.
func parseHandlerCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Handler
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// UnmarshalCaddyfile configures the plugin from Caddyfile.
// Syntax:
//
//   command [<matcher>] <command> [args...] {
//       args        <text>...
//       directory   <text>
//       timeout     <duration>
//       foreground
//   }
//
func (m *Exec) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() {
		return d.ArgErr()
	}

	if !d.Args(&m.Command) {
		return d.ArgErr()
	}
	m.Args = d.RemainingArgs()

	for d.NextBlock(0) {
		switch d.Val() {
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
