// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "go.uber.org/automaxprocs"
	"k8s.io/component-base/cli"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"

	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
	"github.com/superproj/onex/pkg/version"
)

// App is the main structure of a cli application.
// It is recommended that an app be created with the app.NewApp() function.
type App struct {
	name        string
	shortDesc   string
	description string
	run         RunFunc
	cmd         *cobra.Command
	args        cobra.PositionalArgs

	// +optional
	healthCheckFunc HealthCheckFunc

	// +optional
	options CliOptions

	// +optional
	silence bool

	// +optional
	noConfig bool

	// watching and re-reading config files
	// +optional
	watch bool
}

// RunFunc defines the application's startup callback function.
type RunFunc func() error

// HealthCheckFunc defines the health check function for the application.
type HealthCheckFunc func() error

// Option defines optional parameters for initializing the application
// structure.
type Option func(*App)

// WithOptions to open the application's function to read from the command line
// or read parameters from the configuration file.
func WithOptions(opts CliOptions) Option {
	return func(a *App) {
		a.options = opts
	}
}

// WithRunFunc is used to set the application startup callback function option.
func WithRunFunc(run RunFunc) Option {
	return func(a *App) {
		a.run = run
	}
}

// WithDescription is used to set the description of the application.
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

// WithHealthCheckFunc is used to set the health check function for the application.
// The app framework will use the function to start a health check server.
func WithHealthCheckFunc(fn HealthCheckFunc) Option {
	return func(a *App) {
		a.healthCheckFunc = fn
	}
}

// WithDefaultHealthCheckFunc set the default health check function.
func WithDefaultHealthCheckFunc() Option {
	fn := func() HealthCheckFunc {
		return func() error {
			go genericoptions.NewHealthOptions().ServeHealthCheck()

			return nil
		}
	}

	return WithHealthCheckFunc(fn())
}

// WithSilence sets the application to silent mode, in which the program startup
// information, configuration information, and version information are not
// printed in the console.
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

// WithNoConfig set the application does not provide config flag.
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}

// WithValidArgs set the validation function to valid non-flag arguments.
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(a *App) {
		a.args = args
	}
}

// WithDefaultValidArgs set default validation function to valid non-flag arguments.
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}

			return nil
		}
	}
}

// WithWatchConfig watching and re-reading config files.
func WithWatchConfig() Option {
	return func(a *App) {
		a.watch = true
	}
}

// NewApp creates a new application instance based on the given application name,
// binary name, and other options.
func NewApp(name string, shortDesc string, opts ...Option) *App {
	a := &App{
		name:      name,
		run:       func() error { return nil },
		shortDesc: shortDesc,
	}

	for _, o := range opts {
		o(a)
	}

	a.buildCommand()

	return a
}

// buildCommand is used to build a cobra command.
func (a *App) buildCommand() {
	cmd := &cobra.Command{
		Use:   formatBaseName(a.name),
		Short: a.shortDesc,
		Long:  a.description,
		RunE:  a.runCommand,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return nil
		},
		Args: a.args,
	}
	// When error printing is enabled for the Cobra command, a flag parse
	// error gets printed first, then optionally the often long usage
	// text. This is very unreadable in a console because the last few
	// lines that will be visible on screen don't include the error.
	//
	// The recommendation from #sig-cli was to print the usage text, then
	// the error. We implement this consistently for all commands here.
	// However, we don't want to print the usage text when command
	// execution fails for other reasons than parsing. We detect this via
	// the FlagParseError callback.
	//
	// Some commands, like kubectl, already deal with this themselves.
	// We don't change the behavior for those.
	if !cmd.SilenceUsage {
		cmd.SilenceUsage = true
		cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
			// Re-enable usage printing.
			c.SilenceUsage = false
			return err
		})
	}
	// In all cases error printing is done below.
	cmd.SilenceErrors = true

	cmd.SetOutput(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true

	var fss cliflag.NamedFlagSets
	if a.options != nil {
		fss = a.options.Flags()
	}

	version.AddFlags(fss.FlagSet("global"))

	if !a.noConfig {
		AddConfigFlag(fss.FlagSet("global"), a.name, a.watch)
	}

	for _, f := range fss.FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cliflag.SetUsageAndHelpFunc(cmd, fss, cols)

	a.cmd = cmd
}

// Run is used to launch the application.
func (a *App) Run() {
	os.Exit(cli.Run(a.cmd))
}

func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	// display application version information
	version.PrintAndExitIfRequested(a.name)

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if a.options != nil {
		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}

		// set default options
		if err := a.options.Complete(); err != nil {
			return err
		}

		// validate options
		if err := a.options.Validate(); err != nil {
			return err
		}
	}

	// 初始化日志
	log.Init(logOptions())
	defer log.Sync() // Sync 将缓存中的日志刷新到磁盘文件中

	if !a.silence {
		log.Infow("Starting application", "name", a.name, "version", version.Get().ToJSON())
		log.Infow("Golang settings", "GOGC", os.Getenv("GOGC"), "GOMAXPROCS", os.Getenv("GOMAXPROCS"), "GOTRACEBACK", os.Getenv("GOTRACEBACK"))
		if !a.noConfig {
			PrintConfig()
		} else if a.options != nil {
			cliflag.PrintFlags(cmd.Flags())
		}
	}

	if a.healthCheckFunc != nil {
		if err := a.healthCheckFunc(); err != nil {
			return err
		}
	}

	// run application
	return a.run()
}

// Command returns cobra command instance inside the application.
func (a *App) Command() *cobra.Command {
	return a.cmd
}

// formatBaseName is formatted as an executable file name under different
// operating systems according to the given name.
func formatBaseName(name string) string {
	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		name = strings.ToLower(name)
		name = strings.TrimSuffix(name, ".exe")
	}
	return name
}

// logOptions 从 viper 中读取日志配置，构建 `*log.Options` 并返回.
// 注意：`viper.Get<Type>()` 中 key 的名字需要使用 `.` 分割，以跟 YAML 中保持相同的缩进.
func logOptions() *log.Options {
	return &log.Options{
		DisableCaller:     viper.GetBool("log.disable-caller"),
		DisableStacktrace: viper.GetBool("log.disable-stacktrace"),
		Level:             viper.GetString("log.level"),
		Format:            viper.GetString("log.format"),
		EnableColor:       viper.GetBool("log.enable-color"),
		OutputPaths:       viper.GetStringSlice("log.output-paths"),
	}
}

func init() {
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output-paths", []string{"stdout"})
}
