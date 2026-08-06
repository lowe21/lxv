package jwt

import (
	jwt5 "github.com/golang-jwt/jwt/v5"
)

type (
	Payload struct {
		UserCode string `json:"userCode" valid:"required" description:"用户编码" read-only:"true"`
		TokenId  string `json:"tokenId"  valid:"required" description:"令牌ID" read-only:"true"`
	}

	Claims struct {
		*jwt5.RegisteredClaims
		*Payload
	}
)
