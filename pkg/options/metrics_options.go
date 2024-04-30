// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/jinzhu/copier"
	"github.com/spf13/pflag"
	"k8s.io/component-base/metrics"
)

var _ IOptions = (*MetricsOptions)(nil)

// MetricsOptions has all parameters needed for exposing metrics from components.
type MetricsOptions struct {
	ShowHiddenMetricsForVersion string            `json:"show-hidden-metrics-for-version" mapstructure:"show-hidden-metrics-for-version"`
	DisabledMetrics             []string          `json:"disabled-metrics" mapstructure:"disabled-metrics"`
	AllowListMapping            map[string]string `json:"allow-metric-labels" mapstructure:"allow-metric-labels"`
}

// NewMetricsOptions returns default metrics options.
func NewMetricsOptions() *MetricsOptions {
	opts := metrics.NewOptions()

	var o MetricsOptions
	_ = copier.Copy(&o, &opts)
	return &o
}

func (o *MetricsOptions) Native() *metrics.Options {
	var opts metrics.Options
	_ = copier.Copy(&opts, &o)
	return &opts
}

// Validate validates metrics flags options.
func (o *MetricsOptions) Validate() []error {
	return o.Native().Validate()
}

// AddFlags adds flags for exposing component metrics.
func (o *MetricsOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	if o == nil {
		return
	}
	fs.StringVar(&o.ShowHiddenMetricsForVersion, "metrics.show-hidden-metrics-for-version", o.ShowHiddenMetricsForVersion,
		"The previous version for which you want to show hidden metrics. "+
			"Only the previous minor version is meaningful, other values will not be allowed. "+
			"The format is <major>.<minor>, e.g.: '1.16'. "+
			"The purpose of this format is make sure you have the opportunity to notice if the next release hides additional metrics, "+
			"rather than being surprised when they are permanently removed in the release after that.")
	fs.StringSliceVar(&o.DisabledMetrics,
		"metrics.disabled-metrics",
		o.DisabledMetrics,
		"This flag provides an escape hatch for misbehaving metrics. "+
			"You must provide the fully qualified metric name in order to disable it. "+
			"Disclaimer: disabling metrics is higher in precedence than showing hidden metrics.")
	fs.StringToStringVar(&o.AllowListMapping, "metrics.allow-metric-labels", o.AllowListMapping,
		"The map from metric-label to value allow-list of this label. The key's format is <MetricName>,<LabelName>. "+
			"The value's format is <allowed_value>,<allowed_value>..."+
			"e.g. metric1,label1='v1,v2,v3', metric1,label2='v1,v2,v3' metric2,label1='v1,v2,v3'.")
}
