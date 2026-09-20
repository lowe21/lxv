package jwt

import (
	"sync"
	"time"
)

var (
	jwt  *JWT
	once sync.Once
)

func instance() *JWT {
	once.Do(func() {
		jwt = &JWT{
			options: newOptions(),
		}
	})

	return jwt
}

func Generate(payload *Payload) (string, time.Time, error) {
	return instance().Generate(payload)
}

func Parse(token string, leeway bool) (*Payload, error) {
	return instance().Parse(token, leeway)
}
