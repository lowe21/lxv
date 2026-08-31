package cryptorsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/lowe21/lxv/pkg/errcode"
)

const (
	SHA1      = "sha1"
	SHA256    = "sha256"
	SHA384    = "sha384"
	SHA512    = "sha512"
	SHA256PSS = "sha256-pss"
	SHA384PSS = "sha384-pss"
	SHA512PSS = "sha512-pss"
)

type CryptoRsa struct {
	options *Options
}

func (c *CryptoRsa) Sign(privateKey, content string, opts ...Option) (sign string, err error) {
	key, err := c.parsePrivateKey(privateKey)
	if err != nil {
		return
	}

	hash, pss, err := c.hash(opts...)
	if err != nil {
		return
	}

	bytes, err := c.hashSum(hash, content)
	if err != nil {
		return
	}

	signed := make([]byte, 0)
	if pss {
		signed, err = rsa.SignPSS(rand.Reader, key, hash, bytes, &rsa.PSSOptions{
			Hash:       hash,
			SaltLength: rsa.PSSSaltLengthEqualsHash,
		})
	} else {
		signed, err = rsa.SignPKCS1v15(rand.Reader, key, hash, bytes)
	}
	if err != nil {
		return
	}

	return gbase64.EncodeToString(signed), nil
}

func (c *CryptoRsa) Verify(publicKey, content, sign string, opts ...Option) (err error) {
	key, err := c.parsePublicKey(publicKey)
	if err != nil {
		return
	}

	hash, pss, err := c.hash(opts...)
	if err != nil {
		return
	}

	bytes, err := c.hashSum(hash, content)
	if err != nil {
		return
	}

	sign = gstr.Trim(sign)
	if gstr.Contains(sign, " ") {
		sign = gstr.Replace(sign, " ", "+")
	}

	signed, err := gbase64.DecodeString(sign)
	if err != nil {
		return
	}
	if len(signed) != key.Size() {
		return errcode.New("invalid RSA signature length")
	}

	if pss {
		return rsa.VerifyPSS(key, hash, bytes, signed, &rsa.PSSOptions{
			Hash:       hash,
			SaltLength: rsa.PSSSaltLengthEqualsHash,
		})
	}

	return rsa.VerifyPKCS1v15(key, hash, bytes, signed)
}

func (c *CryptoRsa) parsePrivateKey(privateKey string) (key *rsa.PrivateKey, err error) {
	der, pemType, err := c.decodeKey(privateKey)
	if err != nil {
		return
	}

	switch pemType {
	case "RSA PRIVATE KEY":
		key, err = c.parsePKCS1PrivateKey(der)
	case "PRIVATE KEY":
		key, err = c.parsePKCS8PrivateKey(der)
	case "ENCRYPTED PRIVATE KEY", "RSA PRIVATE KEY, ENCRYPTED":
		err = errcode.New("RSA private key format is not supported")
		return
	default:
		key, err = c.parsePKCS1PrivateKey(der)
		if err != nil {
			key, err = c.parsePKCS8PrivateKey(der)
		}
	}
	if err != nil {
		return
	}

	if err = c.verifyPrivateKey(key); err != nil {
		return
	}

	key.Precompute()

	return
}

func (c *CryptoRsa) parsePublicKey(publicKey string) (key *rsa.PublicKey, err error) {
	der, pemType, err := c.decodeKey(publicKey)
	if err != nil {
		return
	}

	switch pemType {
	case "RSA PUBLIC KEY":
		key, err = c.parsePKCS1PublicKey(der)
	case "PUBLIC KEY":
		key, err = c.parsePKIXPublicKey(der)
	case "CERTIFICATE", "TRUSTED CERTIFICATE":
		key, err = c.parseCertificatePublicKey(der)
	default:
		key, err = c.parsePKCS1PublicKey(der)
		if err != nil {
			key, err = c.parsePKIXPublicKey(der)
			if err != nil {
				key, err = c.parseCertificatePublicKey(der)
			}
		}
	}
	if err != nil {
		return
	}

	if err = c.verifyPublicKey(key); err != nil {
		return
	}

	return
}

