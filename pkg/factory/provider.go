package factory

type Provider[T any] interface {
	Name() string
	Options() any
	New(options any) (T, error)
}
