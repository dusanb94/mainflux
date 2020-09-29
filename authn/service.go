// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	loginDuration    = 10 * time.Hour
	recoveryDuration = 5 * time.Minute
	issuerName       = "mainflux.authn"
)

var (
	// ErrUnauthorizedAccess represents unauthorized access.
	ErrUnauthorizedAccess = errors.New("unauthorized access")

	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid owner or ID).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrNotFound indicates a non-existing entity request.
	ErrNotFound = errors.New("entity not found")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")

	errIssueUser = errors.New("failed to issue new user key")
	errIssueTmp  = errors.New("failed to issue new temporary key")
	errRevoke    = errors.New("failed to remove key")
	errRetrieve  = errors.New("failed to retrieve key data")
	errIdentify  = errors.New("failed to validate token")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Issue issues a new Key.
	Issue(ctx context.Context, id, email string, key Key) (Key, string, error)

	// Revoke removes the Key with the provided id that is
	// issued by the user identified by the provided key.
	Revoke(ctx context.Context, issuer, id string) error

	// Retrieve retrieves data for the Key identified by the provided
	// ID, that is issued by the user identified by the provided key.
	Retrieve(ctx context.Context, issuer, id string) (Key, error)

	// Identify validates token token. If token is valid, content
	// is returned. If token is invalid, or invocation failed for some
	// other reason, non-nil error value is returned in response.
	Identify(ctx context.Context, token string) (Identity, error)
}

var _ Service = (*service)(nil)

type service struct {
	keys         KeyRepository
	uuidProvider mainflux.UUIDProvider
	tokenizer    Tokenizer
}

// New instantiates the auth service implementation.
func New(keys KeyRepository, up mainflux.UUIDProvider, tokenizer Tokenizer) Service {
	return &service{
		tokenizer:    tokenizer,
		keys:         keys,
		uuidProvider: up,
	}
}

func (svc service) Issue(ctx context.Context, id, email string, key Key) (Key, string, error) {
	if key.IssuedAt.IsZero() {
		return Key{}, "", ErrInvalidKeyIssuedAt
	}
	switch key.Type {
	case APIKey:
		return svc.userKey(ctx, id, email, key)
	case RecoveryKey:
		return svc.tmpKey(id, email, recoveryDuration, key)
	default:
		return svc.tmpKey(id, email, loginDuration, key)
	}
}

func (svc service) Revoke(ctx context.Context, issuer, id string) error {
	email, err := svc.login(issuer)
	if err != nil {
		return errors.Wrap(errRevoke, err)
	}
	if err := svc.keys.Remove(ctx, email, id); err != nil {
		return errors.Wrap(errRevoke, err)
	}
	return nil
}

func (svc service) Retrieve(ctx context.Context, issuer, id string) (Key, error) {
	email, err := svc.login(issuer)
	if err != nil {
		return Key{}, errors.Wrap(errRetrieve, err)
	}

	return svc.keys.Retrieve(ctx, email, id)
}

func (svc service) Identify(ctx context.Context, token string) (Identity, error) {
	c, err := svc.tokenizer.Parse(token)
	if err != nil {
		return Identity{}, errors.Wrap(errIdentify, err)
	}

	switch c.Type {
	case APIKey:
		k, err := svc.keys.Retrieve(ctx, c.Issuer, c.ID)
		if err != nil {
			return Identity{}, err
		}
		// Auto revoke expired key.
		if k.Expired() {
			svc.keys.Remove(ctx, c.Issuer, c.ID)
			return Identity{}, ErrKeyExpired
		}
		return Identity{Email: c.Email}, nil
	case RecoveryKey, UserKey:
		if c.Issuer != issuerName {
			return Identity{}, ErrUnauthorizedAccess
		}
		return Identity{Email: c.Email}, nil
	default:
		return Identity{}, ErrUnauthorizedAccess
	}
}

func (svc service) tmpKey(id, email string, duration time.Duration, key Key) (Key, string, error) {
	key.Email = email
	key.Issuer = issuerName
	key.ExpiresAt = key.IssuedAt.Add(duration)
	secret, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Key{}, "", errors.Wrap(errIssueTmp, err)
	}

	key.Email = secret
	return key, secret, nil
}

func (svc service) userKey(ctx context.Context, id, email string, key Key) (Key, string, error) {
	email, err := svc.login(id)
	if err != nil {
		return Key{}, "", errors.Wrap(errIssueUser, err)
	}
	key.Issuer = email

	keyID, err := svc.uuidProvider.ID()
	if err != nil {
		return Key{}, "", errors.Wrap(errIssueUser, err)
	}
	key.ID = keyID

	secret, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Key{}, "", errors.Wrap(errIssueUser, err)
	}
	key.Email = secret

	if _, err := svc.keys.Save(ctx, key); err != nil {
		return Key{}, "", errors.Wrap(errIssueUser, err)
	}

	return key, secret, nil
}

func (svc service) login(token string) (string, error) {
	c, err := svc.tokenizer.Parse(token)
	if err != nil {
		return "", err
	}
	// Only user key token is valid for login.
	if c.Type != UserKey || c.Email == "" {
		return "", ErrUnauthorizedAccess
	}

	return c.Email, nil
}
