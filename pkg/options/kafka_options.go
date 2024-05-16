// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/segmentio/kafka-go/snappy"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	stringsutil "github.com/superproj/onex/pkg/util/strings"
)

var _ IOptions = (*KafkaOptions)(nil)

type logger struct {
	v int32
}

func (l *logger) Printf(format string, args ...any) {
	klog.V(klog.Level(l.v)).Infof(format, args...)
}

type WriterOptions struct {
	// Limit on how many attempts will be made to deliver a message.
	//
	// The default is to try at most 10 times.
	MaxAttempts int `mapstructure:"max-attempts"`

	// Number of acknowledges from partition replicas required before receiving
	// a response to a produce request. The default is -1, which means to wait for
	// all replicas, and a value above 0 is required to indicate how many replicas
	// should acknowledge a message to be considered successful.
	//
	// This version of kafka-go (v0.3) does not support 0 required acks, due to
	// some internal complexity implementing this with the Kafka protocol. If you
	// need that functionality specifically, you'll need to upgrade to v0.4.
	RequiredAcks int `mapstructure:"required-acks"`

	// Setting this flag to true causes the WriteMessages method to never block.
	// It also means that errors are ignored since the caller will not receive
	// the returned value. Use this only if you don't care about guarantees of
	// whether the messages were written to kafka.
	Async bool `mapstructure:"async"`

	// Limit on how many messages will be buffered before being sent to a
	// partition.
	//
	// The default is to use a target batch size of 100 messages.
	BatchSize int `mapstructure:"batch-size"`

	// Time limit on how often incomplete message batches will be flushed to
	// kafka.
	//
	// The default is to flush at least every second.
	BatchTimeout time.Duration `mapstructure:"batch-timeout"`

	// Limit the maximum size of a request in bytes before being sent to
	// a partition.
	//
	// The default is to use a kafka default value of 1048576.
	BatchBytes int `mapstructure:"batch-bytes"`
}

type ReaderOptions struct {
	// GroupID holds the optional consumer group id. If GroupID is specified, then
	// Partition should NOT be specified e.g. 0
	GroupID string `mapstructure:"group-id"`

	// GroupTopics allows specifying multiple topics, but can only be used in
	// combination with GroupID, as it is a consumer-group feature. As such, if
	// GroupID is set, then either Topic or GroupTopics must be defined.
	// GroupTopics []string

	// Partition to read messages from.  Either Partition or GroupID may
	// be assigned, but not both
	Partition int `mapstructure:"partition"`

	// The capacity of the internal message queue, defaults to 100 if none is
	// set.
	QueueCapacity int `mapstructure:"queue-capacity"`

	// MinBytes indicates to the broker the minimum batch size that the consumer
	// will accept. Setting a high minimum when consuming from a low-volume topic
	// may result in delayed delivery when the broker does not have enough data to
	// satisfy the defined minimum.
	//
	// Default: 1
	MinBytes int `mapstructure:"min-bytes"`

	// MaxBytes indicates to the broker the maximum batch size that the consumer
	// will accept. The broker will truncate a message to satisfy this maximum, so
	// choose a value that is high enough for your largest message size.
	//
	// Default: 1MB
	MaxBytes int `mapstructure:"max-bytes"`

	// Maximum amount of time to wait for new data to come when fetching batches
	// of messages from kafka.
	//
	// Default: 10s
	MaxWait time.Duration `mapstructure:"max-wait"`

	// ReadBatchTimeout amount of time to wait to fetch message from kafka messages batch.
	//
	// Default: 10s
	ReadBatchTimeout time.Duration `mapstructure:"read-batch-timeout"`

	// ReadLagInterval sets the frequency at which the reader lag is updated.
	// Setting this field to a negative value disables lag reporting.
	// ReadLagInterval time.Duration

	// HeartbeatInterval sets the optional frequency at which the reader sends the consumer
	// group heartbeat update.
	//
	// Default: 3s
	//
	// Only used when GroupID is set
	HeartbeatInterval time.Duration `mapstructure:"heartbeat-interval"`

	// CommitInterval indicates the interval at which offsets are committed to
	// the broker.  If 0, commits will be handled synchronously.
	//
	// Default: 0
	//
	// Only used when GroupID is set
	CommitInterval time.Duration `mapstructure:"commit-interval"`

	// RebalanceTimeout optionally sets the length of time the coordinator will wait
	// for members to join as part of a rebalance.  For kafka servers under higher
	// load, it may be useful to set this value higher.
	//
	// Default: 30s
	//
	// Only used when GroupID is set
	RebalanceTimeout time.Duration `mapstructure:"rebalance-timeout"`

	// StartOffset determines from whence the consumer group should begin
	// consuming when it finds a partition without a committed offset.  If
	// non-zero, it must be set to one of FirstOffset or LastOffset.
	//
	// Default: FirstOffset
	//
	// Only used when GroupID is set
	StartOffset int64 `mapstructure:"start-offset"`

	// Limit of how many attempts will be made before delivering the error.
	//
	// The default is to try 3 times.
	MaxAttempts int `mapstructure:"max-attempts"`
}

