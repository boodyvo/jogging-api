package api

import (
	"context"

	"github.com/boodyvo/jogging-api/services/api/weather"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	"github.com/boodyvo/jogging-api/services/api/auth"
	"github.com/boodyvo/jogging-api/services/api/storage"
)

type Server interface {
	Start() error
	Stop() error
	pb.APIServiceServer
}

type APIServer struct {
	auth    auth.Service
	store   storage.Storage
	logger  *log.Logger
	weather weather.Service

	wq   chan *storage.Tracking
	quit chan struct{}
}

func New(store storage.Storage, auth auth.Service, weather weather.Service, logger *log.Logger) Server {
	return &APIServer{
		auth:    auth,
		weather: weather,
		store:   store,
		logger:  logger,
		wq:      make(chan *storage.Tracking, 100),
		quit:    make(chan struct{}),
	}
}

func (s *APIServer) Start() error {
	for {
		select {
		case tracking := <-s.wq:
			weatherData, err := s.weather.GetWeather(tracking.Date, tracking.Location)
			if err != nil {
				log.
					WithField("err", err).
					WithField("tracking", tracking).
					Error("cannot get weather for tracking")

				continue
			}

			tracking.Weather = weatherData
			if err := s.store.UpdateTracking(tracking); err != nil {
				log.
					WithField("err", err).
					WithField("tracking", tracking).
					Error("cannot save weather for tracking")

				continue
			}
		case <-s.quit:
			return nil
		}
	}
}

func (s *APIServer) Stop() error {
	s.quit <- struct{}{}

	return nil
}

func (s *APIServer) CreateAdmin(_ context.Context, request *pb.CreateAdminRequest) (*pb.CreateAdminResponse, error) {
	s.logger.
		Info("Get create admin request")
	if err := request.Validate(); err != nil {
		return nil, ErrInvalidInputData
	}

	if !isValidEmail(request.Email) {
		return nil, ErrInvalidEmail
	}
	hashedPassword, err := hash(request.Password)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.GetUserByEmail(request.Email); err != storage.ErrNotFound {
		return nil, ErrUserAlreadyExists
	}
	user := storage.NewAdmin(request.Email, hashedPassword)
	if err := s.store.SaveUser(user); err != nil {
		return nil, err
	}

	return &pb.CreateAdminResponse{
		Id: user.ID.String(),
	}, nil
}

