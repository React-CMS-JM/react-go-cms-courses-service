package handler

import (
	"context"
	"net/http"

	"react-go-cms-courses-service/internal/infrastructure/httpx"
	"react-go-cms-courses-service/internal/infrastructure/jwt"
)

type ctxKey int

const (
	subjectKey ctxKey = iota
	authedKey
)

func subject(r *http.Request) string {
	value, _ := r.Context().Value(subjectKey).(string)
	return value
}

func authed(r *http.Request) bool {
	value, _ := r.Context().Value(authedKey).(bool)
	return value
}

func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var claims *jwt.Claims
		var err error
		claims, err = h.parse(r)
		if err != nil {
			httpx.WriteError(w, httpx.Unauthorized("Unauthorized"), errorField)
			return
		}
		ctx := context.WithValue(r.Context(), subjectKey, claims.Subject)
		ctx = context.WithValue(ctx, authedKey, true)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) optional(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := httpx.BearerToken(r)
		if raw == "" {
			next(w, r)
			return
		}
		var claims *jwt.Claims
		var err error
		claims, err = jwt.ParseHS256(h.secret, h.issuer, raw)
		if err != nil || claims.Subject == "" {
			httpx.WriteError(w, httpx.Unauthorized("Unauthorized"), errorField)
			return
		}
		ctx := context.WithValue(r.Context(), subjectKey, claims.Subject)
		ctx = context.WithValue(ctx, authedKey, true)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) parse(r *http.Request) (*jwt.Claims, error) {
	raw := httpx.BearerToken(r)
	if raw == "" {
		return nil, httpx.Unauthorized("Unauthorized")
	}
	var claims *jwt.Claims
	var err error
	claims, err = jwt.ParseHS256(h.secret, h.issuer, raw)
	if err != nil || claims.Subject == "" {
		return nil, httpx.Unauthorized("Unauthorized")
	}
	return claims, nil
}
