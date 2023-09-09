package queue

import (
	"fmt"
)

var (
	embeddedServiceModule = fmt.Sprintf("%s.%s", "boltzmann", "queue.service.embedded")
	redisServiceModule    = fmt.Sprintf("%s.%s", "boltzmann", "queue.service.redis")
)
