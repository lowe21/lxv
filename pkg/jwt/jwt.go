package jwt

import (
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/gogf/gf/v2/os/gtime"
)

type Jwt struct {
	options *Options
}

func (j *Jwt) Generate(payload *Payload) (token string, expires time.Time, err error) {
	now := gtime.Now()

	claims := &Claims{
		RegisteredClaims: &jwtv5.RegisteredClaims{
			Issuer:    j.options.Issuer,
			IssuedAt:  jwtv5.NewNumericDate(now.Time),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(j.options.Expires).Time),
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

func (j *Jwt) Parse(token string, leeway bool) (payload *Payload, err error) {
	claims := &Claims{}
	options := []jwtv5.ParserOption{jwtv5.WithIssuer(j.options.Issuer)}
	if leeway {
		options = append(options, jwtv5.WithLeeway(j.options.Leeway))
	}

	if _, err = jwtv5.ParseWithClaims(token, claims, func(token *jwtv5.Token) (any, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, jwtv5.ErrHashUnavailable
		}

		return j.options.Key, nil
	}, options...); err != nil {
		return
	}

	return claims.Payload, nil
}
