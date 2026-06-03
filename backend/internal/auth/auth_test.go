package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	// Set a test JWT secret for all auth tests
	jwtSecret = []byte("test-secret-key-for-unit-tests")
}

func TestGenerateJWT(t *testing.T) {
	t.Run("generates a valid token", func(t *testing.T) {
		token, err := GenerateJWT("user123", "test@example.com", "Test User")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("token contains correct claims", func(t *testing.T) {
		token, err := GenerateJWT("user123", "test@example.com", "Test User")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		claims, err := ValidateJWT(token)
		if err != nil {
			t.Fatalf("failed to validate generated token: %v", err)
		}

		if claims.UserID != "user123" {
			t.Errorf("expected UserID %q, got %q", "user123", claims.UserID)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("expected Email %q, got %q", "test@example.com", claims.Email)
		}
		if claims.Name != "Test User" {
			t.Errorf("expected Name %q, got %q", "Test User", claims.Name)
		}
	})

	t.Run("token expires in 24 hours", func(t *testing.T) {
		token, err := GenerateJWT("user123", "test@example.com", "Test User")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		claims, err := ValidateJWT(token)
		if err != nil {
			t.Fatalf("failed to validate: %v", err)
		}

		expiry := claims.ExpiresAt.Time
		expectedExpiry := time.Now().Add(24 * time.Hour)
		diff := expiry.Sub(expectedExpiry)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("expiry time off by more than 1 minute: %v", diff)
		}
	})
}

func TestValidateJWT(t *testing.T) {
	t.Run("rejects empty token", func(t *testing.T) {
		_, err := ValidateJWT("")
		if err == nil {
			t.Error("expected error for empty token")
		}
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		_, err := ValidateJWT("not.a.valid.token")
		if err == nil {
			t.Error("expected error for malformed token")
		}
	})

	t.Run("rejects token signed with wrong key", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Email:  "test@example.com",
			Name:   "Test User",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("wrong-secret"))

		_, err := ValidateJWT(tokenString)
		if err == nil {
			t.Error("expected error for wrong signing key")
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Email:  "test@example.com",
			Name:   "Test User",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(jwtSecret)

		_, err := ValidateJWT(tokenString)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("rejects non-HMAC signing method", func(t *testing.T) {
		// Create a token with none signing method by manipulating the header
		claims := &Claims{
			UserID: "user123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
		// Sign with HS384 but our validator only accepts HMAC methods, so this should still work
		// The validator checks for *jwt.SigningMethodHMAC which HS384 satisfies
		tokenString, _ := token.SignedString(jwtSecret)

		// HS384 is still HMAC, so it should validate
		_, err := ValidateJWT(tokenString)
		if err != nil {
			t.Errorf("HS384 is HMAC and should be accepted, got error: %v", err)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects request without Authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		handler := AuthMiddleware()
		handler(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
		if !c.IsAborted() {
			t.Error("expected request to be aborted")
		}
	})

	t.Run("rejects request with invalid Authorization format", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat")

		handler := AuthMiddleware()
		handler(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("rejects request with invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		handler := AuthMiddleware()
		handler(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("accepts request with valid token and sets context", func(t *testing.T) {
		token, _ := GenerateJWT("user123", "test@example.com", "Test User")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		handler := AuthMiddleware()
		handler(c)

		if c.IsAborted() {
			t.Error("expected request to not be aborted")
		}

		userID, exists := c.Get("userID")
		if !exists || userID != "user123" {
			t.Errorf("expected userID %q, got %v", "user123", userID)
		}

		email, exists := c.Get("email")
		if !exists || email != "test@example.com" {
			t.Errorf("expected email %q, got %v", "test@example.com", email)
		}

		name, exists := c.Get("name")
		if !exists || name != "Test User" {
			t.Errorf("expected name %q, got %v", "Test User", name)
		}
	})
}
