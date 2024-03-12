// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package validate is used to validate the basic environment for onexctl to run.
package validate

import (
	"net"
	"net/url"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/superproj/onex/internal/onexctl"
	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

var (
	StatusPass = color.GreenString("Pass")
	StatusFail = color.RedString("Fail")
)

// ValidateOptions is an options struct to support 'validate' sub command.
type ValidateOptions struct {
	client v1.UserCenterHTTPClient
	genericclioptions.IOStreams
}

// ValidateInfo defines the validate information.
type ValidateInfo struct {
	Check   string
	Status  string
	Message string
}

func (vi ValidateInfo) Data() []string {
	return []string{vi.Check, vi.Status, vi.Message}
}

func NewPassValidateInfo(check string) ValidateInfo {
	return ValidateInfo{
		Check:   check,
		Status:  StatusPass,
		Message: "",
	}
}

var validateExample = templates.Examples(`
		# Validate the basic environment for onexctl to run
		onexctl validate`)

// NewValidateOptions returns an initialized ValidateOptions instance.
func NewValidateOptions(ioStreams genericclioptions.IOStreams) *ValidateOptions {
	return &ValidateOptions{
		IOStreams: ioStreams,
	}
}

// NewCmdValidate returns new initialized instance of 'validate' sub command.
func NewCmdValidate(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewValidateOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "validate",
		DisableFlagsInUseLine: true,
		Aliases:               []string{},
		Short:                 "Validate the basic environment for onexctl to run",
		TraverseChildren:      true,
		Long:                  "Validate the basic environment for onexctl to run.",
		Example:               validateExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(f, args))
		},
		SuggestFor: []string{},
	}

	return cmd
}

// Complete completes all the required options.
func (o *ValidateOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.client = f.UserCenterClient()
	return nil
}

// Validate makes sure there is no discrepency in command options.
func (o *ValidateOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Run executes a validate sub command using the specified options.
func (o *ValidateOptions) Run(f cmdutil.Factory, args []string) error {
	data := [][]string{}

	// 1. check usercenter connection status
	vinfo := NewPassValidateInfo("UserCenterConnectionStatus")
	if err := checkUserCenterConnectionStatus(f.GetOptions().UserCenterOptions.Addr); err != nil {
		vinfo.Status = StatusFail
		vinfo.Message = err.Error()
	}
	data = append(data, vinfo.Data())

	// 2. check login permission
	vinfo = NewPassValidateInfo("LoginPermission")
	if _, err := f.Login(); err != nil {
		vinfo.Status = StatusFail
		vinfo.Message = err.Error()
	}
	data = append(data, vinfo.Data())

	// 3. check gateway connection status
	vinfo = NewPassValidateInfo("GatewayConnectionStatus")
	data = append(data, vinfo.Data())

	// vinfo := NewPassValidateInfo("Gateway2UserCenterStatus")

	table := tablewriter.NewWriter(o.Out)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColWidth(onexctl.TableWidth)
	table.SetHeader([]string{"Check", "Status", "Error"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()

	return nil
}

func checkUserCenterConnectionStatus(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	_, err = net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	return nil
}
