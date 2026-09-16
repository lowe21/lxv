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
			options: newOptions(),
		}
	})

	return cryptoRSA
}

func Sign(privateKey, content string, opts ...Option) (string, error) {
	return instance().Sign(privateKey, content, opts...)
}

func Verify(publicKey, content, sign string, opts ...Option) error {
	return instance().Verify(publicKey, content, sign, opts...)
}
