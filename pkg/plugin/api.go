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

package plugin

import (
	"net/http"

	kymarosapi "github.com/kymaroshq/kymaros/internal/api"
)

// Queries provides access to Kymaros CRDs for API route handlers.
type Queries = kymarosapi.Queries

// RouteRegistrar is a function that registers additional HTTP routes.
type RouteRegistrar = kymarosapi.RouteRegistrar

// RegisterRoutes adds API routes at the package level.
// Registered routes are applied when the API server starts.
// Call this from an init() function to add custom endpoints.
//
// Example:
//
//	func init() {
//	    plugin.RegisterRoutes(func(mux *http.ServeMux, q *plugin.Queries) {
//	        mux.HandleFunc("GET /api/v1/custom", handleCustom(q))
//	    })
//	}
func RegisterRoutes(r func(mux *http.ServeMux, q *Queries)) {
	kymarosapi.RegisterGlobalRoutes(r)
}
