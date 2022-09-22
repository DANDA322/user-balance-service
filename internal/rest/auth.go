package rest

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/DANDA322/user-balance-service/internal/models"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	jwt.StandardClaims
	AccountId int    `json:"accountId"`
	Role      string `json:"role"`
}

type claimsType string

var ClaimsKey claimsType = `Claims`

func mustGetPublicKey(keyBytes []byte) *rsa.PublicKey {
	if len(keyBytes) == 0 {
		panic("file public.pub is missing or invalid")
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic("unable to decode public key to blocks")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key.(*rsa.PublicKey)
}

func (h *handler) auth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.writeErrResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			h.writeErrResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if headerParts[0] != "Bearer" {
			h.writeErrResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		claims, err := parseToken(headerParts[1], h.pubKey)
		if err != nil {
			h.writeErrResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		sessionInfo := models.SessionInfo{
			AccountId: claims.AccountId,
			Role:      claims.Role,
		}
		r = r.WithContext(context.WithValue(r.Context(), "sessionInfo", sessionInfo))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func parseToken(accessToken string, key *rsa.PublicKey) (*Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("invalid signing method")
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid access token")
}
