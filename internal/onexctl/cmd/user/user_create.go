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
	createUsageStr = "create USERNAME PASSWORD EMAIL"
)

// CreateOptions is an options struct to support create subcommands.
type CreateOptions struct {
	Email    string
	Nickname string
	Phone    string

	CreateUserRequest *v1.CreateUserRequest

	client v1.UserCenterHTTPClient
	genericclioptions.IOStreams
}

var (
	createLong = templates.LongDesc(`Create a user on onex cloud platform.
If nickname not specified, username will be used.`)

	createExample = templates.Examples(`
		# Create user with given input
		onexctl user create foo Foo@2023

		# Create user wt 
		onexctl user create foo Foo@2023 --email=foo@foxmail.com --phone=18128845xxx --nickname=colin`)

	createUsageErrStr = fmt.Sprintf(
		"expected '%s'.\nUSERNAME PASSWORD are required arguments for the create command",
		createUsageStr,
	)
)

// NewCreateOptions returns an initialized CreateOptions instance.
func NewCreateOptions(ioStreams genericclioptions.IOStreams) *CreateOptions {
	return &CreateOptions{
		Email:     "colin@onex.com",
		Phone:     "1812884xxxx",
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
		Short:                 "Create a user resource",
		TraverseChildren:      true,
		Long:                  createLong,
		Example:               createExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
		SuggestFor: []string{},
	}

	// mark flag as deprecated
	cmd.Flags().StringVar(&o.Email, "email", o.Email, "The email of the user.")
	cmd.Flags().StringVar(&o.Nickname, "nickname", o.Nickname, "The nickname of the user.")
	cmd.Flags().StringVar(&o.Phone, "phone", o.Phone, "The phone number of the user.")

	return cmd
}

// Complete completes all the required options.
func (o *CreateOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return cmdutil.UsageErrorf(cmd, createUsageErrStr)
	}

	if o.Nickname == "" {
		o.Nickname = args[0]
	}

	o.CreateUserRequest = &v1.CreateUserRequest{
		Username: args[0],
		Nickname: o.Nickname,
		Password: args[1],
		Email:    o.Email,
		Phone:    o.Phone,
	}

	o.client = f.UserCenterClient()
	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *CreateOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.CreateUserRequest.Validate()
}

// Run executes a create subcommand using the specified options.
func (o *CreateOptions) Run(args []string) error {
	_, err := o.client.CreateUser(context.Background(), o.CreateUserRequest)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "user/%s created\n", o.CreateUserRequest.Username)

	return nil
}
