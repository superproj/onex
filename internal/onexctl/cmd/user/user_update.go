// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

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
	updateUsageStr = "update USERNAME"
)

// UpdateOptions is an options struct to support update subcommands.
type UpdateOptions struct {
	Name     string
	Nickname string
	Email    string
	Phone    string

	UpdateUserRequest *v1.UpdateUserRequest

	client v1.UserCenterHTTPClient

	genericclioptions.IOStreams
}

var (
	updateLong = templates.LongDesc(`Update a user resource. 

Can only update nickname, email and phone.

NOTICE: field will be updated to zero value if not specified.`)

	updateExample = templates.Examples(`
		# Update use foo's information
		onexctl user update --nickname=foo2 --email=foo@qq.com --phone=1812883xxxx`)

	updateUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nUSERNAME is required arguments for the update command",
		updateUsageStr,
	)
)

// NewUpdateOptions returns an initialized UpdateOptions instance.
func NewUpdateOptions(ioStreams genericclioptions.IOStreams) *UpdateOptions {
	return &UpdateOptions{
		IOStreams: ioStreams,
	}
}

// NewCmdUpdate returns new initialized instance of update sub command.
func NewCmdUpdate(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewUpdateOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   updateUsageStr,
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "Update a user resource",
		TraverseChildren:      true,
		Long:                  updateLong,
		Example:               updateExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f, args))
		},
		SuggestFor: []string{},
	}

	cmd.Flags().StringVar(&o.Nickname, "nickname", o.Nickname, "The nickname of the user.")
	cmd.Flags().StringVar(&o.Email, "email", o.Email, "The email of the user.")
	cmd.Flags().StringVar(&o.Phone, "phone", o.Phone, "The phone number of the user.")

	return cmd
}

// Complete completes all the required options.
func (o *UpdateOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmdutil.UsageErrorf(cmd, updateUsageErrStr)
	}

	o.UpdateUserRequest.Username = args[0]

	if o.Nickname != "" {
		o.UpdateUserRequest.Nickname = &o.Nickname
	}
	if o.Email != "" {
		o.UpdateUserRequest.Email = &o.Email
	}
	if o.Phone != "" {
		o.UpdateUserRequest.Phone = &o.Phone
	}

	o.client = f.UserCenterClient()

	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *UpdateOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.UpdateUserRequest.Validate()
}

// Run executes an update subcommand using the specified options.
func (o *UpdateOptions) Run(f cmdutil.Factory, args []string) error {
	_, err := o.client.UpdateUser(context.Background(), o.UpdateUserRequest)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "user/%s updated\n", o.UpdateUserRequest.Username)

	return nil
}
