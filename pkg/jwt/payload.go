package jwt

import (
	jwt5 "github.com/golang-jwt/jwt/v5"
)

type (
	Payload struct {
		TokenId    string `json:"tokenId"    valid:"required" description:"令牌ID" read-only:"true"`
		IdentityId string `json:"identityId" valid:"required" description:"身份ID" read-only:"true"`
		SessionKey string `json:"sessionKey" valid:"required" description:"会话密钥" read-only:"true"`
	}

	Claims struct {
		*jwt5.RegisteredClaims
		*Payload
	}
)
