// +build integration

package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/require"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	pbclient "github.com/boodyvo/jogging-api/services/api/client"
	"github.com/boodyvo/jogging-api/tests/common"
	"github.com/boodyvo/jogging-api/tests/lib"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestCreatingUsersWithInvalidData(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	// first user

	_, err := client.SignUp(&pb.SignUpRequest{
		Email:    lib.CreateEmail(),
		Password: "",
	})
	r.Error(err, "can create user with empty password")

	_, err = client.SignUp(&pb.SignUpRequest{
		Email:    "",
		Password: common.DefaultPassword,
	})
	r.Error(err, "can create user with empty email")

	_, err = client.SignUp(&pb.SignUpRequest{
		Email:    fmt.Sprintf("%s@A", lib.CreateName()),
		Password: common.DefaultPassword,
	})
	r.Error(err, "can create user with invalid email")

	_, err = client.SignUp(&pb.SignUpRequest{
		Email:    lib.CreateEmail(),
		Password: common.DefaultPassword,
	})
	r.NoError(err, "cannot create user with appropriate credentials")
}

func TestUserBasicFlow(t *testing.T) {
	r := require.New(t)
	client := lib.NewClient(common.DefaultURL)

	// first user

	userFirst, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create first user")

	userResp, err := client.GetUser(userFirst, &empty.Empty{})
	r.NoError(err, "cannot get first user by first user")
	r.Equal(userFirst.ID, userResp.User.Id, "ID for get user request is not valid for first user")
	r.Equal(userFirst.Email, userResp.User.Email, "email for get user request is not valid for first user")

	userResp, err = client.GetUserByID(userFirst, &pb.GetUserRequest{Id: userFirst.ID})
	r.NoError(err, "cannot get first user by id by first user")
	r.Equal(userFirst.ID, userResp.User.Id, "ID for get user request by id is not valid for first user")
	r.Equal(userFirst.Email, userResp.User.Email, "email for get user request by id is not valid for first user")

	// second user

	userSecond, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create second user")

	userResp, err = client.GetUser(userSecond, &empty.Empty{})
	r.NoError(err, "cannot get second user by second user")
	r.Equal(userSecond.ID, userResp.User.Id, "ID for get user request is not valid for second user")
	r.Equal(userSecond.Email, userResp.User.Email, "email for get user request is not valid for second user")

	userResp, err = client.GetUserByID(userSecond, &pb.GetUserRequest{Id: userSecond.ID})
	r.NoError(err, "cannot get second user by id by second user")
	r.Equal(userSecond.ID, userResp.User.Id, "ID for get user request by id is not valid for second user")
	r.Equal(userSecond.Email, userResp.User.Email, "email for get user request by id is not valid for second user")

	// Permission check

	userResp, err = client.GetUserByID(userFirst, &pb.GetUserRequest{Id: userSecond.ID})
	r.Error(err, "can get second user by first user")

	userResp, err = client.GetUserByID(userSecond, &pb.GetUserRequest{Id: userFirst.ID})
	r.Error(err, "can get first user by second user")

	_, err = client.DeleteUserByID(userFirst, &pb.DeleteUserRequest{Id: userSecond.ID})
	r.Error(err, "can delete second user by first user")

	_, err = client.DeleteUserByID(userSecond, &pb.DeleteUserRequest{Id: userFirst.ID})
	r.Error(err, "can delete first user by second user")

	// Delete check

	_, err = client.DeleteUserByID(userFirst, &pb.DeleteUserRequest{Id: userFirst.ID})
	r.NoError(err, "cannot delete first user by first user")

	_, err = client.DeleteUserByID(userSecond, &pb.DeleteUserRequest{Id: userSecond.ID})
	r.NoError(err, "cannot delete second user by second user")

	_, err = client.SignIn(&pb.SignInRequest{
		Email:    userFirst.Email,
		Password: userFirst.Password,
	})
	r.Error(err, "can sign in first user after delete")

	_, err = client.SignIn(&pb.SignInRequest{
		Email:    userSecond.Email,
		Password: userSecond.Password,
	})
	r.Error(err, "can sign in second user after delete")

	_, err = client.GetUser(userFirst, &empty.Empty{})
	r.Error(err, "can get first user by second user after delete")

	_, err = client.GetUser(userSecond, &empty.Empty{})
	r.Error(err, "can get second user by second user after delete")
}

