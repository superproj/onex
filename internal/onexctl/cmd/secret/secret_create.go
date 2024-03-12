// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

const (
	createUsageStr = "create SECRET_NAME"
)

// CreateOptions is an options struct to support create subcommands.
type CreateOptions struct {
	Description string
	Expires     int64

	CreateSecretRequest *v1.CreateSecretRequest
	client              v1.UserCenterHTTPClient

	genericclioptions.IOStreams
}

var (
	createLong = templates.LongDesc(`Create secret resource.

This will generate secretID and secretKey which can be used to sign JWT token.`)

	createExample = templates.Examples(`
		# Create secret which will expired after 2 hours
		onexctl secret create foo

		# Create secret with a specified expire time and description
		onexctl secret create foo --expires=1988121600 --description="secret for onex"`)

	createUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nSECRET_NAME is required arguments for the create command",
		createUsageStr,
	)
)

// NewCreateOptions returns an initialized CreateOptions instance.
func NewCreateOptions(ioStreams genericclioptions.IOStreams) *CreateOptions {
	return &CreateOptions{
		Expires:   time.Now().Add(144 * time.Hour).Unix(),
		IOStreams: ioStreams,
	}
}

// NewCmdCreate returns new initialized instance of create sub command.
func NewCmdCreate(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreateOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   createUsageStr,
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "Create secret resource",
		TraverseChildren:      true,
		Long:                  createLong,
		Example:               createExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f, args))
		},
		SuggestFor: []string{},
	}

	cmd.Flags().StringVar(&o.Description, "description", o.Description, "The descriptin of the secret.")
	cmd.Flags().Int64Var(&o.Expires, "expires", o.Expires, "The expire time of the secret.")

	return cmd
}

// Complete completes all the required options.
func (o *CreateOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, createUsageErrStr)
	}

	o.CreateSecretRequest = &v1.CreateSecretRequest{
		Name:        args[0],
		Expires:     o.Expires,
		Description: o.Description,
	}

	o.client = f.UserCenterClient()
	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *CreateOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.CreateSecretRequest.Validate()
}

// Run executes a create subcommand using the specified options.
func (o *CreateOptions) Run(f cmdutil.Factory, args []string) error {
	_, err := o.client.CreateSecret(context.Background(), o.CreateSecretRequest)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "secret/%s created\n", o.CreateSecretRequest.Name)

	return nil
}
