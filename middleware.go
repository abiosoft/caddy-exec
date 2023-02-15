package command

import (
	"encoding/json"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

var (
	_ caddy.Module                = (*Middleware)(nil)
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
)

func init() {
	caddy.RegisterModule(Middleware{})
}

// Middleware implements an HTTP handler that runs shell command.
type Middleware struct {
	Cmd
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.exec",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error { return m.Cmd.provision(ctx, m) }

// Validate implements caddy.Validator
func (m Middleware) Validate() error { return m.Cmd.validate() }

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	var resp struct {
		Status string `json:"status,omitempty"`
		Error  string `json:"error,omitempty"`
	}

	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	for index, argument := range m.Args {
		m.Args[index] = repl.ReplaceAll(argument, "")
	}

	err := m.run()

	if err == nil {
		resp.Status = "success"
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = err.Error()
	}

	w.Header().Add("content-type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}

// Cleanup implements caddy.Cleanup
// TODO: ensure all running processes are terminated.
func (m *Middleware) Cleanup() error {
	return nil
}