func (c *CryptoRsa) decodeKey(key string) (der []byte, pemType string, err error) {
	key = gstr.Trim(key)
	if key == "" {
		err = errcode.New("RSA key is empty")
		return
	}

	if gstr.Contains(key, "-----BEGIN ") {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			err = errcode.New("decode RSA key from PEM format failed")
			return
		}
		if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
			err = errcode.New("RSA key format is not supported")
			return
		}

		return block.Bytes, block.Type, nil
	}

	key = gstr.Join(gstr.Fields(key), "")
	der, err = gbase64.DecodeString(key)
	if err != nil {
		der, err = base64.RawStdEncoding.DecodeString(key)
	}

	return
}

func (c *CryptoRsa) parsePKCS1PrivateKey(der []byte) (key *rsa.PrivateKey, err error) {
	return x509.ParsePKCS1PrivateKey(der)
}

func (c *CryptoRsa) parsePKCS8PrivateKey(der []byte) (key *rsa.PrivateKey, err error) {
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		err = errcode.New("RSA private key is not PKCS#8 format")
	}

	return
}

func (c *CryptoRsa) parsePKCS1PublicKey(der []byte) (key *rsa.PublicKey, err error) {
	return x509.ParsePKCS1PublicKey(der)
}

func (c *CryptoRsa) parsePKIXPublicKey(der []byte) (key *rsa.PublicKey, err error) {
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return
	}

	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		err = errcode.New("RSA public key is not PKIX format")
	}

	return
}

func (c *CryptoRsa) parseCertificatePublicKey(der []byte) (key *rsa.PublicKey, err error) {
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		return
	}

	key, ok := parsed.PublicKey.(*rsa.PublicKey)
	if !ok {
		err = errcode.New("RSA public key is not certificate format")
	}

	return
}

func (c *CryptoRsa) verifyPrivateKey(key *rsa.PrivateKey) (err error) {
	if key == nil {
		return errcode.New("RSA private key is nil")
	}

	if err = key.Validate(); err != nil {
		return
	}

	if c.options.MinKeyBits > 0 && key.N.BitLen() < c.options.MinKeyBits {
		return errcode.New(fmt.Errorf("RSA private key minimum bits is %d", c.options.MinKeyBits))
	}

	if key.E < 3 || key.E%2 == 0 {
		err = errcode.New("invalid RSA private key exponent")
	}

	return
}

func (c *CryptoRsa) verifyPublicKey(key *rsa.PublicKey) (err error) {
	if key == nil {
		return errcode.New("RSA public key is nil")
	}

	if c.options.MinKeyBits > 0 && key.N.BitLen() < c.options.MinKeyBits {
		return errcode.New(fmt.Errorf("RSA public key minimum bits is %d", c.options.MinKeyBits))
	}

	if key.E < 3 || key.E%2 == 0 {
		err = errcode.New("invalid RSA public key exponent")
	}

	return
}

func (c *CryptoRsa) hash(opts ...Option) (hash crypto.Hash, pss bool, err error) {
	options := *c.options
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	switch options.Hash {
	case SHA1:
		return crypto.SHA1, false, nil
	case SHA256:
		return crypto.SHA256, false, nil
	case SHA384:
		return crypto.SHA384, false, nil
	case SHA512:
		return crypto.SHA512, false, nil
	case SHA256PSS:
		return crypto.SHA256, true, nil
	case SHA384PSS:
		return crypto.SHA384, true, nil
	case SHA512PSS:
		return crypto.SHA512, true, nil
	default:
		err = errcode.New("invalid RSA hash algorithm")
	}

	return
}

func (c *CryptoRsa) hashSum(hash crypto.Hash, content string) (bytes []byte, err error) {
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
