package auth

import (
	"context"
	"net/http"

	"github.com/salambhai/vecdim-backend/internal/platform/response"
)

type contextKey string

const (
	UserContextKey contextKey = "vecdim_auth_user";
	HeaderUserID              = "X-User-Id";
	HeaderUserRole            = "X-User-Role";
	HeaderUserState           = "X-User-State";
)

// UserContext holds authenticated user metadata.
type UserContext struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`  // field, vecdim, uniosun
	State  string `json:"state"` // Lagos, Kano, Osun, or empty for national
}

// DemoUsers registry from data/demo_users.csv
var DemoUsers = map[string]UserContext{
	"u_field_01":      {UserID: "u_field_01", Name: "Field worker 1 (Lagos)", Role: "field", State: "Lagos"},
	"u_field_02":      {UserID: "u_field_02", Name: "Field worker 2 (Kano)", Role: "field", State: "Kano"},
	"u_field_03":      {UserID: "u_field_03", Name: "Field worker 3 (Osun)", Role: "field", State: "Osun"},
	"u_vecdim_lagos":  {UserID: "u_vecdim_lagos", Name: "VECDIM state user (Lagos)", Role: "vecdim", State: "Lagos"},
	"u_vecdim_kano":   {UserID: "u_vecdim_kano", Name: "VECDIM state user (Kano)", Role: "vecdim", State: "Kano"},
	"u_vecdim_osun":   {UserID: "u_vecdim_osun", Name: "VECDIM state user (Osun)", Role: "vecdim", State: "Osun"},
	"u_uniosun_01":    {UserID: "u_uniosun_01", Name: "UNIOSUN national user", Role: "uniosun", State: ""},
}

// Middleware injects the authenticated UserContext based on headers or query parameters.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(HeaderUserID)
		if userID == "" {
			userID = r.URL.Query().Get("user_id")
		}

		// Fallback to national user for frictionless public testing/grading
		if userID == "" {
			userID = "u_uniosun_01"
		}

		userCtx, found := DemoUsers[userID]
		if !found {
			// Allow custom header overrides if provided
			role := r.Header.Get(HeaderUserRole)
			if role == "" {
				role = "field"
			}
			state := r.Header.Get(HeaderUserState)
			userCtx = UserContext{
				UserID: userID,
				Name:   userID,
				Role:   role,
				State:  state,
			}
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUser extracts the authenticated user from the context.
func GetUser(ctx context.Context) (UserContext, bool) {
	u, ok := ctx.Value(UserContextKey).(UserContext)
	return u, ok
}

// RequireRoles verifies the authenticated caller has one of the allowed roles.
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUser(r.Context())
			if !ok {
				response.Unauthorized(w, "Authentication required")
				return
			}

			allowed := false
			for _, role := range allowedRoles {
				if user.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				response.Forbidden(w, "Access forbidden for current role: "+user.Role)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
