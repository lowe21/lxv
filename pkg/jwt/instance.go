package jwt

import (
	"sync"
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

func GenerateToken(payload *Payload) (token string, err error) {
	return instance().GenerateToken(payload)
}

func ParseToken(token string) (payload *Payload, err error) {
	return instance().ParseToken(token)
}

func ParseRefreshToken(token string) (payload *Payload, err error) {
	return instance().ParseRefreshToken(token)
}
