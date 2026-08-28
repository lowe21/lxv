package cryptorsa

import (
	"sync"
)

var (
	cryptoRsa *CryptoRsa
	once      sync.Once
)

func instance() *CryptoRsa {
	once.Do(func() {
		cryptoRsa = &CryptoRsa{
			options: defaultOptions(),
		}
	})

	return cryptoRsa
}

func Sign(privateKey, content string, opts ...Option) (sign string, err error) {
	return instance().Sign(privateKey, content, opts...)
}

func Verify(publicKey, content, sign string, opts ...Option) (err error) {
	return instance().Verify(publicKey, content, sign, opts...)
}
