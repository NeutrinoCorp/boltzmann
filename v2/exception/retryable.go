package exception

type Retryable struct {
	Parent error
}

var _ error = Retryable{}

func (r Retryable) Error() string {
	return r.Parent.Error()
}
