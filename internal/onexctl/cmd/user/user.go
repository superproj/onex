// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package user provides functions to manage users on onex cloud platform.
package user

import (
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

var userLong = templates.LongDesc(`
	User management commands.

Administrator can use all subcommands, non-administrator only allow to use create/get/upate. When call get/update non-administrator 
only allow to operate their own resources, if permission not allowed, will return an 'Permission denied' error.`)

// NewCmdUser returns new initialized instance of 'user' sub command.
func NewCmdUser(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "user SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 "Manage users on onex cloud platform",
		Long:                  userLong,
		Run:                   cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}

	cmd.AddCommand(NewCmdCreate(f, ioStreams))
	cmd.AddCommand(NewCmdList(f, ioStreams))
	cmd.AddCommand(NewCmdGet(f, ioStreams))
	cmd.AddCommand(NewCmdDelete(f, ioStreams))
	cmd.AddCommand(NewCmdUpdate(f, ioStreams))

	return cmd
}

// setHeader set headers for user commands.
func setHeader(table *tablewriter.Table) *tablewriter.Table {
	table.SetHeader([]string{"Name", "UserID", "Nickname", "Email", "Phone", "Created", "Updated"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgWhiteColor})

	return table
}
