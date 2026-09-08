package web

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/api"
)

func pagePermissions(r *http.Request) string {
	principal, ok := api.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	permissions := make([]string, 0, len(principal.Permissions))
	for permission := range principal.Permissions {
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)
	return strings.Join(permissions, ",")
}

func pageCan(r *http.Request, permission string) bool {
	principal, ok := api.PrincipalFromContext(r.Context())
	return ok && principal.Permissions[permission]
}

func requirePage(permissions []string, message string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, permission := range permissions {
			if pageCan(r, permission) {
				next(w, r)
				return
			}
		}
		http.Redirect(w, r, "/config?denied="+url.QueryEscape(message), http.StatusSeeOther)
	}
}
