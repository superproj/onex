// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package minerset provides functions to manage minersets on onex platform.
package minerset

import (
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

var minersetLong = templates.LongDesc(`
	MinerSet management commands.

	This commands allow you to manage your minerset on onex platform.`)

// NewCmdMinerSet returns new initialized instance of 'minerset' sub command.
func NewCmdMinerSet(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "minerset SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 "Manage minersets on onex platform",
		Long:                  minersetLong,
		Run:                   cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}

	// cmd.AddCommand(NewCmdGet(f, ioStreams))
	cmd.AddCommand(NewCmdList(f, ioStreams))

	return cmd
}

// setHeader set headers for minerset commands.
func setHeader(table *tablewriter.Table) *tablewriter.Table {
	table.SetHeader([]string{"Name", "Replicas", "DisplayName", "Created"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgRedColor},
		tablewriter.Colors{tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
	)

	return table
}
