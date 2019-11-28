package servers

import "net/url"

type EmptyOption struct{}

func (EmptyOption) apply(*APIServer) (err error) { return }

type funcOption struct {
	f func(*APIServer) error
}

func (fo *funcOption) apply(h *APIServer) error {

	return fo.f(h)
}

func newFuncOption(f func(*APIServer) error) *funcOption {
	return &funcOption{
		f: f,
	}
}

type OAuth2Option struct {
	AuthServer *url.URL
	Audience   string
	Issuer     string
	ApiSecret  []byte
}