func (s *APIServer) AddPermission(ctx context.Context, request *pb.AddPermissionRequest) (*empty.Empty, error) {
	s.logger.
		WithField("request", request).
		Info("Get add permission request")
	if err := request.Validate(); err != nil {
		return nil, ErrInvalidInputData
	}

	err := s.checkPermission(ctx, storage.UpdateAction, storage.PermissionsScope, "*")
	if err != nil {
		return nil, err
	}
	permission, err := storage.PermissionFromProto(request.Action, request.Scope, request.Item)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user, err := s.store.GetUser(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user.AddPermission(permission)
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *APIServer) AddRole(ctx context.Context, request *pb.AddRoleRequest) (*empty.Empty, error) {
	s.logger.
		WithField("request", request).
		Info("Get add role request")
	if err := request.Validate(); err != nil {
		return nil, ErrInvalidInputData
	}

	err := s.checkPermission(ctx, storage.UpdateAction, storage.PermissionsScope, "*")
	if err != nil {
		return nil, err
	}
	role, err := storage.RoleFromProto(request.Role)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user, err := s.store.GetUser(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user.AddRole(role)
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *APIServer) RemoveRole(ctx context.Context, request *pb.RemoveRoleRequest) (*empty.Empty, error) {
	s.logger.
		WithField("request", request).
		Info("Get remove role request")
	if err := request.Validate(); err != nil {
		return nil, ErrInvalidInputData
	}

	err := s.checkPermission(ctx, storage.UpdateAction, storage.PermissionsScope, "*")
	if err != nil {
		return nil, err
	}
	role, err := storage.RoleFromProto(request.Role)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user, err := s.store.GetUser(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user.RemoveRole(role)
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *APIServer) SignUp(_ context.Context, request *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	s.logger.
		Info("Get sign up request")
	if err := request.Validate(); err != nil {
		return nil, ErrInvalidInputData
	}

	if !isValidEmail(request.Email) {
		return nil, ErrInvalidEmail
	}
	hashedPassword, err := hash(request.Password)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.GetUserByEmail(request.Email); err != storage.ErrNotFound {
		return nil, ErrUserAlreadyExists
	}
	user := storage.NewUser(request.Email, hashedPassword)
	if err := s.store.SaveUser(user); err != nil {
		return nil, err
	}

	return &pb.SignUpResponse{
		Id: user.ID.String(),
	}, nil
}

func (s *APIServer) SignIn(ctx context.Context, request *pb.SignInRequest) (*pb.SignInResponse, error) {
	s.logger.
		Info("Get sign in request")

	user, err := s.store.GetUserByEmail(request.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if !compare(user.Password, request.Password) {
		return nil, ErrUserNotFound
	}

	token, err := s.auth.GenerateToken(ctx, user)
	if err != nil {
		return nil, err
	}

	return &pb.SignInResponse{
		AccessToken:  token.Access,
		RefreshToken: token.Refresh,
		ExpireAt: &timestamp.Timestamp{
			Seconds: token.ExpiresAt.Unix(),
			Nanos:   int32(token.ExpiresAt.Nanosecond()),
		},
	}, nil
}

func (s *APIServer) GetUser(ctx context.Context, _ *empty.Empty) (*pb.GetUserResponse, error) {
	s.logger.
		Info("Get get user request")

	claims, err := s.auth.ParseAuthorizationHeader(ctx)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user, err := s.store.GetUser(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &pb.GetUserResponse{
		User: user.ToProto(),
	}, nil
}

func (s *APIServer) GetUserByID(ctx context.Context, request *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get get user by id request")

	err := s.checkPermission(ctx, storage.ReadAction, storage.UserScope, request.Id)
	if err != nil {
		return nil, err
	}

	idRes, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	userRes, err := s.store.GetUser(idRes)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &pb.GetUserResponse{
		User: userRes.ToProto(),
	}, nil
}

func (s *APIServer) ListUsers(ctx context.Context, request *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get list users request")

	err := s.checkPermission(ctx, storage.ReadAction, storage.UserScope, "*")
	if err != nil {
		return nil, err
	}

	filter, err := storage.UserFilterFromProto(request)
	if err != nil {
		return nil, ErrInvalidFilter
	}
	users, err := s.store.ListUsers(filter)
	if err != nil {
		s.logger.WithField("err", err).Error("error during list trackings")

		return nil, ErrInvalidFilter
	}
	response := storage.ProtoFromListUsersResponse(users)

	return response, nil
}

func (s *APIServer) ListUsersDetailed(ctx context.Context, request *pb.ListUsersRequest) (*pb.ListUsersDetailedResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get list detailed users request")

	err := s.checkPermission(ctx, storage.ReadAction, storage.UserScope, "*")
	if err != nil {
		return nil, err
	}

	filter, err := storage.UserFilterFromProto(request)
	if err != nil {
		return nil, ErrInvalidFilter
	}
	users, err := s.store.ListUsers(filter)
	if err != nil {
		s.logger.WithField("err", err).Error("error during list trackings")

		return nil, ErrInvalidFilter
	}
	response := storage.ProtoFromListUsersDetailedResponse(users)

	return response, nil
}

func (s *APIServer) DeleteUser(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	s.logger.
		Info("Get delete user request")

	claims, err := s.auth.ParseAuthorizationHeader(ctx)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if err := s.store.DeleteUser(id); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *APIServer) DeleteUserByID(ctx context.Context, request *pb.DeleteUserRequest) (*empty.Empty, error) {
	s.logger.
		WithField("request", request).
		Info("Get delete user by id request")

	err := s.checkPermission(ctx, storage.DeleteAction, storage.UserScope, request.Id)
	if err != nil {
		return nil, err
	}

	idRes, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if err := s.store.DeleteUser(idRes); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *APIServer) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	token, err := s.auth.RefreshToken(ctx, request.RefreshToken)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  token.Access,
		RefreshToken: token.Refresh,
		ExpireAt: &timestamp.Timestamp{
			Seconds: token.ExpiresAt.Unix(),
			Nanos:   int32(token.ExpiresAt.Nanosecond()),
		},
	}, nil
}

func (s *APIServer) CreateTracking(ctx context.Context, request *pb.CreateTrackingRequest) (*pb.CreateTrackingResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get create tracking request")

	if err := request.Validate(); err != nil {
		return nil, ErrInvalidInputData
	}

	user, err := s.checkAuthorization(ctx)
	if err != nil {
		return nil, err
	}

	tracking, err := storage.NewTrackingFromProto(request)
	if err != nil {
		return nil, ErrInvalidInputData
	}
	tracking.UserID = user.ID
	user.AddTrackingPermission(tracking.ID)

	// TODO(boodyvo): Not atomic. Could be as transaction.
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}
	if err := s.store.SaveTracking(tracking); err != nil {
		// TODO(boodyvo): Need to remove permissions for unsaved tracking
		return nil, err
	}
	err = s.setWeather(ctx, tracking)
	if err != nil {
		s.logger.
			WithField("err", err).
			WithField("tracking", tracking).
			Info("error while setting weather")
	}

	return &pb.CreateTrackingResponse{Id: tracking.ID.String()}, nil
}

func (s *APIServer) GetTracking(ctx context.Context, request *pb.GetTrackingRequest) (*pb.GetTrackingResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get get tracking request")

	err := s.checkPermission(ctx, storage.ReadAction, storage.TrackingScope, request.Id)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	tracking, err := s.store.GetTracking(id)
	if err != nil {
		return nil, err
	}

	return &pb.GetTrackingResponse{
		Tracking: tracking.ToProto(),
	}, nil
}

func (s *APIServer) DeleteTracking(ctx context.Context, request *pb.DeleteTrackingRequest) (*empty.Empty, error) {
	s.logger.
		WithField("request", request).
		Info("Get delete tracking request")

	err := s.checkPermission(ctx, storage.DeleteAction, storage.TrackingScope, request.Id)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, ErrTrackingNotFound
	}
	tracking, err := s.store.GetTracking(id)
	if err != nil {
		return nil, err
	}
	user, err := s.store.GetUser(tracking.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user.AddTrackingPermission(tracking.ID)

	if err := s.store.DeleteTracking(tracking.ID); err != nil {
		return nil, err
	}
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *APIServer) ListTrackingsForUser(ctx context.Context, request *pb.ListTrackingsRequest) (*pb.ListTrackingsResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get list trackings request")

	user, err := s.checkAuthorization(ctx)
	if err != nil {
		return nil, err
	}

	filter, err := storage.TrackingFilterFromProtoForUser(request, user)
	if err != nil {
		return nil, ErrInvalidFilter
	}
	trackings, err := s.store.ListTrackingsForUser(filter)
	if err != nil {
		s.logger.WithField("err", err).Error("error during list trackings")

		return nil, ErrInvalidFilter
	}
	response := storage.ProtoFromListTrackingsResponse(trackings)

	return response, nil
}

func (s *APIServer) ListTrackings(ctx context.Context, request *pb.ListTrackingsRequest) (*pb.ListTrackingsResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get list all trackings request")

	err := s.checkPermission(ctx, storage.ReadAction, storage.TrackingScope, "*")
	if err != nil {
		return nil, err
	}

	filter, err := storage.TrackingFilterFromProto(request)
	if err != nil {
		return nil, ErrInvalidFilter
	}
	trackings, err := s.store.ListTrackings(filter)
	if err != nil {
		s.logger.WithField("err", err).Error("error during list trackings")

		return nil, ErrInvalidFilter
	}
	response := storage.ProtoFromListTrackingsResponse(trackings)

	return response, nil
}

func (s *APIServer) Report(ctx context.Context, request *pb.ReportRequest) (*pb.ReportResponse, error) {
	s.logger.
		WithField("request", request).
		Info("Get report request")

	user, err := s.checkAuthorization(ctx)
	if err != nil {
		return nil, err
	}

	filter, err := storage.NewReportFilterFromProtoForUser(request, user)
	if err != nil {
		return nil, ErrInvalidFilter
	}
	report, err := s.store.GetReport(filter)
	if err != nil {
		s.logger.WithField("err", err).Error("error during getting report")

		return nil, ErrInvalidFilter
	}

	return report.ToProto(), nil
}

// TODO(boodyvo): Implement message queue
func (s *APIServer) setWeather(_ context.Context, tracking *storage.Tracking) error {
	s.wq <- tracking

	return nil
}

func (s *APIServer) checkPermission(ctx context.Context, action storage.Action, scope storage.Scope, item string) error {
	user, err := s.checkAuthorization(ctx)
	if err != nil {
		return err
	}

	permission := storage.NewPermission(action, scope, item)
	if !user.HasPermission(permission) {
		return ErrForbidden
	}

	return nil
}

func (s *APIServer) checkAuthorization(ctx context.Context) (*storage.User, error) {
	claims, err := s.auth.ParseAuthorizationHeader(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	user, err := s.store.GetUser(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}
