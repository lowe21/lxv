package cryptorsa

import (
	"crypto/rsa"
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

func Sign(privateKey *rsa.PrivateKey, content string, opts ...Option) (string, error) {
	return instance().Sign(privateKey, content, opts...)
}

func Verify(publicKey *rsa.PublicKey, content, sign string, opts ...Option) error {
	return instance().Verify(publicKey, content, sign, opts...)
}

func ParsePrivateKey(key string) (*rsa.PrivateKey, error) {
	return instance().ParsePrivateKey(key)
}

func ParsePublicKey(key string) (*rsa.PublicKey, error) {
	return instance().ParsePublicKey(key)
}
