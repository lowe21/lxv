package jwt

import (
	"sync"
	"time"
)

var (
	jwt   *JWT
	mutex sync.Mutex
)

func instance() *JWT {
	mutex.Lock()
	defer mutex.Unlock()

	if jwt == nil {
		jwt = &JWT{
			options: newOptions(),
		}
	}

	return jwt
}

func Generate(payload *Payload) (string, time.Time, error) {
	return instance().Generate(payload)
}

func Parse(token string, leeway bool) (*Payload, error) {
	return instance().Parse(token, leeway)
}
