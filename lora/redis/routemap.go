// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/lora"
)

const (
	mfxMapPrefix  = "mfx:lora"
	loraMapPrefix = "lora:mfx"
)

var _ lora.RouteMapRepository = (*routerMap)(nil)

type routerMap struct {
	client *redis.Client
	prefix string
}

// NewRouteMapRepository returns redis thing cache implementation.
func NewRouteMapRepository(client *redis.Client, prefix string) lora.RouteMapRepository {
	return &routerMap{
		client: client,
		prefix: prefix,
	}
}

func (mr *routerMap) Save(ctx context.Context, mfxID, loraID string) error {
	tkey := fmt.Sprintf("%s:%s:%s", mr.prefix, mfxMapPrefix, mfxID)
	if err := mr.client.Set(ctx, tkey, loraID, 0).Err(); err != nil {
		return err
	}
	lkey := fmt.Sprintf("%s:%s:%s", mr.prefix, loraMapPrefix, loraID)
	if err := mr.client.Set(ctx, lkey, mfxID, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (mr *routerMap) Get(ctx context.Context, mfxID string) (string, error) {
	lKey := fmt.Sprintf("%s:%s:%s", mr.prefix, loraMapPrefix, mfxID)
	mval, err := mr.client.Get(ctx, lKey).Result()
	if err != nil {
		return "", err
	}

	return mval, nil
}

func (mr *routerMap) Remove(ctx context.Context, mfxID string) error {
	mkey := fmt.Sprintf("%s:%s:%s", mr.prefix, mfxMapPrefix, mfxID)
	lval, err := mr.client.Get(ctx, mkey).Result()
	if err != nil {
		return err
	}

	lkey := fmt.Sprintf("%s:%s:%s", mr.prefix, loraMapPrefix, lval)
	return mr.client.Del(ctx, mkey, lkey).Err()
}
