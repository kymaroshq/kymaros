package api

import (
	"fmt"
	"net/http"

	"github.com/kymaroshq/kymaros/internal/license"
)

// requireTier wraps a handler and rejects requests when the active license
// tier is below minTier, returning 403 with an upgrade message.
func requireTier(lm *license.LicenseManager, minTier license.Tier, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !lm.HasTier(minTier) {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"error":        fmt.Sprintf("This feature requires Kymaros %s", minTier),
				"currentTier":  string(lm.Tier()),
				"requiredTier": string(minTier),
				"upgradeURL":   "https://kymaros.io/#pricing",
			})
			return
		}
		next(w, r)
	}
}
