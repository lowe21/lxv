package cryptorsa

import (
	"sync"
)

var (
	cryptoRSA *CryptoRSA
	once      sync.Once
)

func instance() *CryptoRSA {
	once.Do(func() {
		cryptoRSA = &CryptoRSA{
			options: defaultOptions(),
		}
	})

	return cryptoRSA
}

func Sign(privateKey, content string, opts ...Option) (sign string, err error) {
	return instance().Sign(privateKey, content, opts...)
}

func Verify(publicKey, content, sign string, opts ...Option) (err error) {
	return instance().Verify(publicKey, content, sign, opts...)
}
