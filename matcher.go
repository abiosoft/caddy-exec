package command

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

var (
	_ caddy.App                = (*Matcher)(nil)
	_ caddy.Module             = (*Matcher)(nil)
	_ caddy.Provisioner        = (*Matcher)(nil)
	_ caddyfile.Unmarshaler    = (*Matcher)(nil)
	_ caddyhttp.RequestMatcher = (*Matcher)(nil)
)

func init() {
	caddy.RegisterModule(Matcher{})
}

// Matcher is a matcher that blocks all request.
// It's primary purpose is to ensure the command is not
// executed when no route/matcher is specified.
type Matcher struct {
	Label string `json:"label,omitempty"`
	log   *zap.Logger
}

// CaddyModule implements caddy.Module
func (Matcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.execnomatch",
		New: func() caddy.Module { return new(Matcher) },
	}
}

// Start ...
func (m Matcher) Start() error {
	m.log.Named(m.Label).Info("starting up...")
	return nil
}

// Stop ...
func (m Matcher) Stop() error {
	m.log.Named(m.Label).Info("shutting down...")
	return nil
}

// Provision implements caddy.Provisioner
func (m *Matcher) Provision(ctx caddy.Context) error {
	if m.Label == "" {
		m.Label = "default"
	}
	m.log = ctx.Logger(m)
	return nil
}

// Match implements caddy.Matcher
func (m Matcher) Match(r *http.Request) bool {
	m.log.Named(m.Label).Info("blocking request to exec handler")
	return false
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (m *Matcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() {
		return d.Err("label required")
	}
	if !d.Args(&m.Label) {
		return d.ArgErr()
	}
	if d.NextBlock(0) {
		return d.Err("no block expected")
	}
	return nil
}
