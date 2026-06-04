package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"brew-detective-backend/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var oauthStateStore = &stateStore{
	states: make(map[string]time.Time),
}

type stateStore struct {
	mu     sync.Mutex
	states map[string]time.Time
}

func (s *stateStore) Save(state string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(ttl)
	s.cleanup()
}

func (s *stateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return time.Now().Before(expiry)
}

func (s *stateStore) cleanup() {
	now := time.Now()
	for state, expiry := range s.states {
		if now.After(expiry) {
			delete(s.states, state)
		}
	}
}

var (
	googleOauthConfig *oauth2.Config
	jwtSecret         []byte
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// Authenticator abstracts the OAuth and JWT operations so handlers can be tested
// without hitting Google.
type Authenticator interface {
	GenerateOAuthURL() string
	ValidateOAuthState(state string) bool
	GetUserFromOAuthCode(code string) (*GoogleUser, error)
	GenerateJWT(userID, email, name string) (string, error)
}

// GoogleAuthenticator is the real implementation backed by Google OAuth.
type GoogleAuthenticator struct{}

func (g *GoogleAuthenticator) GenerateOAuthURL() string {
	state := generateOAuthState()
	oauthStateStore.Save(state, 10*time.Minute)
	return googleOauthConfig.AuthCodeURL(state)
}

func (g *GoogleAuthenticator) ValidateOAuthState(state string) bool {
	return oauthStateStore.Validate(state)
}

func (g *GoogleAuthenticator) GetUserFromOAuthCode(code string) (*GoogleUser, error) {
	return GetUserDataFromGoogle(code)
}

func (g *GoogleAuthenticator) GenerateJWT(userID, email, name string) (string, error) {
	return generateJWT(userID, email, name)
}

// generateOAuthState creates a cryptographically secure random state string.
func generateOAuthState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		timestamp := time.Now().UnixNano()
		for i := range b {
			b[i] = byte((timestamp + int64(i)) % 256)
		}
	}
	return fmt.Sprintf("%x", b)
}

// generateJWT is the internal implementation used by both the package-level
// function and the Authenticator interface.
func generateJWT(userID, email, name string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func InitAuth() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		panic("JWT_SECRET environment variable not set")
	}
	jwtSecret = []byte(jwtSecretStr)
}

func GenerateOAuthState(c *gin.Context) string {
	return generateOAuthState()
}

func GetGoogleOauthConfig() *oauth2.Config {
	return googleOauthConfig
}

func GetUserDataFromGoogle(code string) (*GoogleUser, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	var user GoogleUser
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return &user, nil
}

func GenerateJWT(userID, email, name string) (string, error) {
	return generateJWT(userID, email, name)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Next()
	}
}

// AdminMiddleware ensures the user is authenticated and has admin privileges.
// It uses the provided Store to look up user data instead of accessing Firestore directly.
func AdminMiddleware(s store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check if user has admin privileges
		user, err := s.GetUser(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if user.Type != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("userType", user.Type)
		c.Next()
	}
}