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
	deleteUsageStr = "delete SECRET_NAME"
)

// DeleteOptions is an options struct to support delete subcommands.
type DeleteOptions struct {
	Name string

	DeleteSecretRequest *v1.DeleteSecretRequest
	client              v1.UserCenterHTTPClient

	genericclioptions.IOStreams
}

var (
	deleteExample = templates.Examples(`
		# Delete secret foo
		onexctl secret delete foo`)

	deleteUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nSECRET_NAME is required arguments for the delete command",
		deleteUsageStr,
	)
)

// NewDeleteOptions returns an initialized DeleteOptions instance.
func NewDeleteOptions(ioStreams genericclioptions.IOStreams) *DeleteOptions {
	return &DeleteOptions{
		IOStreams: ioStreams,
	}
}

// NewCmdDelete returns new initialized instance of delete sub command.
func NewCmdDelete(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewDeleteOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   deleteUsageStr,
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "Delete a secret resource",
		TraverseChildren:      true,
		Long:                  "Delete a secret resource.",
		Example:               deleteExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f))
		},
		SuggestFor: []string{},
	}

	return cmd
}

// Complete completes all the required options.
func (o *DeleteOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, deleteUsageErrStr)
	}

	o.DeleteSecretRequest = &v1.DeleteSecretRequest{
		Name: args[0],
	}

	o.client = f.UserCenterClient()

	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *DeleteOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.DeleteSecretRequest.Validate()
}

// Run executes a delete subcommand using the specified options.
func (o *DeleteOptions) Run(f cmdutil.Factory) error {
	_, err := o.client.DeleteSecret(context.Background(), o.DeleteSecretRequest)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "secret/%s deleted\n", o.DeleteSecretRequest.Name)

	return nil
}
