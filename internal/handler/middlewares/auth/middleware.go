package auth

import "file_share/internal/service/roles"

type Middleware struct {
	rolesProvider roles.RolesProvider
}

func NewMiddleware(rolesProvider roles.RolesProvider) *Middleware {

}