// KafkaOptions defines options for kafka cluster.
// Common options for kafka-go reader and writer.
type KafkaOptions struct {
	// kafka-go reader and writer common options
	Brokers       []string      `mapstructure:"brokers"`
	Topic         string        `mapstructure:"topic"`
	ClientID      string        `mapstructure:"client-id"`
	Timeout       time.Duration `mapstructure:"timeout"`
	TLSOptions    *TLSOptions   `mapstructure:"tls"`
	SASLMechanism string        `mapstructure:"mechanism"`
	Username      string        `mapstructure:"username"`
	Password      string        `mapstructure:"password"`
	Algorithm     string        `mapstructure:"algorithm"`
	Compressed    bool          `mapstructure:"compressed"`

	// kafka-go writer options
	WriterOptions WriterOptions `mapstructure:"writer"`

	// kafka-go reader options
	ReaderOptions ReaderOptions `mapstructure:"reader"`
}

// NewKafkaOptions create a `zero` value instance.
func NewKafkaOptions() *KafkaOptions {
	return &KafkaOptions{
		TLSOptions: NewTLSOptions(),
		Timeout:    3 * time.Second,
		WriterOptions: WriterOptions{
			RequiredAcks: 1,
			MaxAttempts:  10,
			Async:        true,
			BatchSize:    100,
			BatchTimeout: 1 * time.Second,
			BatchBytes:   1 * MiB,
		},
		ReaderOptions: ReaderOptions{
			QueueCapacity:     100,
			MinBytes:          1,
			MaxBytes:          1 * MiB,
			MaxWait:           10 * time.Second,
			ReadBatchTimeout:  10 * time.Second,
			HeartbeatInterval: 3 * time.Second,
			CommitInterval:    0 * time.Second,
			RebalanceTimeout:  30 * time.Second,
			StartOffset:       kafka.FirstOffset,
			MaxAttempts:       3,
		},
	}
}

// Validate verifies flags passed to KafkaOptions.
func (o *KafkaOptions) Validate() []error {
	errs := []error{}

	if len(o.Brokers) == 0 {
		errs = append(errs, fmt.Errorf("kafka broker can not be empty"))
	}

	if !o.TLSOptions.UseTLS && o.SASLMechanism != "" {
		errs = append(errs, fmt.Errorf("SASL-Mechanism is setted but use_ssl is false"))
	}

	if !stringsutil.StringIn(strings.ToLower(o.SASLMechanism), []string{"plain", "scram", ""}) {
		errs = append(errs, fmt.Errorf("doesn't support '%s' SASL mechanism", o.SASLMechanism))
	}

	if o.Timeout <= 0 {
		errs = append(errs, fmt.Errorf("--kafka.timeout cannot be negative"))
	}

	if o.ReaderOptions.GroupID != "" && o.ReaderOptions.Partition != 0 {
		errs = append(errs, fmt.Errorf("either Partition or GroupID may be assigned, but not both"))
	}

	if o.WriterOptions.BatchTimeout <= 0 {
		errs = append(errs, fmt.Errorf("--kafka.writer.batch-timeout cannot be negative"))
	}

	errs = append(errs, o.TLSOptions.Validate()...)

	return errs
}

