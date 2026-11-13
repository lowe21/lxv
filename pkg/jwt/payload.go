package jwt

import (
	jwt5 "github.com/golang-jwt/jwt/v5"
)

type (
	Payload struct {
		IdentityID   string `json:"identityID"   valid:"required" read-only:"true" description:"身份ID"`
		IdentityType string `json:"identityType" valid:"required" read-only:"true" description:"身份类型"`
		TokenID      string `json:"tokenID"      valid:"required" read-only:"true" description:"令牌ID"`
		SessionKey   string `json:"sessionKey"   valid:"required" read-only:"true" description:"会话密钥"`
	}

	Claims struct {
		*jwt5.RegisteredClaims
		*Payload
	}
)