func TestAdminUserFlow(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	client := lib.NewClient(common.DefaultURL)
	grpcClient, err := pbclient.New(ctx, common.DefaultGRPCURL)
	r.NoError(err, "cannot create grpc client")

	// create admin user
	adminUser := &lib.User{
		Email:    lib.CreateEmail(),
		Password: common.DefaultPassword,
	}
	_, err = grpcClient.CreateAdmin(ctx, &pb.CreateAdminRequest{
		Email:    adminUser.Email,
		Password: adminUser.Password,
	})
	r.NoError(err, "cannot create admin user")

	signInResp, err := client.SignIn(&pb.SignInRequest{
		Email:    adminUser.Email,
		Password: adminUser.Password,
	})
	r.NoError(err, "cannot sign in admin user")
	adminUser.RefreshToken = signInResp.RefreshToken
	adminUser.AccessToken = signInResp.AccessToken

	// create common user

	commonUser, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create common user")

	// validate different permissions

	_, err = client.CreateRandomTracking(commonUser)
	r.NoError(err, "cannot create tracking for common user")

	// list trackings

	_, err = client.ListTrackings(commonUser, &pb.ListTrackingsRequest{})
	r.Error(err, "common user can list all trackings")

	listOwnTrackings, err := client.ListOwnTrackings(commonUser, &pb.ListTrackingsRequest{})
	r.NoError(err, "common user cannot list own trackings")
	r.Equal(int64(1), listOwnTrackings.Total, "cannot get common user's tracking by common user")

	listTrackingResponse, err := client.ListTrackings(adminUser, &pb.ListTrackingsRequest{})
	r.NoError(err, "admin user cannot list trackings")
	r.Less(int64(0), listTrackingResponse.Total, "cannot get other user's tracking by admin user")

	// list users

	listUsersResponse, err := client.ListUsers(commonUser, &pb.ListUsersRequest{})
	r.Error(err, "common user can list users")

	listUsersResponse, err = client.ListUsers(adminUser, &pb.ListUsersRequest{})
	r.NoError(err, "admin user cannot list users")
	r.Less(int64(1), listUsersResponse.Total, "cannot get other users by admin users")

	// create common user

	secondCommonUser, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create common user")

	// check deletion of user

	_, err = client.DeleteUserByID(commonUser, &pb.DeleteUserRequest{
		Id: secondCommonUser.ID,
	})
	r.Error(err, "common user can delete another user")

	// add roles/permissions

	_, err = client.AddPermission(commonUser, &pb.AddPermissionRequest{
		UserId: adminUser.ID,
		Scope:  pb.Scope_SCOPE_TRACKINGS,
		Action: pb.Action_ACTION_READ,
		Item:   "*",
	})
	r.Error(err, "common user can add permissions the users")

	_, err = client.AddRole(commonUser, &pb.AddRoleRequest{
		UserId: commonUser.ID,
		Role:   pb.Role_ROLE_MANAGER,
	})
	r.Error(err, "common user can add roles the users")

	_, err = client.AddPermission(adminUser, &pb.AddPermissionRequest{
		UserId: commonUser.ID,
		Scope:  pb.Scope_SCOPE_TRACKINGS,
		Action: pb.Action_ACTION_READ,
		Item:   "*",
	})
	r.NoError(err, "admin user cannot add permissions the user")
	_, err = client.AddRole(adminUser, &pb.AddRoleRequest{
		UserId: commonUser.ID,
		Role:   pb.Role_ROLE_MANAGER,
	})
	r.NoError(err, "admin user cannot add roles the user")

	// check permissions

	_, err = client.ListTrackings(commonUser, &pb.ListTrackingsRequest{})
	r.NoError(err, "common user cannot list all trackings after getting permissions")

	_, err = client.ListUsers(commonUser, &pb.ListUsersRequest{})
	r.NoError(err, "common user cannot list users after getting manager role")

	_, err = client.DeleteUserByID(commonUser, &pb.DeleteUserRequest{
		Id: secondCommonUser.ID,
	})
	r.NoError(err, "manager user cannot delete another user")

	// delete users

	_, err = client.DeleteUserByID(adminUser, &pb.DeleteUserRequest{
		Id: commonUser.ID,
	})
	r.NoError(err, "admin user could not delete the common user")
}