// AddFlags adds flags related to redis storage for a specific APIServer to the specified FlagSet.
func (o *KafkaOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	o.TLSOptions.AddFlags(fs, "kafka")

	fs.StringSliceVar(&o.Brokers, "kafka.brokers", o.Brokers, "The list of brokers used to discover the partitions available on the kafka cluster.")
	fs.StringVar(&o.Topic, "kafka.topic", o.Topic, "The topic that the writer/reader will produce/consume messages to.")
	fs.StringVar(&o.ClientID, "kafka.client-id", o.ClientID, " Unique identifier for client connections established by this Dialer. ")
	fs.DurationVar(&o.Timeout, "kafka.timeout", o.Timeout, "Timeout is the maximum amount of time a dial will wait for a connect to complete.")
	fs.StringVar(&o.SASLMechanism, "kafka.mechanism", o.SASLMechanism, "Configures the Dialer to use SASL authentication.")
	fs.StringVar(&o.Username, "kafka.username", o.Username, "Username of the kafka cluster.")
	fs.StringVar(&o.Password, "kafka.password", o.Password, "Password of the kafka cluster.")
	fs.StringVar(&o.Algorithm, "kafka.algorithm", o.Algorithm, "Algorithm used to create sasl.Mechanism.")
	fs.BoolVar(&o.Compressed, "kafka.compressed", o.Compressed, "compressed is used to specify whether compress Kafka messages.")
	fs.IntVar(&o.WriterOptions.RequiredAcks, "kafka.required-acks", o.WriterOptions.RequiredAcks, ""+
		"Number of acknowledges from partition replicas required before receiving a response to a produce request.")
	fs.IntVar(&o.WriterOptions.MaxAttempts, "kafka.writer.max-attempts", o.WriterOptions.MaxAttempts, ""+
		"Limit on how many attempts will be made to deliver a message.")
	fs.BoolVar(&o.WriterOptions.Async, "kafka.writer.async", o.WriterOptions.Async, "Limit on how many attempts will be made to deliver a message.")
	fs.IntVar(&o.WriterOptions.BatchSize, "kafka.writer.batch-size", o.WriterOptions.BatchSize, ""+
		"Limit on how many messages will be buffered before being sent to a partition.")
	fs.DurationVar(&o.WriterOptions.BatchTimeout, "kafka.writer.batch-timeout", o.WriterOptions.BatchTimeout, ""+
		"Time limit on how often incomplete message batches will be flushed to kafka.")
	fs.IntVar(&o.WriterOptions.BatchBytes, "kafka.writer.batch-bytes", o.WriterOptions.BatchBytes, ""+
		"Limit the maximum size of a request in bytes before being sent to a partition.")
	fs.StringVar(&o.ReaderOptions.GroupID, "kafka.reader.group-id", o.ReaderOptions.GroupID, ""+
		"GroupID holds the optional consumer group id. If GroupID is specified, then Partition should NOT be specified e.g. 0.")
	fs.IntVar(&o.ReaderOptions.Partition, "kafka.reader.partition", o.ReaderOptions.Partition, "Partition to read messages from.")
	fs.IntVar(&o.ReaderOptions.QueueCapacity, "kafka.reader.queue-capacity", o.ReaderOptions.QueueCapacity, ""+
		"The capacity of the internal message queue, defaults to 100 if none is set.")
	fs.IntVar(&o.ReaderOptions.MinBytes, "kafka.reader.min-bytes", o.ReaderOptions.MinBytes, ""+
		"MinBytes indicates to the broker the minimum batch size that the consumer will accept.")
	fs.IntVar(&o.ReaderOptions.MaxBytes, "kafka.reader.max-bytes", o.ReaderOptions.MaxBytes, ""+
		"MaxBytes indicates to the broker the maximum batch size that the consumer will accept.")
	fs.DurationVar(&o.ReaderOptions.MaxWait, "kafka.reader.max-wait", o.ReaderOptions.MaxWait, ""+
		"Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.")
	fs.DurationVar(&o.ReaderOptions.ReadBatchTimeout, "kafka.reader.read-batch-timeout", o.ReaderOptions.ReadBatchTimeout, ""+
		"ReadBatchTimeout amount of time to wait to fetch message from kafka messages batch.")
	fs.DurationVar(&o.ReaderOptions.HeartbeatInterval, "kafka.reader.heartbeat-interval", o.ReaderOptions.HeartbeatInterval, ""+
		"HeartbeatInterval sets the optional frequency at which the reader sends the consumer group heartbeat update.")
	fs.DurationVar(&o.ReaderOptions.CommitInterval, "kafka.reader.commit-interval", o.ReaderOptions.CommitInterval, ""+
		"CommitInterval indicates the interval at which offsets are committed to the broker.")
	fs.DurationVar(&o.ReaderOptions.RebalanceTimeout, "kafka.reader.rebalance-timeout", o.ReaderOptions.RebalanceTimeout, ""+
		"RebalanceTimeout optionally sets the length of time the coordinator will wait for members to join as part of a rebalance.")
	fs.Int64Var(&o.ReaderOptions.StartOffset, "kafka.reader.start-offset", o.ReaderOptions.StartOffset, ""+
		"StartOffset determines from whence the consumer group should begin consuming when it finds a partition without a committed offset.")
	fs.IntVar(&o.ReaderOptions.MaxAttempts, "kafka.reader.max-attempts", o.ReaderOptions.MaxAttempts, ""+
		"Limit of how many attempts will be made before delivering the error. ")
}

