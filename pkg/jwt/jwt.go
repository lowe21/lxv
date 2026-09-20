package jwt

import (
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	options *Options
}

func (j *JWT) Generate(payload *Payload) (token string, expires time.Time, err error) {
	now := time.Now()

	claims := &Claims{
		RegisteredClaims: &jwtv5.RegisteredClaims{
			Issuer:    j.options.Issuer,
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(j.options.Expires)),
		},
		Payload: payload,
	}

	token, err = jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString(j.options.Key)
	if err != nil {
		return
	}

	expires = claims.ExpiresAt.Time

	return
}

func (j *JWT) Parse(token string, leeway bool) (payload *Payload, err error) {
	claims := &Claims{}

	options := []jwtv5.ParserOption{
		jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
		jwtv5.WithIssuer(j.options.Issuer),
		jwtv5.WithExpirationRequired(),
	}
	if leeway {
		options = append(options, jwtv5.WithLeeway(j.options.Leeway))
	}

	if _, err = jwtv5.ParseWithClaims(token, claims, func(*jwtv5.Token) (any, error) {
		return j.options.Key, nil
	}, options...); err != nil {
		return
	}

	return claims.Payload, nil
}
