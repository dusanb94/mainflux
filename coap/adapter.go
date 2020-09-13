// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package coap contains the domain concept definitions needed to support
// Mainflux coap adapter service functionality. All constant values are taken
// from RFC, and could be adjusted based on specific use case.
package coap

import (
	"context"
	"fmt"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/messaging"
)

// Exported errors
var (
	ErrUnauthorized = errors.New("unauthorized access")
	ErrUnsubscribe  = errors.New("unable to unsubscribe")
)

// Service specifies CoAP service API.
type Service interface {
	// Publish Messssage
	Publish(ctx context.Context, key string, msg messaging.Message) error

	// Subscribes to channel with specified id, subtopic and adds subscription to
	// service map of subscriptions under given ID.
	Subscribe(ctx context.Context, key, chanID, subtopic string, o Observer) error

	// Unsubscribe method is used to stop observing resource.
	Unsubscribe(ctx context.Context, key, chanID, subptopic, token string) error
}

var _ Service = (*adapterService)(nil)

// Observers is a map of maps,
type adapterService struct {
	auth      mainflux.ThingsServiceClient
	ps        messaging.PubSub
	observers map[string]observers
	obsLock   sync.RWMutex
}

// New instantiates the CoAP adapter implementation.
func New(auth mainflux.ThingsServiceClient, ps messaging.PubSub) Service {
	as := &adapterService{
		auth: auth,
		// ps:        ps,
		observers: make(map[string]observers),
		obsLock:   sync.RWMutex{},
	}

	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 5)
	// 		fmt.Println("testing size... ")
	// 		as.obsLock.RLock()
	// 		fmt.Println(as.observers)
	// 		fmt.Println("Number of goroutines", runtime.NumGoroutine())
	// 		as.obsLock.RUnlock()
	// 	}
	// }()

	return as
}

func (svc *adapterService) Publish(ctx context.Context, key string, msg messaging.Message) error {
	ar := &mainflux.AccessByKeyReq{
		Token:  key,
		ChanID: msg.Channel,
	}
	thid, err := svc.auth.CanAccessByKey(ctx, ar)
	if err != nil {
		return err
	}
	msg.Publisher = thid.GetValue()

	endpoint := fmt.Sprintf("%s.%s", msg.Channel, msg.Subtopic)
	svc.obsLock.RLock()
	for _, o := range svc.observers[endpoint] {
		o.Handle(msg)
	}
	svc.obsLock.RUnlock()
	return nil
	// return svc.ps.Publish(msg.Channel, msg)
}

func (svc *adapterService) Subscribe(ctx context.Context, key, chanID, subtopic string, o Observer) error {
	ar := &mainflux.AccessByKeyReq{
		Token:  key,
		ChanID: chanID,
	}
	_, err := svc.auth.CanAccessByKey(ctx, ar)
	if err != nil {
		return errors.Wrap(ErrUnauthorized, err)
	}
	endpoint := fmt.Sprintf("%s.%s", chanID, subtopic)

	go func() {
		<-o.Done()
		fmt.Println("finished", endpoint, o.Token())
		svc.remove(endpoint, o.Token())
	}()
	return svc.put(endpoint, o.Token(), o)
}

func (svc *adapterService) Unsubscribe(ctx context.Context, key, chanID, subtopic, token string) error {
	ar := &mainflux.AccessByKeyReq{
		Token:  key,
		ChanID: chanID,
	}
	_, err := svc.auth.CanAccessByKey(ctx, ar)
	if err != nil {
		return errors.Wrap(ErrUnauthorized, err)
	}
	endpoint := fmt.Sprintf("%s.%s", chanID, subtopic)

	return svc.remove(endpoint, token)
}

func (svc *adapterService) get(topic, token string) (Observer, bool) {
	svc.obsLock.RLock()
	defer svc.obsLock.RUnlock()

	obs, ok := svc.observers[topic]
	if !ok {
		return nil, ok
	}
	o, ok := obs[token]
	return o, ok
}

func (svc *adapterService) put(endpoint, token string, o Observer) error {
	svc.obsLock.Lock()
	defer svc.obsLock.Unlock()

	obs, ok := svc.observers[endpoint]
	// If there are no observers, create map and assign it to the endpoint.
	if !ok {
		obs = observers{token: o}
		svc.observers[endpoint] = obs
		return nil
	}
	// If observer exists, cancel subscription and replace it.
	if sub, ok := obs[token]; ok {
		if err := sub.Cancel(); err != nil {
			return errors.Wrap(ErrUnsubscribe, err)
		}
	}
	obs[token] = o
	return nil
}

func (svc *adapterService) remove(endpoint, token string) error {
	svc.obsLock.Lock()
	defer svc.obsLock.Unlock()

	obs, ok := svc.observers[endpoint]
	if !ok {
		return nil
	}
	if current, ok := obs[token]; ok {
		if err := current.Cancel(); err != nil {
			return errors.Wrap(ErrUnsubscribe, err)
		}
	}
	delete(obs, token)
	// If there are no observers left for the endpint, remove the map.
	if len(obs) == 0 {
		delete(svc.observers, endpoint)
	}
	return nil
}
