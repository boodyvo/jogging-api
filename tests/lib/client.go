package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/boodyvo/jogging-api/lib"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
	"github.com/boodyvo/jogging-api/tests/common"
)

type client struct {
	url    string
	client *http.Client
}

func NewClient(url string) HttpClient {
	rand.Seed(time.Now().UTC().UnixNano())

	return &client{
		url:    url,
		client: &http.Client{},
	}
}

func (c *client) SignUp(request *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	buf, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/v1/signup", c.url),
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"wrong status code: expected 200, got %v, details: %v",
			resp.StatusCode,
			string(body),
		)
	}

	var result pb.SignUpResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) SignIn(request *pb.SignInRequest) (*pb.SignInResponse, error) {
	buf, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/v1/signin", c.url),
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	var result pb.SignInResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) AddPermission(user *User, request *pb.AddPermissionRequest) (*empty.Empty, error) {
	buf, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/api/v1/user/permissions", c.url),
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return &empty.Empty{}, nil
}

func (c *client) AddRole(user *User, request *pb.AddRoleRequest) (*empty.Empty, error) {
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/api/v1/user/%s/roles/%s", c.url, request.UserId, request.Role),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return &empty.Empty{}, nil
}

func (c *client) RemoveRole(user *User, request *pb.RemoveRoleRequest) (*empty.Empty, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/api/v1/user/%s/roles/%s", c.url, request.UserId, request.Role),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return &empty.Empty{}, nil
}

func (c *client) GetUser(user *User, _ *empty.Empty) (*pb.GetUserResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/user", c.url),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	var result pb.GetUserResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetUserByID(user *User, request *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/user/%s", c.url, request.Id),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	var result pb.GetUserResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) DeleteUser(user *User, _ *empty.Empty) (*empty.Empty, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/api/v1/user", c.url),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return &empty.Empty{}, nil
}

func (c *client) DeleteUserByID(user *User, request *pb.DeleteUserRequest) (*empty.Empty, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/api/v1/user/%s", c.url, request.Id),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return &empty.Empty{}, nil
}

func (c *client) ListUsers(user *User, request *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/users", c.url),
		nil,
	)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if request.Cursor != "" {
		q.Add("cursor", request.Cursor)
	}
	if request.Query != "" {
		q.Add("query", request.Query)
	}
	if request.PerReq != 0 {
		q.Add("per_req", strconv.FormatInt(request.PerReq, 10))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}

	var result pb.ListUsersResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) CreateTracking(user *User, request *CreateTrackingRequest) (*pb.CreateTrackingResponse, error) {
	buf, err := json.Marshal(createTrackingRequestSerialized{
		Date:     request.Date.Format(lib.DateFormat),
		Time:     request.Time,
		Distance: request.Distance,
		Location: request.Location,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/v1/tracking", c.url),
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("wrong status code: %d, details: %s", resp.StatusCode, string(body))
	}

	var result pb.CreateTrackingResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetTracking(user *User, request *pb.GetTrackingRequest) (*pb.GetTrackingResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/tracking/%s", c.url, request.Id),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	var result pb.GetTrackingResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) ListOwnTrackings(user *User, request *pb.ListTrackingsRequest) (*pb.ListTrackingsResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/trackings", c.url),
		nil,
	)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if request.Cursor != "" {
		q.Add("cursor", request.Cursor)
	}
	if request.Query != "" {
		q.Add("query", request.Query)
	}
	if request.PerReq != 0 {
		q.Add("per_req", strconv.FormatInt(request.PerReq, 10))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	var result pb.ListTrackingsResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) ListTrackings(user *User, request *pb.ListTrackingsRequest) (*pb.ListTrackingsResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/trackings/all", c.url),
		nil,
	)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if request.Cursor != "" {
		q.Add("cursor", request.Cursor)
	}
	if request.Query != "" {
		q.Add("query", request.Query)
	}
	if request.PerReq != 0 {
		q.Add("per_req", strconv.FormatInt(request.PerReq, 10))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	var result pb.ListTrackingsResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) Report(user *User, request *ReportRequest) (*pb.ReportResponse, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/api/v1/trackings/report", c.url),
		nil,
	)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if request.Duration != "" {
		q.Add("duration", request.Duration)
	}
	if !request.FromDate.IsZero() {
		q.Add("from_date", request.FromDate.Format(lib.DateFormat))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	var result pb.ReportResponse

	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) CreateRandomTracking(user *User) (*Tracking, error) {
	tracking := Tracking{
		UserID:   user.ID,
		Location: CreateLocation(),
		Date:     time.Now().AddDate(0, -3, -5),
		Time:     time.Duration(int64(time.Second) * rand.Int63n(1000)).String(),
		Distance: float32(rand.Int31n(100000)) / 100,
	}
	createTrackingRequest := &CreateTrackingRequest{
		Location: tracking.Location,
		Time:     tracking.Time,
		Distance: tracking.Distance,
		Date:     tracking.Date,
	}
	resp, err := c.CreateTracking(user, createTrackingRequest)
	if err != nil {
		return nil, err
	}
	tracking.ID = resp.Id

	return &tracking, nil
}

func (c *client) CreateRandomAuthorizedUser() (*User, error) {
	user := &User{
		Email:    CreateEmail(),
		Password: common.DefaultPassword,
	}
	signUpResp, err := c.SignUp(&pb.SignUpRequest{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return nil, err
	}
	user.ID = signUpResp.Id

	signInResp, err := c.SignIn(&pb.SignInRequest{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return nil, err
	}

	user.RefreshToken = signInResp.RefreshToken
	user.AccessToken = signInResp.AccessToken

	return user, nil
}
