package agent

type Registry map[string]Agent

func (r Registry) Get(key string) (Agent, error) {
	driver, ok := r[key]
	if !ok {
		return nil, ErrDriverNotFound
	}

	return driver, nil
}

func (r Registry) Register(key string, agent Agent) {
	r[key] = Retryable{
		Next: agent,
	}
}
