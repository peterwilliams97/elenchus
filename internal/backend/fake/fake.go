package fake

// fake is a backend.Backend test double: Complete delegates to a caller-supplied function, so a test
// can drive the mode runners through the real backend seam with canned responses and no network. It
// is the interface-level counterpart to assay's cfg.call func seam — use fake when the code under
// test dispatches through cfg.backend, cfg.call when it takes the func directly.

import "assay/internal/backend"

// Backend answers Complete from Fn. NameStr/ModelStr/Web back the trivial accessors; a nil Fn panics
// on Complete, which is the intended loud failure for a test that forgot to set a response.
type Backend struct {
	NameStr  string
	ModelStr string
	Web      bool
	Fn       func(backend.Request) (backend.Response, error)
}

// New builds a fake with the given name, model, and response function; web search is off.
func New(name, model string, fn func(backend.Request) (backend.Response, error)) *Backend {
	return &Backend{NameStr: name, ModelStr: model, Fn: fn}
}

func (b *Backend) Name() string            { return b.NameStr }
func (b *Backend) Model() string           { return b.ModelStr }
func (b *Backend) SupportsWebSearch() bool { return b.Web }

func (b *Backend) Complete(req backend.Request) (backend.Response, error) { return b.Fn(req) }
