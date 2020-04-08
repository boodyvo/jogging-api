package storage

import (
	"fmt"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
)

// TODO(boodyvo): Move roles and permissions to DB and make migration so could be dynamic

type Action string

const (
	ReadAction   Action = "read"
	UpdateAction Action = "update"
	DeleteAction Action = "delete"
)

type Scope string

const (
	UserScope        Scope = "user"
	TrackingScope    Scope = "tracking"
	PermissionsScope Scope = "permissions"
)

type Resource struct {
	Scope Scope  `json:"scope" bson:"scope"`
	Item  string `json:"item" bson:"item"`
}

// Permissions

type Permission struct {
	Action   Action   `json:"action" bson:"action"`
	Resource Resource `json:"resource" bson:"resource"`
}

var (
	ReadUsersPermission = Permission{
		Action: ReadAction,
		Resource: Resource{
			Scope: UserScope,
			Item:  "*",
		},
	}
	UpdateUsersPermission = Permission{
		Action: UpdateAction,
		Resource: Resource{
			Scope: UserScope,
			Item:  "*",
		},
	}
	DeleteUsersPermission = Permission{
		Action: DeleteAction,
		Resource: Resource{
			Scope: UserScope,
			Item:  "*",
		},
	}

	ReadTrackingsPermission = Permission{
		Action: ReadAction,
		Resource: Resource{
			Scope: TrackingScope,
			Item:  "*",
		},
	}
	UpdateTrackingsPermission = Permission{
		Action: UpdateAction,
		Resource: Resource{
			Scope: TrackingScope,
			Item:  "*",
		},
	}
	DeleteTrackingsPermission = Permission{
		Action: DeleteAction,
		Resource: Resource{
			Scope: TrackingScope,
			Item:  "*",
		},
	}

	ReadPermissionsPermission = Permission{
		Action: ReadAction,
		Resource: Resource{
			Scope: PermissionsScope,
			Item:  "*",
		},
	}
	UpdatePermissionsPermission = Permission{
		Action: UpdateAction,
		Resource: Resource{
			Scope: PermissionsScope,
			Item:  "*",
		},
	}
	DeletePermissionsPermission = Permission{
		Action: DeleteAction,
		Resource: Resource{
			Scope: PermissionsScope,
			Item:  "*",
		},
	}
)

func NewPermission(action Action, scope Scope, item string) Permission {
	return Permission{
		Action: action,
		Resource: Resource{
			Scope: scope,
			Item:  item,
		},
	}
}

func (p *Permission) String() string {
	return fmt.Sprintf("action:%s,scope:%s,items:%v", p.Action, p.Resource.Scope, p.Resource.Item)
}

func (p *Permission) Equal(another Permission) bool {
	return p.String() == another.String()
}

func (p Permission) Include(another Permission) bool {
	if p.Action != another.Action || p.Resource.Scope != another.Resource.Scope {
		return false
	}

	if p.Resource.Item == "*" {
		return true
	}

	return p.Resource.Item == another.Resource.Item
}

func PermissionFromProto(action pb.Action, scope pb.Scope, item string) (Permission, error) {
	var a Action
	var s Scope

	switch action {
	case pb.Action_ACTION_READ:
		a = ReadAction
	case pb.Action_ACTION_UPDATE:
		a = UpdateAction
	case pb.Action_ACTION_DELETE:
		a = DeleteAction
	default:
		return Permission{}, ErrUnknownAction
	}

	switch scope {
	case pb.Scope_SCOPE_USERS:
		s = UserScope
	case pb.Scope_SCOPE_TRACKINGS:
		s = TrackingScope
	case pb.Scope_SCOPE_PERMISSIONS:
		s = PermissionsScope
	default:
		return Permission{}, ErrUnknownScope
	}

	return NewPermission(a, s, item), nil
}

// Roles

type Role struct {
	Name        string       `json:"name" bson:"name"`
	Permissions []Permission `json:"permissions" bson:"permissions"`
}

var (
	UserRole = Role{
		Name:        "UserRole",
		Permissions: []Permission{},
	}
	ManagerRole = Role{
		Name: "ManagerRole",
		Permissions: []Permission{
			ReadUsersPermission, UpdateUsersPermission, DeleteUsersPermission,
		},
	}
	AdminRole = Role{
		Name: "AdminRole",
		Permissions: []Permission{
			ReadUsersPermission, UpdateUsersPermission, DeleteUsersPermission,
			ReadTrackingsPermission, UpdateTrackingsPermission, DeleteTrackingsPermission,
			ReadPermissionsPermission, UpdatePermissionsPermission, DeletePermissionsPermission,
		},
	}
)

func (r *Role) String() string {
	return r.Name
}

func (r *Role) HasPermission(permission Permission) bool {
	for _, p := range r.Permissions {
		if p.Include(permission) {
			return true
		}
	}

	// deny by default
	return false
}

func (r *Role) Equal(another Role) bool {
	return r.Name == another.Name
}

func RoleFromProto(role pb.Role) (Role, error) {
	switch role {
	case pb.Role_ROLE_ADMIN:
		return AdminRole, nil
	case pb.Role_ROLE_MANAGER:
		return ManagerRole, nil
	case pb.Role_ROLE_USER:
		return UserRole, nil
	default:
		return Role{}, ErrUnknownRole
	}
}

var Roles = map[string]*Role{
	UserRole.Name:    &UserRole,
	ManagerRole.Name: &ManagerRole,
	AdminRole.Name:   &AdminRole,
}
