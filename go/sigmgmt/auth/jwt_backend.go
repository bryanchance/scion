// Copyright 2018 Anapaya Systems

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/scionproto/scion/go/sigmgmt/config"
)

type JWTAuth struct {
	key      []byte
	username string
	password string
	level    int
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JwtToken struct {
	Token string `json:"token"`
}

// JWTExpirationDelta is the number of hours a token is valid
const JWTExpirationDelta = 72

func NewJWTAuth(cfg *config.Global) *JWTAuth {
	return &JWTAuth{
		username: cfg.Username,
		password: cfg.Password,
		key:      []byte(cfg.Key),
		level:    cfg.Features.Level,
	}
}

func (c *JWTAuth) GenerateToken(userUUID string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp":   time.Now().Add(time.Hour * time.Duration(JWTExpirationDelta)).Unix(),
			"iat":   time.Now().Unix(),
			"sub":   userUUID,
			"level": c.level,
		},
	)
	tokenString, err := token.SignedString(c.key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (c *JWTAuth) Login(user *User, level config.FeatureLevel) (int, []byte) {
	if user.Username == c.username && user.Password == c.password {
		token, err := c.GenerateToken(user.Username)
		if err != nil {
			return http.StatusInternalServerError, []byte("")
		} else {
			response, _ := json.Marshal(JwtToken{token})
			return http.StatusOK, response
		}
	}

	return http.StatusUnauthorized, []byte("")
}

func (c *JWTAuth) RefreshToken(user *User) (int, []byte) {
	token, err := c.GenerateToken(user.Username)
	if err != nil {
		return http.StatusInternalServerError, []byte("")
	}
	response, err := json.Marshal(JwtToken{token})
	if err != nil {
		return http.StatusInternalServerError, []byte("")
	}
	return http.StatusOK, response
}

func (c *JWTAuth) RequiredAuthenticated(w http.ResponseWriter, r *http.Request,
	next http.HandlerFunc) {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	splitToken := strings.Split(tokenString, "Bearer")
	if len(splitToken) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bearer := strings.TrimSpace(splitToken[1])
	token, err := jwt.Parse(bearer, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return c.key, nil
	})
	// Could also test claims here
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	next(w, r)
}
