// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"time"

	"github.com/jinzhu/copier"
	"github.com/spf13/pflag"
	logsapi "k8s.io/component-base/logs/api/v1"
)

var _ IOptions = (*LogsOptions)(nil)

// LogsOptions contains configuration items related to log.
type LogsOptions struct {
	// Format Flag specifies the structure of log messages.
	// default value of format is `text`
	Format string `json:"format,omitempty" mapstructure:"format"`
	// Maximum number of nanoseconds (i.e. 1s = 1000000000) between log
	// flushes. Ignored if the selected logging backend writes log
	// messages without buffering.
	FlushFrequency time.Duration `json:"flush-frequency" mapstructure:"flush-frequency"`
	// Verbosity is the threshold that determines which log messages are
	// logged. Default is zero which logs only the most important
	// messages. Higher values enable additional messages. Error messages
	// are always logged.
	Verbosity logsapi.VerbosityLevel `json:"verbosity" mapstructure:"verbosity"`
	// VModule overrides the verbosity threshold for individual files.
	// Only supported for "text" log format.
	VModule logsapi.VModuleConfiguration `json:"vmodule,omitempty" mapstructure:"vmodule"`
	// [Alpha] Options holds additional parameters that are specific
	// to the different logging formats. Only the options for the selected
	// format get used, but all of them get validated.
	// Only available when the LoggingAlphaOptions feature gate is enabled.
	Options logsapi.FormatOptions `json:"-,omitempty" mapstructure:"-"`
}

// NewLogsOptions creates an Options object with default parameters.
func NewLogsOptions() *LogsOptions {
	c := logsapi.LoggingConfiguration{}
	logsapi.SetRecommendedLoggingConfiguration(&c)

	var opts LogsOptions
	_ = copier.Copy(&opts, &c)
	return &opts
}

// Validate verifies flags passed to LogsOptions.
func (o *LogsOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds command line flags for the configuration.
func (o *LogsOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Format, "logs.format", o.Format, "Sets the log format. Permitted formats: json, text.")
	fs.DurationVar(&o.FlushFrequency, "log.flush-frequency", o.FlushFrequency, "Maximum number of seconds between log flushes.")
	fs.VarP(logsapi.VerbosityLevelPflag(&o.Verbosity), "logs.verbosity", "", " Number for the log level verbosity.")
	fs.Var(logsapi.VModuleConfigurationPflag(&o.VModule), "logs.vmodule", "Comma-separated list of pattern=N settings for file-filtered logging (only works for text log format).")
}

func (o *LogsOptions) Native() *logsapi.LoggingConfiguration {
	var cfg logsapi.LoggingConfiguration
	_ = copier.Copy(&cfg, &o)
	return &cfg
}
