package onex

import (
	"go.uber.org/ratelimit"
)

const (
	// defaultMaxConcurrency defines the default maximum number of concurrent embeddings.
	defaultMaxConcurrency = 100
	// defaultRateLimit defines the default rate limit for embedding requests.
	defaultRateLimit = 2000
)

// Option is a function type that modifies the onexEmbedder configuration.
type Option func(emb *onexEmbedder)

// WithMaxConcurrency returns an Option that sets the maximum concurrency level for the embedder.
func WithMaxConcurrency(concurrency int) Option {
	return func(emb *onexEmbedder) {
		emb.concurrency = concurrency
	}
}

// WithRateLimiter returns an Option that sets a custom rate limiter for the embedder.
func WithRateLimiter(rl ratelimit.Limiter) Option {
	return func(emb *onexEmbedder) {
		emb.rl = rl
	}
}
