package agent

import "container/list"

type Registry struct {
	middlewares *list.List
	agents      map[string]Agent
}

func NewRegistry() Registry {
	return Registry{
		middlewares: list.New(),
		agents:      map[string]Agent{},
	}
}

func (r Registry) Get(key string) (Agent, error) {
	driver, ok := r.agents[key]
	if !ok {
		return nil, ErrDriverNotFound
	}

	return driver, nil
}

func (r Registry) AddMiddleware(middleware Middleware) {
	// LIFO - stack calls
	if r.middlewares.Len() == 0 {
		r.middlewares.PushFront(middleware)
		return
	}
	prevNode := r.middlewares.Front()
	middleware.SetNext(prevNode.Value.(Agent))
	r.middlewares.PushFront(middleware)
}

func (r Registry) Register(key string, agent Agent) {
	if r.middlewares.Len() > 0 {
		mwCopy := r.middlewares.Back().Value.(Middleware)
		mwCopy.SetNext(agent)
		agent = r.middlewares.Front().Value.(Middleware)
	}
	r.agents[key] = agent
}
