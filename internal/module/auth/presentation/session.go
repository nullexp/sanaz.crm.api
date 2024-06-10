package presentation

import (
	"context"
	"net/http"

	httpapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model/openapi"
	application "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application"
	request "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/request"
	response "github.com/nullexp/sanaz.crm.api/pkg/module/auth/application/dto/response"
)

const SessionBaseURL = "/sessions"

func NewSession(session application.Session) httpapi.Module {
	return Session{sessionService: session}
}

type Session struct {
	sessionService application.Session
}

func (s Session) GetRequestHandlers() []*httpapi.RequestDefinition {
	return []*httpapi.RequestDefinition{
		s.PostSession(), s.PostRefreshToken(),
	}
}

func (s Session) GetBaseURL() string {
	return SessionBaseURL
}

const (
	SessionManagement  = "Session Management"
	SessionDescription = "Authenticate through these apis"
)

func (s Session) GetTag() openapi.Tag {
	return openapi.Tag{
		Name:        SessionManagement,
		Description: SessionDescription,
	}
}

const RouteRefresh = "/refresh"

func (s Session) PostSession() *httpapi.RequestDefinition {
	return &httpapi.RequestDefinition{
		Route:     "",
		Dto:       &request.Session{},
		FreeRoute: true,
		Method:    http.MethodPost,
		ResponseDefinitions: []httpapi.ResponseDefinition{
			{
				Status:      http.StatusCreated,
				Description: "If auth info is valid",
				Dto:         &response.Token{},
			},
			{
				Status:      http.StatusBadRequest,
				Description: "If auth info is not valid",
			},
		},
		Handler: func(req httpapi.Request) {
			dto := req.MustGetDTO().(*request.Session)
			token, err := s.sessionService.Authenticate(context.Background(), *dto)
			req.Negotiate(http.StatusCreated, err, token)
		},
	}
}

func (s Session) PostRefreshToken() *httpapi.RequestDefinition {
	return &httpapi.RequestDefinition{
		Route:     RouteRefresh,
		FreeRoute: true,
		Dto:       &request.RefreshSession{},
		Method:    http.MethodPost,
		ResponseDefinitions: []httpapi.ResponseDefinition{
			{
				Status:      http.StatusCreated,
				Description: "If auth info is valid",
				Dto:         &response.AccessToken{},
			},
			{
				Status:      http.StatusBadRequest,
				Description: "If auth info is not valid",
			},
		},
		Handler: func(req httpapi.Request) {
			dto := req.MustGetDTO().(*request.RefreshSession)
			token, err := s.sessionService.RefreshToken(context.Background(), *dto)
			req.Negotiate(http.StatusCreated, err, token)
		},
	}
}
