package jwt

import (
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/gogf/gf/v2/os/gtime"
)

type Jwt struct {
	options *Options
}

func (j *Jwt) GenerateToken(payload *Payload) (token string, err error) {
	now := gtime.Now()

	claims := &Claims{
		RegisteredClaims: &jwtv5.RegisteredClaims{
			Issuer:    j.options.Issuer,
			IssuedAt:  jwtv5.NewNumericDate(now.Time),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(j.options.Expires).Time),
		},
		Payload: payload,
	}

	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString(j.options.Key)
}

func (j *Jwt) ParseToken(token string, options ...jwtv5.ParserOption) (payload *Payload, err error) {
	claims := &Claims{}

	if _, err = jwtv5.ParseWithClaims(token, claims, func(token *jwtv5.Token) (any, error) {
		return j.options.Key, nil
	}, append(options, jwtv5.WithIssuer(j.options.Issuer))...); err != nil {
		return
	}

	return claims.Payload, nil
}

func (j *Jwt) ParseRefreshToken(token string, options ...jwtv5.ParserOption) (payload *Payload, err error) {
	return j.ParseToken(token, append(options, jwtv5.WithLeeway(j.options.Refresh))...)
}
