// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package coap contains the domain concept definitions needed to support
// Mainflux coap adapter service functionality. All constant values are taken
// from RFC, and could be adjusted based on specific use case.
package coap

import (
	"fmt"
	"sync"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	chanID    = "id"
	keyHeader = "key"

	// AckRandomFactor is default ACK coefficient.
	AckRandomFactor = 1.5
	// AckTimeout is the amount of time to wait for a response.
	AckTimeout = 2000 * time.Millisecond
	// MaxRetransmit is the maximum number of times a message will be retransmitted.
	MaxRetransmit = 4
)

// Service specifies CoAP service API.
type Service interface {
	// Publish Messssage
	Publish(key string, msg messaging.Message) error

	// Subscribes to channel with specified id, subtopic and adds subscription to
	// service map of subscriptions under given ID.
	Subscribe(key, endpoint string, o Observer) error

	// Unsubscribe method is used to stop observing resource.
	Unsubscribe(key, endpoint, token string)
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
		// auth:      auth,
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

func (svc *adapterService) Publish(key string, msg messaging.Message) error {
	endpoint := fmt.Sprintf("%s.%s", msg.Channel, msg.Subtopic)
	svc.obsLock.RLock()
	for _, o := range svc.observers[endpoint] {
		o.Handle(msg)
	}
	svc.obsLock.RUnlock()
	return nil
	// return svc.ps.Publish(msg.Channel, msg)
}

func (svc *adapterService) Subscribe(key, endpoint string, o Observer) error {
	// subject := chanID
	// if subtopic != "" {
	// 	subject = fmt.Sprintf("%s.%s", chanID, subtopic)
	// }

	// err := svc.ps.Subscribe(subject, func(msg messaging.Message) error {
	// go func() {
	// 	for {
	// 		for _, o := range svc.observers[endpoint] {
	// 			err := o.Handle(messaging.Message{Payload: []byte("ddbdbdb")})
	// 			fmt.Println(err, reflect.TypeOf(err))
	// 			if err == context.Canceled {
	// 				fmt.Println("AAAA")
	// 				return
	// 			}
	// 			time.Sleep(time.Second * 5)
	// 		}
	// 	}
	// }()
	// })
	// if err != nil {
	// 	return err
	// }

	// go func() {
	// 	<-o.Cancel
	// 	if err := svc.ps.Unsubscribe(subject); err != nil {
	// 		// svc.log.Error(fmt.Sprintf("Failed to unsubscribe from %s.%s due to %s", chanID, subtopic, err))
	// 	}
	// }()

	// Put method removes Observer if already exists.
	go func() {
		<-o.Done()
		fmt.Println("finished", endpoint, o.Token())
		svc.remove(endpoint, o.Token())
	}()
	return svc.put(endpoint, o.Token(), o)
}

func (svc *adapterService) Unsubscribe(key, endpoint, token string) {
	svc.remove(endpoint, token)
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
	// If observer exists, cancel it and replace.
	if current, ok := obs[token]; ok {
		if err := current.Cancel(); err != nil {
			return err
		}
	}
	obs[token] = o
	return nil
}

func (svc *adapterService) remove(endpoint, token string) {
	svc.obsLock.Lock()
	defer svc.obsLock.Unlock()

	obs, ok := svc.observers[endpoint]
	if !ok {
		return
	}
	if current, ok := obs[token]; ok {
		if err := current.Cancel(); err != nil {
			return
		}
	}
	delete(obs, token)
	// If there are no observers left for the endpint, remove the map.
	if len(obs) == 0 {
		delete(svc.observers, endpoint)
	}
}
