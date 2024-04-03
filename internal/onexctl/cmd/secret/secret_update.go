// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

const (
	updateUsageStr = "update SECRET_NAME"
)

// UpdateOptions is an options struct to support update subcommands.
type UpdateOptions struct {
	Description string
	Expires     int64
	Status      int32

	UpdateSecretRequest *v1.UpdateSecretRequest
	client              v1.UserCenterHTTPClient

	genericclioptions.IOStreams
}

var (
	updateExample = templates.Examples(`
		# Update a secret resource
		onexctl secret update foo --expires=4h --description="new description"`)

	updateUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nSECRET_NAME is required arguments for the update command",
		updateUsageStr,
	)
)

// NewUpdateOptions returns an initialized UpdateOptions instance.
func NewUpdateOptions(ioStreams genericclioptions.IOStreams) *UpdateOptions {
	return &UpdateOptions{
		Expires:   -1,
		Status:    -1,
		IOStreams: ioStreams,
	}
}

// NewCmdUpdate returns new initialized instance of update sub command.
func NewCmdUpdate(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewUpdateOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "update SECRET_NAME",
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "Update a secret resource",
		TraverseChildren:      true,
		Long:                  "Update a secret resource.",
		Example:               updateExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f, args))
		},
		SuggestFor: []string{},
	}

	cmd.Flags().StringVar(&o.Description, "description", o.Description, "The description of the secret.")
	cmd.Flags().Int32Var(&o.Status, "status", o.Status, "The status of the secret.")
	cmd.Flags().Int64Var(&o.Expires, "expires", o.Expires, "The expires of the secret.")

	return cmd
}

// Complete completes all the required options.
func (o *UpdateOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, updateUsageErrStr)
	}

	o.UpdateSecretRequest = &v1.UpdateSecretRequest{
		Name: args[0],
	}

	if o.Expires != -1 {
		o.UpdateSecretRequest.Expires = &o.Expires
	}
	if o.Status != -1 {
		o.UpdateSecretRequest.Status = &o.Status
	}
	if o.Description != "" {
		o.UpdateSecretRequest.Description = &o.Description
	}

	o.client = f.UserCenterClient()

	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *UpdateOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.UpdateSecretRequest.Validate()
}

// Run executes a update subcommand using the specified options.
func (o *UpdateOptions) Run(f cmdutil.Factory, args []string) error {
	_, err := o.client.UpdateSecret(context.Background(), o.UpdateSecretRequest)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "secret/%s updated\n", o.UpdateSecretRequest.Name)

	return nil
}
