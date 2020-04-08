package lib

import (
	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	"github.com/golang/protobuf/ptypes/empty"
)

type HttpClient interface {
	// auth
	SignUp(request *pb.SignUpRequest) (*pb.SignUpResponse, error)
	SignIn(request *pb.SignInRequest) (*pb.SignInResponse, error)

	// managing
	AddPermission(user *User, request *pb.AddPermissionRequest) (*empty.Empty, error)
	AddRole(user *User, request *pb.AddRoleRequest) (*empty.Empty, error)
	RemoveRole(user *User, request *pb.RemoveRoleRequest) (*empty.Empty, error)

	// users
	GetUser(user *User, _ *empty.Empty) (*pb.GetUserResponse, error)
	GetUserByID(user *User, request *pb.GetUserRequest) (*pb.GetUserResponse, error)
	DeleteUser(user *User, _ *empty.Empty) (*empty.Empty, error)
	DeleteUserByID(user *User, request *pb.DeleteUserRequest) (*empty.Empty, error)
	ListUsers(user *User, request *pb.ListUsersRequest) (*pb.ListUsersResponse, error)

	// trackings
	CreateTracking(user *User, request *CreateTrackingRequest) (*pb.CreateTrackingResponse, error)
	GetTracking(user *User, request *pb.GetTrackingRequest) (*pb.GetTrackingResponse, error)
	ListOwnTrackings(user *User, request *pb.ListTrackingsRequest) (*pb.ListTrackingsResponse, error)
	ListTrackings(user *User, request *pb.ListTrackingsRequest) (*pb.ListTrackingsResponse, error)
	Report(user *User, request *ReportRequest) (*pb.ReportResponse, error)

	// util methods
	CreateRandomTracking(user *User) (*Tracking, error)
	CreateRandomAuthorizedUser() (*User, error)
}
