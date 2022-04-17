package authorization

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/gerladeno/authorization-service/pkg/common"
	"github.com/gerladeno/authorization-service/pkg/models"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

type Authenticator interface {
	Authenticate(ctx context.Context, phone, code string) error
	VerifyPhone(ctx context.Context, phone string) error
}

type Claims struct {
	jwt.StandardClaims
	ID string `json:"id"`
}

type Authorizer struct {
	log  *logrus.Entry
	auth Authenticator
	key  *rsa.PrivateKey
}

func New(log *logrus.Logger, auth Authenticator, key string) *Authorizer {
	a := Authorizer{
		log:  log.WithField("module", "authorizer"),
		auth: auth,
		key:  mustGetPrivateKey(key),
	}
	return &a
}

func (a *Authorizer) SignIn(ctx context.Context, user *models.User, code string) (string, error) {
	err := a.auth.Authenticate(ctx, user.Phone, code)
	switch {
	case err == nil:
	case errors.Is(err, common.ErrUnauthenticated):
		return "", err
	default:
		err = fmt.Errorf("err authenticating %s: %w", user.Phone, err)
		a.log.Warn(err)
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &Claims{
		StandardClaims: jwt.StandardClaims{},
		ID:             user.ID,
	})
	return token.SignedString(a.key)
}

func (a *Authorizer) StartAuthentication(ctx context.Context, user *models.User) error {
	return a.auth.VerifyPhone(ctx, user.Phone)
}

func mustGetPrivateKey(encodedKey string) *rsa.PrivateKey {
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		panic(err)
	}
	if len(keyBytes) == 0 {
		panic("env PRIVATE_SIGNING_KEY not set")
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic("unable to decode private key to blocks")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key
}

func (a *Authorizer) ParseToken(accessToken string) (string, error) {
	return parseToken(accessToken, a.key)
}

func parseToken(accessToken string, key *rsa.PrivateKey) (string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, common.ErrInvalidSigningMethod
		}
		return &key.PublicKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.ID, nil
	}
	return "", common.ErrInvalidAccessToken
}
