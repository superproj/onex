package klog

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

// C returns a logger derived from the context, falling back to the klog
// background logger when the context carries no logger.
func C(ctx context.Context) logr.Logger {
	return klog.FromContext(ctx)
}
