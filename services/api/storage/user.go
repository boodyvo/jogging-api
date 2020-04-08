package storage

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID     `json:"id" bson:"_id"`
	Email     string        `json:"email" bson:"email"`
	Password  string        `json:"-" bson:"password"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	Roles     []string      `json:"-" bson:"roles"`
	ACL       []Permission  `json:"-" bson:"acl"`
	Cursor    bson.ObjectId `json:"-" bson:"cursor"`
}

func NewUser(email, password string) *User {
	user := newUser(email, password)
	user.Roles = []string{UserRole.Name}
	// Adding basic permissions for particular user
	user.AddPermission(NewPermission(ReadAction, UserScope, user.ID.String()))
	user.AddPermission(NewPermission(UpdateAction, UserScope, user.ID.String()))
	user.AddPermission(NewPermission(DeleteAction, UserScope, user.ID.String()))

	return user
}

func NewManager(email, password string) *User {
	user := newUser(email, password)
	user.Roles = []string{ManagerRole.Name}

	return user
}

func NewAdmin(email, password string) *User {
	user := newUser(email, password)
	user.Roles = []string{AdminRole.Name}

	return user
}

func newUser(email, password string) *User {
	return &User{
		ID:        uuid.New(),
		Cursor:    bson.NewObjectId(),
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		ACL:       []Permission{},
	}
}

// TODO(boodyvo): for now just full search trough role but need to make roles by IDs
func (u *User) HasPermission(permission Permission) bool {
	// check in roles
	for _, role := range u.Roles {
		if Roles[role].HasPermission(permission) {
			return true
		}
	}
	// check in ACL
	for _, perm := range u.ACL {
		if perm.Include(permission) {
			return true
		}
	}

	return false
}

func (u *User) AddPermission(permission Permission) {
	u.ACL = append(u.ACL, permission)
}

func (u *User) RemovePermission(permission Permission) {
	newACL := make([]Permission, 0, len(u.ACL))
	for _, perm := range u.ACL {
		if !permission.Equal(perm) {
			newACL = append(newACL, perm)
		}
	}

	u.ACL = newACL
}

func (u *User) AddRole(role Role) {
	for _, r := range u.Roles {
		if role.Name == r {
			return
		}
	}

	u.Roles = append(u.Roles, role.Name)
}

func (u *User) RemoveRole(role Role) {
	roles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		if role.Name != r {
			roles = append(roles, r)
		}
	}

	u.Roles = roles
}

// TODO(boodyvo): Could check permissions by owning but for now in such way
// TODO(boodyvo): Implement permissions saving as separate table
func (u *User) AddTrackingPermission(id uuid.UUID) {
	// Adding basic permissions for particular user
	u.AddPermission(NewPermission(ReadAction, TrackingScope, id.String()))
	u.AddPermission(NewPermission(UpdateAction, TrackingScope, id.String()))
	u.AddPermission(NewPermission(DeleteAction, TrackingScope, id.String()))
}

func (u *User) RemoveTrackingPermission(id uuid.UUID) {
	// Adding basic permissions for particular user
	u.RemovePermission(NewPermission(ReadAction, TrackingScope, id.String()))
	u.RemovePermission(NewPermission(UpdateAction, TrackingScope, id.String()))
	u.RemovePermission(NewPermission(DeleteAction, TrackingScope, id.String()))
}

func (u *User) ToProto() *pb.User {
	return &pb.User{
		Id:    u.ID.String(),
		Email: u.Email,
	}
}

func (u *User) ToDetailedProto() *pb.DetailedUser {
	permissions := make([]string, 0, len(u.ACL))
	for _, acl := range u.ACL {
		permissions = append(permissions, acl.String())
	}
	return &pb.DetailedUser{
		Id:          u.ID.String(),
		Email:       u.Email,
		Roles:       u.Roles,
		Permissions: permissions,
	}
}

type UserFilter struct {
	PerRequest int64
	Cursor     string
	Query      string
}

func UserFilterFromProto(user *pb.ListUsersRequest) (*UserFilter, error) {
	return &UserFilter{
		Cursor:     user.Cursor,
		PerRequest: user.PerReq,
		Query:      user.Query,
	}, nil
}

type ListUsersResponse struct {
	Total int64
	Users []*User
}

func ProtoFromListUsersResponse(response *ListUsersResponse) *pb.ListUsersResponse {
	cursor := ""
	users := make([]*pb.User, 0, len(response.Users))
	for _, user := range response.Users {
		users = append(users, user.ToProto())
		cursor = user.Cursor.Hex()
	}
	return &pb.ListUsersResponse{
		Cursor: cursor,
		Total:  response.Total,
		Users:  users,
	}
}

func ProtoFromListUsersDetailedResponse(response *ListUsersResponse) *pb.ListUsersDetailedResponse {
	cursor := ""
	users := make([]*pb.DetailedUser, 0, len(response.Users))
	for _, user := range response.Users {
		users = append(users, user.ToDetailedProto())
		cursor = user.Cursor.Hex()
	}
	return &pb.ListUsersDetailedResponse{
		Cursor: cursor,
		Total:  response.Total,
		Users:  users,
	}
}
