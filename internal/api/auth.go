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
	"os"
	"strings"
)

// AuthMiddleware protects write endpoints (POST, PUT, DELETE) with a static
// bearer token read from the KYMAROS_API_TOKEN environment variable.
//
// Behaviour:
//   - If KYMAROS_API_TOKEN is empty, all requests pass through (no auth).
//   - GET, HEAD, and OPTIONS requests always pass through (read-only).
//   - Write requests require an "Authorization: Bearer <token>" header.
func AuthMiddleware(next http.Handler) http.Handler {
	token := os.Getenv("KYMAROS_API_TOKEN")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}

		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != token {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
