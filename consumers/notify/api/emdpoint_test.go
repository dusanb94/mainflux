// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/consumers/notify"
	httpapi "github.com/mainflux/mainflux/consumers/notify/api"
	"github.com/mainflux/mainflux/consumers/notify/mocks"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

const (
	contentType = "application/json"
	email       = "user@example.com"
	token       = "token"
	wrongValue  = "wrong_value"
	wrongID     = 0
	maxNameSize = 1024
)

var (
	thing = things.Thing{
		Name:     "test_app",
		Metadata: map[string]interface{}{"test": "data"},
	}
	channel = things.Channel{
		Name:     "test",
		Metadata: map[string]interface{}{"test": "data"},
	}
	invalidName = strings.Repeat("m", maxNameSize+1)
	notFoundRes = toJSON(errorRes{things.ErrNotFound.Error()})
	unauthRes   = toJSON(errorRes{things.ErrUnauthorizedAccess.Error()})
)

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}
	return tr.client.Do(req)
}

func newService(tokens map[string]string) notify.Service {
	auth := mocks.NewAuth(tokens)
	repo := mocks.NewRepo(make(map[string]notify.Subscription))
	idp := uuid.NewMock()
	notif := mocks.NewNotifier()
	return notify.New(auth, repo, idp, notif)
}

func newServer(svc notify.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc, mocktracer.New())
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestCreate(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ss := newServer(svc)
	defer ss.Close()

	sub := notify.Subscription{
		Topic:   "topic",
		Contact: "contact@example.com",
	}

	data := toJSON(sub)

	emptyTopic := toJSON(notify.Subscription{Contact: "contact1@example.com"})
	emptyContact := toJSON(notify.Subscription{Topic: "topic123"})

	cases := []struct {
		desc        string
		req         string
		contentType string
		auth        string
		status      int
		location    string
	}{
		{
			desc:        "add a valid subscription",
			req:         data,
			contentType: contentType,
			auth:        token,
			status:      http.StatusCreated,
			location:    fmt.Sprintf("/subscriptions/%s%012d", uuid.Prefix, 1),
		},
		{
			desc:        "add an existing subscription",
			req:         data,
			contentType: contentType,
			auth:        token,
			status:      http.StatusConflict,
			location:    "",
		},
		{
			desc:        "add with empty topic",
			req:         emptyTopic,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
		{
			desc:        "add with empty contact",
			req:         emptyContact,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
		{
			desc:        "add thing with invalid auth token",
			req:         data,
			contentType: contentType,
			auth:        wrongValue,
			status:      http.StatusUnauthorized,
			location:    "",
		},
		{
			desc:        "add thing with empty auth token",
			req:         data,
			contentType: contentType,
			auth:        "",
			status:      http.StatusUnauthorized,
			location:    "",
		},
		{
			desc:        "add thing with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
		{
			desc:        "add thing without content type",
			req:         data,
			contentType: "",
			auth:        token,
			status:      http.StatusUnsupportedMediaType,
			location:    "",
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ss.Client(),
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/subscriptions", ss.URL),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))

		location := res.Header.Get("Location")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.location, location, fmt.Sprintf("%s: expected location %s got %s", tc.desc, tc.location, location))
	}
}

type errorRes struct {
	Err string `json:"error"`
}
