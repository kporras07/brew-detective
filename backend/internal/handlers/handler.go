package handlers

import (
	"brew-detective-backend/internal/auth"
	"brew-detective-backend/internal/store"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	Store store.Store
	Auth  auth.Authenticator
}

// NewHandler creates a new Handler with the given store and authenticator.
func NewHandler(s store.Store, a auth.Authenticator) *Handler {
	return &Handler{Store: s, Auth: a}
}
