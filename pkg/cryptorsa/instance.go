package cryptorsa

import (
	"crypto/rsa"
	"sync"
)

var (
	cryptoRSA *CryptoRSA
	mutex     sync.Mutex
)

func instance() *CryptoRSA {
	mutex.Lock()
	defer mutex.Unlock()

	if cryptoRSA == nil {
		cryptoRSA = &CryptoRSA{
			options: newOptions(),
		}
	}

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
