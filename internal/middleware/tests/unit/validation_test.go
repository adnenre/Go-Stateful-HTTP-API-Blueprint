package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rest-api-blueprint/internal/middleware"

	"github.com/stretchr/testify/assert"
)

// mockResolver returns a DTO only for a specific route
func mockResolver(r *http.Request) (any, bool) {
	if r.URL.Path == "/test" && r.Method == "POST" {
		return &struct {
			Name string `json:"name" validate:"required"`
		}{}, true
	}
	return nil, false
}

func TestValidateRequest_Valid(t *testing.T) {
	// ARRANGE
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure body is still readable (restored)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "John", body["name"])
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware.ValidateRequest(mockResolver)(nextHandler)

	body := bytes.NewBufferString(`{"name":"John"}`)
	req := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	// ACT
	handler.ServeHTTP(w, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRequest_Invalid(t *testing.T) {
	// ARRANGE
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	})
	handler := middleware.ValidateRequest(mockResolver)(nextHandler)

	body := bytes.NewBufferString(`{"name":""}`) // empty name fails required
	req := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	// ACT
	handler.ServeHTTP(w, req)

	// ASSERT
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var problem map[string]interface{}
	json.NewDecoder(w.Body).Decode(&problem)
	assert.Equal(t, "/errors/validation", problem["type"])
	assert.Contains(t, problem["errors"], "name")
}

func TestValidateRequest_MalformedJSON(t *testing.T) {
	// ARRANGE
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	})
	handler := middleware.ValidateRequest(mockResolver)(nextHandler)

	body := bytes.NewBufferString(`{"name": "John"`) // missing closing brace
	req := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	// ACT
	handler.ServeHTTP(w, req)

	// ASSERT
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
