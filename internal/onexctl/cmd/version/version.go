// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package version print the client and server version information.
package version

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
	"github.com/superproj/onex/pkg/version"
)

// Version is a struct for version information.
type Version struct {
	ClientVersion *version.Info `json:"clientVersion,omitempty" yaml:"clientVersion,omitempty"`
	ServerVersion *version.Info `json:"gatewayVersion,omitempty" yaml:"gatewayVersion,omitempty"`
}

var versionExample = templates.Examples(`
		# Print the client and server versions for the current context
		onexctl version`)

// Options is a struct to support version command.
type Options struct {
	ClientOnly bool
	Short      bool
	Output     string

	client v1.GatewayHTTPClient
	genericclioptions.IOStreams
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdVersion returns a cobra command for fetching versions.
func NewCmdVersion(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the client and server version information",
		Long:    "Print the client and server version information for the current context",
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.ClientOnly, "client", o.ClientOnly, "If true, shows client version only (no server required).")
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")
	cmd.Flags().BoolVar(&o.Short, "short", o.Short, "If true, print just the version number.")

	return cmd
}

// Complete completes all the required options.
func (o *Options) Complete(f cmdutil.Factory, cmd *cobra.Command) error {
	if o.ClientOnly {
		return nil
	}

	// Use gateway
	o.client = f.GatewayClient()

	return nil
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return errors.New(`--output must be 'yaml' or 'json'`)
	}

	return nil
}

// Run executes version command.
func (o *Options) Run() error {
	var (
		serverErr   error
		versionInfo Version
	)

	clientVersion := version.Get()
	versionInfo.ClientVersion = &clientVersion

	if !o.ClientOnly && o.client != nil {
		// Always request fresh data from the server
		vinfo, err := o.client.GetVersion(context.Background(), &emptypb.Empty{})
		if err != nil {
			return err
		}
		versionInfo.ServerVersion = &version.Info{
			GitVersion:   vinfo.GitVersion,
			GitCommit:    vinfo.GitCommit,
			GitTreeState: vinfo.GitTreeState,
			BuildDate:    vinfo.BuildDate,
			GoVersion:    vinfo.GoVersion,
			Compiler:     vinfo.Compiler,
			Platform:     vinfo.Platform,
		}
	}

	switch o.Output {
	case "":
		if o.Short {
			fmt.Fprintf(o.Out, "Client Version: %s\n", clientVersion.GitVersion)

			if versionInfo.ServerVersion != nil {
				fmt.Fprintf(o.Out, "Server Version: %s\n", versionInfo.ServerVersion.GitVersion)
			}
		} else {
			fmt.Fprintf(o.Out, "Client Version: %s\n", fmt.Sprintf("%#v", clientVersion))
			if versionInfo.ServerVersion != nil {
				fmt.Fprintf(o.Out, "Server Version: %s\n", fmt.Sprintf("%#v", *versionInfo.ServerVersion))
			}
		}
	case "yaml":
		marshaled, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	case "json":
		marshaled, err := json.MarshalIndent(&versionInfo, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	default:
		// There is a bug in the program if we hit this case.
		// However, we follow a policy of never panicking.
		return fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", o.Output)
	}

	if versionInfo.ServerVersion != nil {
		if err := printVersionSkewWarning(o.ErrOut, *versionInfo.ClientVersion, *versionInfo.ServerVersion); err != nil {
			return err
		}
	}

	return serverErr
}
