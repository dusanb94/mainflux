// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/consumers/notify"
	httpapi "github.com/mainflux/mainflux/consumers/notify/api"
	"github.com/mainflux/mainflux/consumers/notify/mocks"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	contentType = "application/json"
	email       = "user@example.com"
	token       = "token"
	wrongValue  = "wrong_value"
)

var (
	notFoundRes = toJSON(errorRes{notify.ErrNotFound.Error()})
	unauthRes   = toJSON(errorRes{notify.ErrUnauthorizedAccess.Error()})
	invalidRes  = toJSON(errorRes{"invalid query parameters"})
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
			desc:        "add successfully",
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
			desc:        "add with invalid auth token",
			req:         data,
			contentType: contentType,
			auth:        wrongValue,
			status:      http.StatusUnauthorized,
			location:    "",
		},
		{
			desc:        "add with empty auth token",
			req:         data,
			contentType: contentType,
			auth:        "",
			status:      http.StatusUnauthorized,
			location:    "",
		},
		{
			desc:        "add with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
		{
			desc:        "add without content type",
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

func TestView(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ss := newServer(svc)
	defer ss.Close()

	sub := notify.Subscription{
		Topic:   "topic",
		Contact: "contact@example.com",
	}
	id, err := svc.CreateSubscription(context.Background(), token, sub)
	require.Nil(t, err, fmt.Sprintf("got an error creating id: %s", err))
	sr := subRes{
		ID:      id,
		OwnerID: email,
		Contact: sub.Contact,
		Topic:   sub.Topic,
	}
	data := toJSON(sr)

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
		res    string
	}{
		{
			desc:   "view successfully",
			id:     id,
			auth:   token,
			status: http.StatusOK,
			res:    data,
		},
		{
			desc:   "view not existing",
			id:     "not existing",
			auth:   token,
			status: http.StatusNotFound,
			res:    notFoundRes,
		},
		{
			desc:   "view with invalid auth token",
			id:     id,
			auth:   wrongValue,
			status: http.StatusUnauthorized,
			res:    unauthRes,
		},
		{
			desc:   "view with empty auth token",
			id:     id,
			auth:   "",
			status: http.StatusUnauthorized,
			res:    unauthRes,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: ss.Client(),
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/subscriptions/%s", ss.URL, tc.id),
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		data := strings.Trim(string(body), "\n")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, data, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, data))
	}
}

func TestList(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ss := newServer(svc)
	defer ss.Close()

	const numSubs = 100
	var subs []subRes

	for i := 0; i < numSubs; i++ {
		sub := notify.Subscription{
			Topic:   fmt.Sprintf("topic.subtopic.%d", i),
			Contact: "contact1@example.com",
		}
		if i%2 == 0 {
			sub.Contact = "contact2@example.com"
		}
		id, err := svc.CreateSubscription(context.Background(), token, sub)
		sr := subRes{
			ID:      id,
			OwnerID: email,
			Contact: sub.Contact,
			Topic:   sub.Topic,
		}
		require.Nil(t, err, fmt.Sprintf("got an error creating id: %s", err))
		subs = append(subs, sr)
	}
	noLimit := toJSON(page{Offset: 5, Limit: 20, Total: 100, Subscriptions: subs[5:25]})
	one := toJSON(page{Offset: 0, Limit: 20, Total: 1, Subscriptions: subs[10:11]})

	var contact2Subs []subRes
	for i := 20; i < 40; i += 2 {
		contact2Subs = append(contact2Subs, subs[i])
	}
	contact2 := toJSON(page{Offset: 10, Limit: 10, Total: 50, Subscriptions: contact2Subs})

	cases := map[string]struct {
		query  map[string]string
		auth   string
		status int
		res    string
	}{
		"list default limit": {
			query: map[string]string{
				"offset": "5",
			},
			auth:   token,
			status: http.StatusOK,
			res:    noLimit,
		},
		"list not existing": {
			query: map[string]string{
				"topic": "not-found-topic",
			},
			auth:   token,
			status: http.StatusNotFound,
			res:    notFoundRes,
		},
		"list one with topic": {
			query: map[string]string{
				"topic": "topic.subtopic.10",
			},
			auth:   token,
			status: http.StatusOK,
			res:    one,
		},
		"list with contact": {
			query: map[string]string{
				"contact": "contact2@example.com",
				"offset":  "10",
				"limit":   "10",
			},
			auth:   token,
			status: http.StatusOK,
			res:    contact2,
		},
		"list with invalid query": {
			query: map[string]string{
				"offset": "two",
			},
			auth:   token,
			status: http.StatusBadRequest,
			res:    invalidRes,
		},
		"list with invalid auth token": {
			auth:   wrongValue,
			status: http.StatusUnauthorized,
			res:    unauthRes,
		},
		"list with empty auth token": {
			auth:   "",
			status: http.StatusUnauthorized,
			res:    unauthRes,
		},
	}

	for desc, tc := range cases {
		req := testRequest{
			client: ss.Client(),
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/subscriptions%s", ss.URL, makeQuery(tc.query)),
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		data := strings.Trim(string(body), "\n")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, data, fmt.Sprintf("%s: expected body %s got %s", desc, tc.res, data))
	}
}

func TestRemove(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ss := newServer(svc)
	defer ss.Close()

	sub := notify.Subscription{
		Topic:   "topic",
		Contact: "contact@example.com",
	}
	id, err := svc.CreateSubscription(context.Background(), token, sub)
	require.Nil(t, err, fmt.Sprintf("got an error creating id: %s", err))

	cases := map[string]struct {
		desc   string
		id     string
		auth   string
		status int
		res    string
	}{
		"remove successfully": {
			id:     id,
			auth:   token,
			status: http.StatusNoContent,
		},
		"remove not existing": {
			id:     "not existing",
			auth:   token,
			status: http.StatusNoContent,
		},
		"remove empty id": {
			id:     "",
			auth:   token,
			status: http.StatusNotFound,
		},
		"view with invalid auth token": {
			id:     id,
			auth:   wrongValue,
			status: http.StatusUnauthorized,
			res:    unauthRes,
		},
		"view with empty auth token": {
			id:     id,
			auth:   "",
			status: http.StatusUnauthorized,
			res:    unauthRes,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: ss.Client(),
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/subscriptions/%s", ss.URL, tc.id),
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func makeQuery(m map[string]string) string {
	var ret string
	for k, v := range m {
		ret += fmt.Sprintf("&%s=%s", k, v)
	}
	if ret != "" {
		return fmt.Sprintf("?%s", ret[1:])
	}
	return ""
}

type errorRes struct {
	Err string `json:"error"`
}

type subRes struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
	Contact string `json:"contact"`
	Topic   string `json:"topic"`
}
type page struct {
	Offset        uint     `json:"offset"`
	Limit         int      `json:"limit"`
	Total         uint     `json:"total,omitempty"`
	Subscriptions []subRes `json:"subscriptions,omitempty"`
}
