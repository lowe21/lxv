package cryptorsa

import (
	"bytes"
	"crypto"
	_ "crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/gogf/gf/v2/text/gstr"

	"github.com/lowe21/lxv/pkg/errcode"
)

const (
	MD5       = "md5"
	SHA1      = "sha1"
	SHA256    = "sha256"
	SHA384    = "sha384"
	SHA512    = "sha512"
	SHA256PSS = "sha256-pss"
	SHA384PSS = "sha384-pss"
	SHA512PSS = "sha512-pss"
)

type CryptoRSA struct {
	options *Options
}

func (c *CryptoRSA) Sign(privateKey *rsa.PrivateKey, content string, opts ...Option) (sign string, err error) {
	hash, pss, err := c.hash(opts...)
	if err != nil {
		return
	}

	digest, err := c.hashSum(hash, content)
	if err != nil {
		return
	}

	data := make([]byte, 0)
	if pss {
		data, err = rsa.SignPSS(rand.Reader, privateKey, hash, digest, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       hash,
		})
	} else {
		data, err = rsa.SignPKCS1v15(rand.Reader, privateKey, hash, digest)
	}
	if err != nil {
		return
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (c *CryptoRSA) Verify(publicKey *rsa.PublicKey, content, sign string, opts ...Option) (err error) {
	hash, pss, err := c.hash(opts...)
	if err != nil {
		return
	}

	digest, err := c.hashSum(hash, content)
	if err != nil {
		return
	}

	data, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return
	}
	if len(data) != publicKey.Size() {
		return errcode.New("invalid RSA signature length")
	}

	if pss {
		err = rsa.VerifyPSS(publicKey, hash, digest, data, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       hash,
		})
	} else {
		err = rsa.VerifyPKCS1v15(publicKey, hash, digest, data)
	}

	return
}

func (c *CryptoRSA) ParsePrivateKey(key string) (privateKey *rsa.PrivateKey, err error) {
	der, pemType, err := c.decodeKey(key)
	if err != nil {
		return
	}

	switch pemType {
	case "RSA PRIVATE KEY":
		privateKey, err = c.parsePKCS1PrivateKey(der)
	case "PRIVATE KEY":
		privateKey, err = c.parsePKCS8PrivateKey(der)
	case "ENCRYPTED PRIVATE KEY", "RSA PRIVATE KEY, ENCRYPTED":
		err = errcode.New("RSA private key format is not supported")
		return
	default:
		privateKey, err = c.parsePKCS1PrivateKey(der)
		if err != nil {
			privateKey, err = c.parsePKCS8PrivateKey(der)
		}
	}
	if err != nil {
		return
	}

	if err = c.verifyPrivateKey(privateKey); err == nil {
		privateKey.Precompute()
	}

	return
}

func (c *CryptoRSA) ParsePublicKey(key string) (publicKey *rsa.PublicKey, err error) {
	der, pemType, err := c.decodeKey(key)
	if err != nil {
		return
	}

	switch pemType {
	case "RSA PUBLIC KEY":
		publicKey, err = c.parsePKCS1PublicKey(der)
	case "PUBLIC KEY":
		publicKey, err = c.parsePKIXPublicKey(der)
	case "CERTIFICATE", "TRUSTED CERTIFICATE":
		publicKey, err = c.parseCertificatePublicKey(der)
	default:
		publicKey, err = c.parsePKCS1PublicKey(der)
		if err != nil {
			publicKey, err = c.parsePKIXPublicKey(der)
			if err != nil {
				publicKey, err = c.parseCertificatePublicKey(der)
			}
		}
	}
	if err != nil {
		return
	}

	if err = c.verifyPublicKey(publicKey); err != nil {
		return
	}

	return
}