func TestAdminUserCreateAdmin(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	client := lib.NewClient(common.DefaultURL)
	grpcClient, err := pbclient.New(ctx, common.DefaultGRPCURL)
	r.NoError(err, "cannot create grpc client")

	// create admin user
	adminUser := &lib.User{
		Email:    lib.CreateEmail(),
		Password: common.DefaultPassword,
	}
	_, err = grpcClient.CreateAdmin(ctx, &pb.CreateAdminRequest{
		Email:    adminUser.Email,
		Password: adminUser.Password,
	})
	r.NoError(err, "cannot create admin user")

	signInResp, err := client.SignIn(&pb.SignInRequest{
		Email:    adminUser.Email,
		Password: adminUser.Password,
	})
	r.NoError(err, "cannot sign in admin user")
	adminUser.RefreshToken = signInResp.RefreshToken
	adminUser.AccessToken = signInResp.AccessToken

	// create common user

	commonUser, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create common user")

	// add role

	_, err = client.AddRole(adminUser, &pb.AddRoleRequest{
		UserId: commonUser.ID,
		Role:   pb.Role_ROLE_ADMIN,
	})
	r.NoError(err, "admin user cannot add roles")

	// create common user

	secondCommonUser, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create second common user")

	// add role

	_, err = client.AddRole(commonUser, &pb.AddRoleRequest{
		UserId: secondCommonUser.ID,
		Role:   pb.Role_ROLE_ADMIN,
	})
	r.NoError(err, "common user cannot add roles after got admin role")

	// remove role

	_, err = client.RemoveRole(secondCommonUser, &pb.RemoveRoleRequest{
		UserId: commonUser.ID,
		Role:   pb.Role_ROLE_ADMIN,
	})
	r.NoError(err, "second user cannot remove roles")

	// add role

	_, err = client.AddRole(commonUser, &pb.AddRoleRequest{
		UserId: secondCommonUser.ID,
		Role:   pb.Role_ROLE_MANAGER,
	})
	r.Error(err, "common user can add roles after removing admin permissions")

	// remove role

	_, err = client.RemoveRole(adminUser, &pb.RemoveRoleRequest{
		UserId: secondCommonUser.ID,
		Role:   pb.Role_ROLE_ADMIN,
	})
	r.NoError(err, "admin user cannot add roles the users")

	// add role

	_, err = client.AddRole(secondCommonUser, &pb.AddRoleRequest{
		UserId: commonUser.ID,
		Role:   pb.Role_ROLE_ADMIN,
	})
	r.Error(err, "second common user can add roles after removing admin permissions")
}

func TestAdminUserListQuery(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	client := lib.NewClient(common.DefaultURL)
	grpcClient, err := pbclient.New(ctx, common.DefaultGRPCURL)
	r.NoError(err, "cannot create grpc client")

	// create admin user
	adminUser := &lib.User{
		Email:    lib.CreateEmail(),
		Password: common.DefaultPassword,
	}
	_, err = grpcClient.CreateAdmin(ctx, &pb.CreateAdminRequest{
		Email:    adminUser.Email,
		Password: adminUser.Password,
	})
	r.NoError(err, "cannot create admin user")

	signInResp, err := client.SignIn(&pb.SignInRequest{
		Email:    adminUser.Email,
		Password: adminUser.Password,
	})
	r.NoError(err, "cannot sign in admin user")
	adminUser.RefreshToken = signInResp.RefreshToken
	adminUser.AccessToken = signInResp.AccessToken

	// create common user

	commonUser, err := client.CreateRandomAuthorizedUser()
	r.NoError(err, "cannot create common user")

	// list

	listUsersResponse, err := client.ListUsers(adminUser, &pb.ListUsersRequest{})
	r.NoError(err, "admin user cannot list users")
	r.Less(int64(1), listUsersResponse.Total, "cannot get other users by admin users")

	listUsersResponse, err = client.ListUsers(adminUser, &pb.ListUsersRequest{
		Query: fmt.Sprintf("email eq %s", commonUser.Email),
	})
	r.NoError(err, "admin user cannot list users")
	r.Equal(int64(1), listUsersResponse.Total, "cannot get users by query")
	r.Equal(commonUser.ID, listUsersResponse.Users[0].Id, "wrong query response")
	r.Equal(commonUser.Email, listUsersResponse.Users[0].Email, "wrong query response")

}
