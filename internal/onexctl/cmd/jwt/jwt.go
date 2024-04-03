// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package jwt can be used to sign/show/verify jwt token with given secretID and secretKey.
package jwt

import (
	"github.com/spf13/cobra"

	cmdutil "github.com/superproj/onex/internal/onexctl/cmd/util"
	"github.com/superproj/onex/internal/onexctl/util/templates"
	"github.com/superproj/onex/pkg/cli/genericclioptions"
)

var jwtLong = templates.LongDesc(`
	JWT command.

	This commands is used to sigin/show/verify jwt token.`)

// NewCmdJWT returns new initialized instance of 'jwt' sub command.
func NewCmdJWT(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "jwt SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 "JWT command-line tool",
		Long:                  jwtLong,
		Run:                   cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}

	// add subcommands
	cmd.AddCommand(NewCmdSign(f, ioStreams))
	cmd.AddCommand(NewCmdShow(f, ioStreams))
	cmd.AddCommand(NewCmdVerify(f, ioStreams))

	return cmd
}