func (c *CryptoRSA) decodeKey(key string) (der []byte, pemType string, err error) {
	key = gstr.Trim(key)
	if key == "" {
		err = errcode.New("RSA key is empty")
		return
	}

	if gstr.Contains(key, "-----BEGIN ") {
		block, rest := pem.Decode([]byte(key))
		if block == nil {
			err = errcode.New("decode RSA key from PEM format failed")
			return
		}
		if len(bytes.TrimSpace(rest)) > 0 {
			err = errcode.New("RSA key contains unexpected trailing data")
			return
		}
		if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
			err = errcode.New("encrypted RSA key is not supported")
			return
		}

		return block.Bytes, block.Type, nil
	}

	key = gstr.Join(gstr.Fields(key), "")
	der, err = base64.StdEncoding.DecodeString(key)
	if err != nil {
		der, err = base64.RawStdEncoding.DecodeString(key)
	}

	return
}

func (c *CryptoRSA) parsePKCS1PrivateKey(der []byte) (privateKey *rsa.PrivateKey, err error) {
	return x509.ParsePKCS1PrivateKey(der)
}

func (c *CryptoRSA) parsePKCS8PrivateKey(der []byte) (privateKey *rsa.PrivateKey, err error) {
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return
	}

	privateKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		err = errcode.New("PKCS#8 private key is not RSA")
	}

	return
}

func (c *CryptoRSA) parsePKCS1PublicKey(der []byte) (publicKey *rsa.PublicKey, err error) {
	return x509.ParsePKCS1PublicKey(der)
}

func (c *CryptoRSA) parsePKIXPublicKey(der []byte) (publicKey *rsa.PublicKey, err error) {
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return
	}

	publicKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		err = errcode.New("PKIX public key is not RSA")
	}

	return
}

func (c *CryptoRSA) parseCertificatePublicKey(der []byte) (publicKey *rsa.PublicKey, err error) {
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		return
	}

	publicKey, ok := parsed.PublicKey.(*rsa.PublicKey)
	if !ok {
		err = errcode.New("X.509 certificate public key is not RSA")
	}

	return
}

func (c *CryptoRSA) verifyPrivateKey(privateKey *rsa.PrivateKey) (err error) {
	if privateKey == nil {
		return errcode.New("RSA private key is nil")
	}

	if err = privateKey.Validate(); err != nil {
		return
	}

	if c.options.MinKeyBits > 0 && privateKey.N.BitLen() < c.options.MinKeyBits {
		return errcode.New(fmt.Errorf("RSA private key minimum bits is %d", c.options.MinKeyBits))
	}

	if privateKey.E < 3 || privateKey.E%2 == 0 {
		err = errcode.New("invalid RSA private key exponent")
	}

	return
}

func (c *CryptoRSA) verifyPublicKey(publicKey *rsa.PublicKey) (err error) {
	if publicKey == nil {
		return errcode.New("RSA public key is nil")
	}

	if c.options.MinKeyBits > 0 && publicKey.N.BitLen() < c.options.MinKeyBits {
		return errcode.New(fmt.Errorf("RSA public key minimum bits is %d", c.options.MinKeyBits))
	}

	if publicKey.E < 3 || publicKey.E%2 == 0 {
		err = errcode.New("invalid RSA public key exponent")
	}

	return
}

func (c *CryptoRSA) hash(opts ...Option) (hash crypto.Hash, pss bool, err error) {
	options := *c.options
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	switch options.HashType {
	case MD5:
		hash = crypto.MD5
	case SHA1:
		hash = crypto.SHA1
	case SHA256, SHA256PSS:
		hash = crypto.SHA256
		pss = options.HashType == SHA256PSS
	case SHA384, SHA384PSS:
		hash = crypto.SHA384
		pss = options.HashType == SHA384PSS
	case SHA512, SHA512PSS:
		hash = crypto.SHA512
		pss = options.HashType == SHA512PSS
	default:
		err = errcode.New("invalid RSA hash algorithm")
	}

	return
}

func (c *CryptoRSA) hashSum(hash crypto.Hash, content string) (digest []byte, err error) {
	if !hash.Available() {
		err = errcode.New("RSA hash algorithm is not available")
		return
	}

	h := hash.New()
	if _, err = h.Write([]byte(content)); err != nil {
		return
	}

	return h.Sum(nil), nil
}
