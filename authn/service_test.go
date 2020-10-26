// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/authn"
	"github.com/mainflux/mainflux/authn/jwt"
	"github.com/mainflux/mainflux/authn/mocks"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	secret = "secret"
	email  = "test@example.com"
)

func newService() authn.Service {
	repo := mocks.NewKeyRepository()
	uuidProvider := uuid.NewMock()
	t := jwt.New(secret)
	return authn.New(repo, uuidProvider, t)
}

func TestIssue(t *testing.T) {
	svc := newService()
	userKey, secret, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	cases := []struct {
		desc   string
		key    authn.Key
		issuer string
		err    error
	}{
		{
			desc: "issue user key",
			key: authn.Key{
				Type:     authn.UserKey,
				IssuedAt: time.Now(),
			},
			issuer: email,
			err:    nil,
		},
		{
			desc: "issue user key with no time",
			key: authn.Key{
				Type: authn.UserKey,
			},
			issuer: email,
			err:    authn.ErrInvalidKeyIssuedAt,
		},
		{
			desc: "issue API key",
			key: authn.Key{
				Type:     authn.APIKey,
				IssuedAt: time.Now(),
			},
			issuer: email,
			err:    nil,
		},
		{
			desc: "issue API key unauthorized",
			key: authn.Key{
				Type:     authn.APIKey,
				IssuedAt: time.Now(),
			},
			issuer: "",
			err:    authn.ErrUnauthorizedAccess,
		},
		{
			desc: "issue API key with no time",
			key: authn.Key{
				Type: authn.APIKey,
			},
			issuer: email,
			err:    authn.ErrInvalidKeyIssuedAt,
		},
		{
			desc: "issue recovery key",
			key: authn.Key{
				Type:     authn.RecoveryKey,
				IssuedAt: time.Now(),
			},
			issuer: email,
			err:    nil,
		},
		{
			desc: "issue recovery with no issue time",
			key: authn.Key{
				Type: authn.RecoveryKey,
			},
			issuer: userKey.Subject,
			err:    authn.ErrInvalidKeyIssuedAt,
		},
	}

	for _, tc := range cases {
		_, _, err := svc.Issue(context.Background(), tc.issuer, tc.issuer, tc.key)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
	}
}
func TestRevoke(t *testing.T) {
	svc := newService()
	loginKey, secret, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := authn.Key{
		Type:     authn.APIKey,
		IssuedAt: time.Now(),
	}
	newKey, _, err := svc.Issue(context.Background(), "", loginKey.Email, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing user's key expected to succeed: %s", err))

	cases := []struct {
		desc   string
		id     string
		issuer string
		err    error
	}{
		{
			desc:   "revoke user key",
			id:     newKey.ID,
			issuer: email,
			err:    nil,
		},
		{
			desc:   "revoke non-existing user key",
			id:     newKey.ID,
			issuer: email,
			err:    nil,
		},
		{
			desc:   "revoke unauthorized",
			id:     newKey.ID,
			issuer: "",
			err:    authn.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		err := svc.Revoke(context.Background(), tc.issuer, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
	}
}
func TestRetrieve(t *testing.T) {
	svc := newService()
	loginKey, _, err := svc.Issue(context.Background(), "", authn.Key{Type: authn.UserKey, IssuedAt: time.Now(), Subject: email, IssuerID: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := authn.Key{
		ID:       "id",
		Type:     authn.APIKey,
		IssuerID: email,
		Subject:  email,
		IssuedAt: time.Now(),
	}

	userKey, userToken, err := svc.Issue(context.Background(), "", authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	apiKey, apiToken, err := svc.Issue(context.Background(), email, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing user's key expected to succeed: %s", err))
	resetKey, resetToken, err := svc.Issue(context.Background(), "", authn.Key{Type: authn.RecoveryKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing reset key expected to succeed: %s", err))

	cases := []struct {
		desc   string
		id     string
		issuer string
		err    error
	}{
		{
			desc:   "retrieve user key",
			id:     newKey.ID,
			issuer: userToken,
			err:    nil,
		},
		{
			desc:   "retrieve non-existing user key",
			id:     "invalid",
			issuer: email,
			err:    authn.ErrNotFound,
		},
		{
			desc:   "retrieve unauthorized",
			id:     newKey.ID,
			issuer: "wrong",
			err:    authn.ErrUnauthorizedAccess,
		},
		{
			desc:   "retrieve with user token",
			id:     newKey.ID,
			issuer: userToken,
			err:    authn.ErrUnauthorizedAccess,
		},
		{
			desc:   "retrieve with reset token",
			id:     newKey.ID,
			issuer: resetToken,
			err:    authn.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		_, err := svc.Retrieve(context.Background(), tc.issuer, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
	}
}
func TestIdentify(t *testing.T) {
	svc := newService()
	loginKey, _, err := svc.Issue(context.Background(), email, email, authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	recoveryKey, _, err := svc.Issue(context.Background(), email, email, authn.Key{Type: authn.RecoveryKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing reset key expected to succeed: %s", err))

	userKey, _, err := svc.Issue(context.Background(), loginKey.Email, loginKey.Email, authn.Key{Type: authn.APIKey, IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Minute)})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	exp1 := time.Now().Add(-2 * time.Second)
	expKey, _, err := svc.Issue(context.Background(), loginKey.Email, loginKey.Email, authn.Key{Type: authn.APIKey, IssuedAt: time.Now(), ExpiresAt: exp1})
	assert.Nil(t, err, fmt.Sprintf("Issuing expired user key expected to succeed: %s", err))

	invalidKey, _, err := svc.Issue(context.Background(), loginKey.Email, loginKey.Email, authn.Key{Type: 22, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	cases := []struct {
		desc string
		key  string
		id   string
		err  error
	}{
		{
			desc: "identify login key",
			key:  loginKey.Subject,
			id:   email,
			err:  nil,
		},
		{
			desc: "identify recovery key",
			key:  recoveryKey.Subject,
			id:   email,
			err:  nil,
		},
		{
			desc: "identify user key",
			key:  userKey.Subject,
			id:   email,
			err:  nil,
		},
		{
			desc: "identify expired user key",
			key:  expKey.Subject,
			id:   "",
			err:  authn.ErrKeyExpired,
		},
		{
			desc: "identify expired key",
			key:  invalidKey.Subject,
			id:   "",
			err:  authn.ErrUnauthorizedAccess,
		},
		{
			desc: "identify invalid key",
			key:  "invalid",
			id:   "",
			err:  authn.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		id, err := svc.Identify(context.Background(), tc.key)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.id, id, fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.id, id))
	}
}
