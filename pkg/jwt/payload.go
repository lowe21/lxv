package jwt

import (
	jwt5 "github.com/golang-jwt/jwt/v5"
)

type (
	Payload struct {
		TokenID    string `json:"tokenID"    valid:"required" description:"令牌ID" read-only:"true"`
		IdentityID string `json:"identityID" valid:"required" description:"身份ID" read-only:"true"`
	}

	Claims struct {
		*jwt5.RegisteredClaims
		*Payload
	}
)
