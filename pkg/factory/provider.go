package factory

type Provider[T any] interface {
	Prepare(name string, overrides map[string]any) (instanceKey string, optionsMap map[string]any, err error)
	New(optionsMap map[string]any) (instance T, err error)
}
