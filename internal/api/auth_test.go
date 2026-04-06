/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ok is a simple handler that returns 200.
func ok(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "POST should pass when no token is configured")
}

func TestAuthMiddleware_GETPassesWithoutAuth(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "secret123")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tests", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "GET should pass without auth header")
}

func TestAuthMiddleware_POSTWithoutHeader(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "secret123")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_POSTWithWrongToken(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "secret123")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_POSTWithCorrectToken(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "secret123")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_DELETEWithCorrectToken(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "secret123")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tests/foo", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_OPTIONSAlwaysPasses(t *testing.T) {
	t.Setenv("KYMAROS_API_TOKEN", "secret123")
	handler := AuthMiddleware(http.HandlerFunc(ok))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/tests", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "OPTIONS should always pass for CORS preflight")
}
