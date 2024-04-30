// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var _ IOptions = (*MongoOptions)(nil)

// MongoOptions contains options for connecting to a MongoDB server.
type MongoOptions struct {
	URL        string        `json:"url" mapstructure:"url"`
	Database   string        `json:"database" mapstructure:"database"`
	Collection string        `json:"collection" mapstructure:"collection"`
	Username   string        `json:"username" mapstructure:"username"`
	Password   string        `json:"password" mapstructure:"password"`
	Timeout    time.Duration `json:"timeout" mapstructure:"timeout"`
	TLSOptions *TLSOptions   `json:"tls" mapstructure:"tls"`
}

// NewMongoOptions create a `zero` value instance.
func NewMongoOptions() *MongoOptions {
	return &MongoOptions{
		Timeout:    30 * time.Second,
		TLSOptions: NewTLSOptions(),
	}
}

// Validate verifies flags passed to MongoOptions.
func (o *MongoOptions) Validate() []error {
	errs := []error{}

	if _, err := url.Parse(o.URL); err != nil {
		errs = append(errs, fmt.Errorf("unable to parse connection URL: %w", err))
	}

	if o.Database == "" {
		errs = append(errs, fmt.Errorf("--mongo.database can not be empty"))
	}

	if o.Collection == "" {
		errs = append(errs, fmt.Errorf("--mongo.collection can not be empty"))
	}

	if o.TLSOptions != nil {
		errs = append(errs, o.TLSOptions.Validate()...)
	}

	return errs
}

// AddFlags adds flags related to redis storage for a specific APIServer to the specified FlagSet.
func (o *MongoOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	o.TLSOptions.AddFlags(fs, "mongo")

	fs.DurationVar(&o.Timeout, "mongo.timeout", o.Timeout, "Timeout is the maximum amount of time a dial will wait for a connect to complete.")
	fs.StringVar(&o.URL, "mongo.url", o.URL, "The MongoDB server address.")
	fs.StringVar(&o.Database, "mongo.database", o.Database, "The MongoDB database name.")
	fs.StringVar(&o.Collection, "mongo.collection", o.Collection, "The MongoDB collection name.")
	fs.StringVar(&o.Username, "mongo.username", o.Username, "Username of the MongoDB database (optional).")
	fs.StringVar(&o.Password, "mongo.password", o.Password, "Password of the MongoDB database (optional).")
}

// NewClient creates a new MongoDB client based on the provided options.
func (o *MongoOptions) NewClient() (*mongo.Client, error) {
	// Set client options
	opts := options.Client().ApplyURI(o.URL).SetReadPreference(readpref.Primary())
	if o.Timeout > 0 {
		opts.SetConnectTimeout(o.Timeout).SetSocketTimeout(o.Timeout).SetServerSelectionTimeout(o.Timeout)
	}

	if o.Username != "" || o.Password != "" {
		opts.SetAuth(options.Credential{
			AuthSource: o.Database,
			Username:   o.Username,
			Password:   o.Password,
		})
	}

	if o.TLSOptions != nil {
		tlsConf, err := o.TLSOptions.TLSConfig()
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tlsConf)
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to check the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}
