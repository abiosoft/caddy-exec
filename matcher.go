package command

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

var (
	_ caddy.Module             = (*NopMatcher)(nil)
	_ caddy.Provisioner        = (*NopMatcher)(nil)
	_ caddyhttp.RequestMatcher = (*NopMatcher)(nil)
)

func init() {
	caddy.RegisterModule(NopMatcher{})
}

// NopMatcher is a matcher that blocks all request.
// It's primary purpose is to ensure the command is not
// executed when no route/matcher is specified.
// Limitation of Caddyfile config. JSON/API config do not need this.
type NopMatcher struct {
	Label string `json:"label,omitempty"`
	log   *zap.Logger
}

// CaddyModule implements caddy.Module
func (NopMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.execnopmatch",
		New: func() caddy.Module { return new(NopMatcher) },
	}
}

// Provision implements caddy.Provisioner
func (m *NopMatcher) Provision(ctx caddy.Context) error {
	if m.Label == "" {
		m.Label = "default"
	}
	m.log = ctx.Logger(m)
	return nil
}

// Match implements caddy.Matcher
func (m NopMatcher) Match(r *http.Request) bool {
	// block all requests
	return false
}
