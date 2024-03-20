package lock

type Factory interface {
	NewLock(name string) (Lock, error)
}
