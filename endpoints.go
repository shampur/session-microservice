package session

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints exposed by the service
type Endpoints struct {
	loginEndpoint endpoint.Endpoint
	logoutEndpoint endpoint.Endpoint
	validateappEndpoint endpoint.Endpoint
	apiEndpoint endpoint.Endpoint
}

// MakeServerEndpoints function prepares the server Endpoints
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		loginEndpoint: MakeLoginEnpoint(s),
		logoutEndpoint: MakeLogoutEndpoint(s),
		validateappEndpoint: MakeValidateappEndpoint(s),
		apiEndpoint: MakeApiEndpoint(s),
	}
}

// MakeLoginEnpoint prepars the login endpoint
func MakeLoginEnpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(LoginRequest)
		result, err := s.login(ctx, req)
		return result, err
	}
}

func MakeLogoutEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(LogoutRequest)
		result, err := s.logout(ctx, req)
		return result, err
	}
}

func MakeValidateappEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(validateAppRequest)
		result, err := s.validateapp(ctx, req)
		return result, err
	}
}

func MakeApiEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(apiRequest)
		result, err := s.apiprocess(ctx, req)
		return result, err
	}
}
