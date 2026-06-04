package auth

import "fmt"

// MockAuthenticator is a test implementation of Authenticator.
type MockAuthenticator struct {
	OAuthURL          string
	ValidateState     bool
	GoogleUser        *GoogleUser
	GoogleErr         error
	JWTToken          string
	JWTErr            error
}

func (m *MockAuthenticator) GenerateOAuthURL() string {
	return m.OAuthURL
}

func (m *MockAuthenticator) ValidateOAuthState(state string) bool {
	return m.ValidateState
}

func (m *MockAuthenticator) GetUserFromOAuthCode(code string) (*GoogleUser, error) {
	if m.GoogleErr != nil {
		return nil, m.GoogleErr
	}
	if m.GoogleUser == nil {
		return nil, fmt.Errorf("no user configured in mock")
	}
	return m.GoogleUser, nil
}

func (m *MockAuthenticator) GenerateJWT(userID, email, name string) (string, error) {
	if m.JWTErr != nil {
		return "", m.JWTErr
	}
	if m.JWTToken != "" {
		return m.JWTToken, nil
	}
	return "mock-jwt-token", nil
}
