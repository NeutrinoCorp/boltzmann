package boltzmann

type Identifiable[T comparable] interface {
	GetID() T
}
