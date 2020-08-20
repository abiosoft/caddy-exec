package command

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

var (
	_ caddy.Module             = (*NoOpMatcher)(nil)
	_ caddy.Provisioner        = (*NoOpMatcher)(nil)
	_ caddyhttp.RequestMatcher = (*NoOpMatcher)(nil)
)

func init() {
	caddy.RegisterModule(NoOpMatcher{})
}

// NoOpMatcher is a matcher that blocks all requests.
// It's primary purpose is to ensure the command is not
// executed when no route/matcher is specified.
// Limitation of Caddyfile config. JSON/API config do not need this.
type NoOpMatcher struct {
	Label string `json:"label,omitempty"`
	log   *zap.Logger
}

// CaddyModule implements caddy.Module
func (NoOpMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.exec_noop",
		New: func() caddy.Module { return new(NoOpMatcher) },
	}
}

// Provision implements caddy.Provisioner
func (m *NoOpMatcher) Provision(ctx caddy.Context) error {
	if m.Label == "" {
		m.Label = "default"
	}
	m.log = ctx.Logger(m)
	return nil
}

// Match implements caddy.Matcher
func (m NoOpMatcher) Match(r *http.Request) bool {
	// block all requests
	return false
}
