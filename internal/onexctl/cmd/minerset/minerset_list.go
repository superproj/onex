// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package minerset

import (
	"context"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

const (
	defaltLimit = 1000
)

// ListOptions is an options struct to support list subcommands.
type ListOptions struct {
	Offset int64
	Limit  int64

	ListMinerSetRequest *v1.ListMinerSetRequest
	client              v1.GatewayHTTPClient
	genericclioptions.IOStreams
}

var listExample = templates.Examples(`
		# List all minersets
		onexctl minerset list

		# List minersets with limit and offset 
		onexctl minerset list --offset=0 --limit=5`)

// NewListOptions returns an initialized ListOptions instance.
func NewListOptions(ioStreams genericclioptions.IOStreams) *ListOptions {
	return &ListOptions{
		IOStreams: ioStreams,
		Offset:    0,
		Limit:     defaltLimit,
	}
}

// NewCmdList returns new initialized instance of list sub command.
func NewCmdList(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewListOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "list",
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "Display all minerset resources",
		TraverseChildren:      true,
		Long:                  "Display all minerset resources.",
		Example:               listExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f, args))
		},
		SuggestFor: []string{},
	}

	cmd.Flags().Int64VarP(&o.Offset, "offset", "o", o.Offset, "Specify the offset of the first row to be returned.")
	cmd.Flags().Int64VarP(&o.Limit, "limit", "l", o.Limit, "Specify the amount records to be returned.")

	return cmd
}

// Complete completes all the required options.
func (o *ListOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.ListMinerSetRequest = &v1.ListMinerSetRequest{
		Limit:  o.Limit,
		Offset: o.Offset,
	}
	o.client = f.GatewayClient()

	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *ListOptions) Validate(cmd *cobra.Command, args []string) error {
	return o.ListMinerSetRequest.Validate()
}

// Run executes a list subcommand using the specified options.
func (o *ListOptions) Run(f cmdutil.Factory, args []string) error {
	minersets, err := o.client.ListMinerSet(context.Background(), o.ListMinerSetRequest)
	if err != nil {
		return err
	}

	data := make([][]string, 0, len(minersets.MinerSets))
	table := tablewriter.NewWriter(o.Out)

	for _, minerset := range minersets.MinerSets {
		data = append(data, []string{
			minerset.Name,
			strconv.FormatInt(int64(minerset.Replicas), 10),
			minerset.DisplayName,
			minerset.CreatedAt.AsTime().Format(time.DateTime),
		})
	}

	table = setHeader(table)
	table = cmdutil.TableWriterDefaultConfig(table)
	table.AppendBulk(data)
	table.Render()

	return nil
}
