// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config contains necessary redis options.
type Config struct {
	Addr     string
	Username string
	Password string
	Database int
	// Sore key prefix.
	KeyPrefix string
}

// Store redis storage.
type Store struct {
	cli    *redis.Client
	prefix string
}

// NewStore create an *Store instance to handle token storage, deletion, and checking.
func NewStore(cfg *Config) *Store {
	// The reason `github.com/superproj/onex/pkg/db` is not used here is
	// to minimize dependencies, and use `github.com/redis/go-redis/v9` to
	// create redis client is not complex.
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.Database,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	return &Store{cli: cli, prefix: cfg.KeyPrefix}
}

// wrapperKey is used to build the key name in Redis.
func (s *Store) wrapperKey(key string) string {
	return fmt.Sprintf("%s%s", s.prefix, key)
}

// Set call the Redis client to set a key-value pair with an
// expiration time, where the key name format is <prefix><accessToken>.
func (s *Store) Set(ctx context.Context, accessToken string, expiration time.Duration) error {
	cmd := s.cli.Set(ctx, s.wrapperKey(accessToken), "1", expiration)
	return cmd.Err()
}

// Delete delete the specified JWT Token in Redis.
func (s *Store) Delete(ctx context.Context, accessToken string) (bool, error) {
	cmd := s.cli.Del(ctx, s.wrapperKey(accessToken))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Check check if the specified JWT Token exists in Redis.
func (s *Store) Check(ctx context.Context, accessToken string) (bool, error) {
	cmd := s.cli.Exists(ctx, s.wrapperKey(accessToken))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Close is used to close the redis client.
func (s *Store) Close() error {
	return s.cli.Close()
}
