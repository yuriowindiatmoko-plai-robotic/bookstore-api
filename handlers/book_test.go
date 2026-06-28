package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestParseID_ValidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/books/123", nil)
	c.Params = []gin.Param{{Key: "id", Value: "123"}}

	id, ok := parseID(c)

	if !ok {
		t.Fatal("parseID should return true for valid ID")
	}

	if id != 123 {
		t.Errorf("Expected id 123, got %d", id)
	}
}

func TestParseID_InvalidID_NonNumeric(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/books/abc", nil)
	c.Params = []gin.Param{{Key: "id", Value: "abc"}}

	id, ok := parseID(c)

	if ok {
		t.Fatal("parseID should return false for non-numeric ID")
	}

	if id != 0 {
		t.Errorf("Expected id 0, got %d", id)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestParseID_ZeroID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/books/0", nil)
	c.Params = []gin.Param{{Key: "id", Value: "0"}}

	id, ok := parseID(c)

	if ok {
		t.Fatal("parseID should return false for zero ID")
	}

	if id != 0 {
		t.Errorf("Expected id 0, got %d", id)
	}
}

func TestParseID_NegativeID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/books/-1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "-1"}}

	id, ok := parseID(c)

	if ok {
		t.Fatal("parseID should return false for negative ID")
	}

	if id != 0 {
		t.Errorf("Expected id 0, got %d", id)
	}
}

func TestParseID_EmptyString(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/books/", nil)
	c.Params = []gin.Param{{Key: "id", Value: ""}}

	id, ok := parseID(c)

	if ok {
		t.Fatal("parseID should return false for empty string")
	}

	if id != 0 {
		t.Errorf("Expected id 0, got %d", id)
	}
}