func (o *KafkaOptions) GetMechanism() (sasl.Mechanism, error) {
	var mechanism sasl.Mechanism

	switch o.SASLMechanism {
	case "":
		break
	case "PLAIN", "plain":
		mechanism = plain.Mechanism{Username: o.Username, Password: o.Password}
	case "SCRAM", "scram":
		algorithm := scram.SHA256
		if o.Algorithm == "sha-512" || o.Algorithm == "SHA-512" {
			algorithm = scram.SHA512
		}
		var err error
		mechanism, err = scram.Mechanism(algorithm, o.Username, o.Password)
		if err != nil {
			return nil, fmt.Errorf("failed initialize kafka mechanism: %w", err)
		}
	default:
	}

	return mechanism, nil
}

func (o *KafkaOptions) Dialer() (*kafka.Dialer, error) {
	tlsConfig, err := o.TLSOptions.TLSConfig()
	if err != nil {
		return nil, err
	}

	mechanism, err := o.GetMechanism()
	if err != nil {
		return nil, err
	}

	return &kafka.Dialer{
		Timeout:       o.Timeout,
		ClientID:      o.ClientID,
		TLS:           tlsConfig,
		SASLMechanism: mechanism,
	}, nil
}

func (o *KafkaOptions) Writer() (*kafka.Writer, error) {
	dialer, err := o.Dialer()
	if err != nil {
		return nil, err
	}

	// Kafka writer connection config
	config := kafka.WriterConfig{
		Brokers:      o.Brokers,
		Topic:        o.Topic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		WriteTimeout: o.Timeout,
		ReadTimeout:  o.Timeout,

		Async:        o.WriterOptions.Async,
		BatchSize:    o.WriterOptions.BatchSize,
		BatchBytes:   o.WriterOptions.BatchBytes,
		BatchTimeout: o.WriterOptions.BatchTimeout,
		MaxAttempts:  o.WriterOptions.MaxAttempts,
		Logger:       &logger{4},
		ErrorLogger:  &logger{1},
	}

	if o.Compressed {
		config.CompressionCodec = snappy.NewCompressionCodec()
	}

	kafkaWriter := kafka.NewWriter(config)
	return kafkaWriter, nil
}
