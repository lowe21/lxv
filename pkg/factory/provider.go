package factory

type Provider[T any] interface {
	Options() (options any)
	New(options any) (instance T, err error)
}
