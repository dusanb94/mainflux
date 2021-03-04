package lora

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	protocol      = "lora"
	thingSuffix   = "thing"
	channelSuffix = "channel"
)

var (
	// ErrMalformedMessage indicates malformed LoRa message.
	ErrMalformedMessage = errors.New("malformed message received")

	// ErrNotFoundDev indicates a non-existent route map for a device EUI.
	ErrNotFoundDev = errors.New("route map not found for this device EUI")

	// ErrNotFoundApp indicates a non-existent route map for an application ID.
	ErrNotFoundApp = errors.New("route map not found for this application ID")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThing creates thingID:devEUI route-map
	CreateThing(ctx context.Context, thingID string, devEUI string) error

	// UpdateThing updates thingID:devEUI route-map
	UpdateThing(ctx context.Context, thingID string, devEUI string) error

	// RemoveThing removes thingID:devEUI route-map
	RemoveThing(ctx context.Context, thingID string) error

	// CreateChannel creates channelID:appID route-map
	CreateChannel(ctx context.Context, chanID string, appID string) error

	// UpdateChannel updates channelID:appID route-map
	UpdateChannel(ctx context.Context, chanID string, appID string) error

	// RemoveChannel removes channelID:appID route-map
	RemoveChannel(ctx context.Context, chanID string) error

	// Publish forwards messages from the LoRa MQTT broker to Mainflux NATS broker
	Publish(ctx context.Context, token string, msg Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher  messaging.Publisher
	thingsRM   RouteMapRepository
	channelsRM RouteMapRepository
}

// New instantiates the LoRa adapter implementation.
func New(publisher messaging.Publisher, thingsRM, channelsRM RouteMapRepository) Service {
	return &adapterService{
		publisher:  publisher,
		thingsRM:   thingsRM,
		channelsRM: channelsRM,
	}
}

// Publish forwards messages from Lora MQTT broker to Mainflux NATS broker
func (as *adapterService) Publish(ctx context.Context, token string, m Message) error {
	// Get route map of lora application
	thing, err := as.thingsRM.Get(ctx, m.DevEUI)
	if err != nil {
		return ErrNotFoundDev
	}

	// Get route map of lora application
	channel, err := as.channelsRM.Get(ctx, m.ApplicationID)
	if err != nil {
		return ErrNotFoundApp
	}

	// Use the SenML message decoded on LoRa server application if
	// field Object isn't empty. Otherwise, decode standard field Data.
	var payload []byte
	switch m.Object {
	case nil:
		payload, err = base64.StdEncoding.DecodeString(m.Data)
		if err != nil {
			return ErrMalformedMessage
		}
	default:
		jo, err := json.Marshal(m.Object)
		if err != nil {
			return err
		}
		payload = []byte(jo)
	}

	// Publish on Mainflux NATS broker
	msg := messaging.Message{
		Publisher: thing,
		Protocol:  protocol,
		Channel:   channel,
		Payload:   payload,
		Created:   time.Now().UnixNano(),
	}

	return as.publisher.Publish(msg.Channel, msg)
}

func (as *adapterService) CreateThing(ctx context.Context, thingID string, devEUI string) error {
	return as.thingsRM.Save(ctx, thingID, devEUI)
}

func (as *adapterService) UpdateThing(ctx context.Context, thingID string, devEUI string) error {
	return as.thingsRM.Save(ctx, thingID, devEUI)
}

func (as *adapterService) RemoveThing(ctx context.Context, thingID string) error {
	return as.thingsRM.Remove(ctx, thingID)
}

func (as *adapterService) CreateChannel(ctx context.Context, chanID string, appID string) error {
	return as.channelsRM.Save(ctx, chanID, appID)
}

func (as *adapterService) UpdateChannel(ctx context.Context, chanID string, appID string) error {
	return as.channelsRM.Save(ctx, chanID, appID)
}

func (as *adapterService) RemoveChannel(ctx context.Context, chanID string) error {
	return as.channelsRM.Remove(ctx, chanID)
}
