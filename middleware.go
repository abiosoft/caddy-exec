package command

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

var (
	_ caddy.Module                = (*Middleware)(nil)
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterDirective("exec", parseHandlerCaddyfile)
}

// Middleware implements an HTTP handler that runs shell command.
type Middleware struct {
	moduleConfig
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.exec",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	appI, err := ctx.App(Matcher{}.CaddyModule().String())
	if err != nil {
		return err
	}
	matcher := appI.(*Matcher)
	matcher.log.Info("yeeeeah.... loaded at runtime")
	// hackyway to provision the main App
	// not sure this is recommended
	// for _, config := range configs {
	// 	appI, err := ctx.App("exec")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	app := appI.(*App)
	// 	app.moduleConfig = config
	// 	if err := app.Provision(ctx); err != nil {
	// 		return err
	// 	}
	// 	if err := app.Validate(); err != nil {
	// 		return err
	// 	}

	// 	configs = nil
	// 	break
	// }

	return m.moduleConfig.provision(ctx, m)
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.log.Info("responding to request.... at", zap.Time("now", time.Now()))
	// if !m.handler {
	// 	// ignore
	// 	return next.ServeHTTP(w, r)
	// }

	var resp struct {
		Status string `json:"status,omitempty"`
		Error  string `json:"error,omitempty"`
	}

	err := m.run()

	if err == nil {
		resp.Status = "success"
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = err.Error()
	}

	return json.NewEncoder(w).Encode(resp)
}

// Cleanup implements caddy.Cleanup
func (m *Middleware) Cleanup() error {
	// m.log.Info("shutting down")
	return nil
}
