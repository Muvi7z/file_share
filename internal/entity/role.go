package entity

type Role string

const (
	RoleViewer Role = "viewer"
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
)

func (r Role) IsViewer() bool {
	return r == RoleViewer
}

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}

func (r Role) IsValid() bool {
	return r == RoleViewer || r == RoleAdmin
}
