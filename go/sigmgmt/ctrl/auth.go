// Copyright 2018 Anapaya Systems

package ctrl

import (
	"encoding/json"
	"net/http"

	"github.com/scionproto/scion/go/sigmgmt/auth"
	"github.com/scionproto/scion/go/sigmgmt/config"
)

type AuthController struct {
	cfg  *config.Global
	auth *auth.JWTAuth
}

func NewAuthController(cfg *config.Global, auth *auth.JWTAuth) *AuthController {
	return &AuthController{
		cfg:  cfg,
		auth: auth,
	}
}

func (c *AuthController) GetToken(w http.ResponseWriter, r *http.Request) {
	var user auth.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}
	status, token := c.auth.Login(&user, c.cfg.Features)
	respond(w, token, status)
}

func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	var user auth.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondError(w, err, JSONDecodeError, http.StatusBadRequest)
		return
	}

	status, token := c.auth.RefreshToken(&user)
	respond(w, token, status)
}
