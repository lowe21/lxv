package factory

import (
	"github.com/gogf/gf/v2/util/gutil"
)

type Provider[T, O any] interface {
	Name() string
	Options() *O
	New(options *O) (T, error)
}

type providerEntry[T any] interface {
	Options() any
	New(options any) (T, error)
}

type providerAdapter[T, O any] struct {
	provider Provider[T, O]
}

func (p *providerAdapter[T, O]) Options() any {
	options := p.provider.Options()
	if options != nil {
		options = gutil.Copy(options).(*O)
	}

	return options
}

func (p *providerAdapter[T, O]) New(options any) (T, error) {
	return p.provider.New(options.(*O))
}
