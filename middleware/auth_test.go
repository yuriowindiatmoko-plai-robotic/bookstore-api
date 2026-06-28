package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-key-for-testing"

func setupTestEnv(t *testing.T) {
	t.Helper()
	os.Setenv("JWT_SECRET", testJWTSecret)
	t.Cleanup(func() {
		os.Unsetenv("JWT_SECRET")
	})
}

func generateTestToken(userID uint, email string, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"email":   email,
		"exp":     time.Now().Add(exp).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(testJWTSecret))
}

func setupTestContext(token string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		c.Request.Header.Set("Authorization", "Bearer "+token)
	}
	return w, c
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	setupTestEnv(t)

	token, err := generateTestToken(42, "test@example.com", 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, c := setupTestContext(token)
	middleware := AuthMiddleware()
	middleware(c)

	if c.IsAborted() {
		t.Fatal("Middleware should not abort for valid token")
	}

	userID, exists := c.Get("user_id")
	if !exists {
		t.Fatal("user_id should be set in context")
	}

	if userID.(uint) != 42 {
		t.Errorf("Expected user_id 42, got %v", userID)
	}

	email, exists := c.Get("email")
	if !exists {
		t.Fatal("email should be set in context")
	}

	if email.(string) != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", email)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	setupTestEnv(t)

	w, c := setupTestContext("")
	middleware := AuthMiddleware()
	middleware(c)

	if !c.IsAborted() {
		t.Fatal("Middleware should abort for missing header")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	setupTestEnv(t)

	w, c := setupTestContext("invalid-token-value")
	middleware := AuthMiddleware()
	middleware(c)

	if !c.IsAborted() {
		t.Fatal("Middleware should abort for invalid token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_NonHMACSigningMethod(t *testing.T) {
	setupTestEnv(t)

	// jwt.SigningMethodNone fails the HMAC type assertion in the middleware
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": float64(1),
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	w, c := setupTestContext(tokenString)
	middleware := AuthMiddleware()
	middleware(c)

	if !c.IsAborted() {
		t.Fatal("Middleware should abort for non-HMAC signing method")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	setupTestEnv(t)

	_, c := setupTestContext("")
	c.Request.Header.Set("Authorization", "Bearer ")

	middleware := AuthMiddleware()
	middleware(c)

	if !c.IsAborted() {
		t.Fatal("Middleware should abort for empty token")
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	setupTestEnv(t)

	token, err := generateTestToken(1, "expired@example.com", -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	w, c := setupTestContext(token)
	middleware := AuthMiddleware()
	middleware(c)

	if !c.IsAborted() {
		t.Fatal("Middleware should abort for expired token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidBearerPrefix(t *testing.T) {
	setupTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
	c.Request.Header.Set("Authorization", "Token some-token")

	middleware := AuthMiddleware()
	middleware(c)

	if !c.IsAborted() {
		t.Fatal("Middleware should abort for invalid Bearer prefix")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
