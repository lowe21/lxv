package jwt

import (
	"sync"
	"time"
)

var (
	jwt  *Jwt
	once sync.Once
)

func instance() *Jwt {
	once.Do(func() {
		jwt = &Jwt{
			options: defaultOptions(),
		}
	})

	return jwt
}

func Generate(payload *Payload) (token string, expires time.Time, err error) {
	return instance().Generate(payload)
}

func Parse(token string, leeway bool) (payload *Payload, err error) {
	return instance().Parse(token, leeway)
}
