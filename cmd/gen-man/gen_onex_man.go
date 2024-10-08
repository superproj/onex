// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:funlen,gocritic
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	mangen "github.com/cpuguy83/go-md2man/v2/md2man"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/cmd/genutils"

	apiservapp "github.com/superproj/onex/cmd/onex-apiserver/app"
	ctrlmgrapp "github.com/superproj/onex/cmd/onex-controller-manager/app"
	fakeserverapp "github.com/superproj/onex/cmd/onex-fakeserver/app"
	gwapp "github.com/superproj/onex/cmd/onex-gateway/app"
	minerctrlapp "github.com/superproj/onex/cmd/onex-miner-controller/app"
	minersetctrlapp "github.com/superproj/onex/cmd/onex-minerset-controller/app"
	nwapp "github.com/superproj/onex/cmd/onex-nightwatch/app"
	pumpapp "github.com/superproj/onex/cmd/onex-pump/app"
	toyblcapp "github.com/superproj/onex/cmd/onex-toyblc/app"
	usercenterapp "github.com/superproj/onex/cmd/onex-usercenter/app"
	onexctlcmd "github.com/superproj/onex/internal/onexctl/cmd"
)

func main() {
	// use os.Args instead of "flags" because "flags" will mess up the man pages!
	path := "docs/man/man1"
	module := ""
	if len(os.Args) == 3 {
		path = os.Args[1]
		module = os.Args[2]
	} else {
		fmt.Fprintf(os.Stderr, "usage: %s [output directory] [module] \n", os.Args[0])
		os.Exit(1)
	}

	outDir, err := genutils.OutDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get output directory: %v\n", err)
		os.Exit(1)
	}

	// Set environment variables used by command so the output is consistent,
	// regardless of where we run.
	_ = os.Setenv("HOME", "/home/username")

	switch module {
	case "onex-fakeserver":
		// generate manpage for onexfakeserver-
		fakeserver := fakeserverapp.NewApp().Command()
		genMarkdown(fakeserver, "", outDir)
		for _, c := range fakeserver.Commands() {
			genMarkdown(c, "onex-fakeserver", outDir)
		}
	case "onex-usercenter":
		// generate manpage for onex-usercenter
		usercenter := usercenterapp.NewApp().Command()
		genMarkdown(usercenter, "", outDir)
		for _, c := range usercenter.Commands() {
			genMarkdown(c, "onex-usercenter", outDir)
		}
	case "onex-apiserver":
		// generate manpage for onex-apiserver
		apiserver := apiservapp.NewAPIServerCommand()
		genMarkdown(apiserver, "", outDir)
		for _, c := range apiserver.Commands() {
			genMarkdown(c, "onex-apiserver", outDir)
		}
	case "onex-gateway":
		// generate manpage for onex-gateway
		gwserver := gwapp.NewApp().Command()
		genMarkdown(gwserver, "", outDir)
		for _, c := range gwserver.Commands() {
			genMarkdown(c, "onex-gateway", outDir)
		}
	case "onex-nightwatch":
		// generate manpage for onex-nightwatch
		nw := nwapp.NewApp("onex-nightwatch").Command()
		genMarkdown(nw, "", outDir)
		for _, c := range nw.Commands() {
			genMarkdown(c, "onex-nightwatch", outDir)
		}
	case "onex-pump":
		// generate manpage for onex-pump
		pump := pumpapp.NewApp().Command()
		genMarkdown(pump, "", outDir)
		for _, c := range pump.Commands() {
			genMarkdown(c, "onex-pump", outDir)
		}
	case "onex-toyblc":
		// generate manpage for onex-toyblc
		toyblc := toyblcapp.NewApp().Command()
		genMarkdown(toyblc, "", outDir)
		for _, c := range toyblc.Commands() {
			genMarkdown(c, "onex-toyblc", outDir)
		}
	case "onex-controller-manager":
		// generate manpage for onex-controller-manager
		ctrlmgr := ctrlmgrapp.NewControllerManagerCommand()
		genMarkdown(ctrlmgr, "", outDir)
		for _, c := range ctrlmgr.Commands() {
			genMarkdown(c, "onex-controller-manager", outDir)
		}
	case "onex-minerset-controller":
		// generate manpage for onex-minerset-controller
		minersetctrl := minersetctrlapp.NewControllerCommand()
		genMarkdown(minersetctrl, "", outDir)
		for _, c := range minersetctrl.Commands() {
			genMarkdown(c, "onex-minerset-controller", outDir)
		}
	case "onex-miner-controller":
		// generate manpage for onex-miner-controller
		minerctrl := minerctrlapp.NewControllerCommand()
		genMarkdown(minerctrl, "", outDir)
		for _, c := range minerctrl.Commands() {
			genMarkdown(c, "onex-miner-controller", outDir)
		}
	case "onexctl":
		// generate manpage for onexctl
		// TODO os.Stdin should really be something like ioutil.Discard, but a Reader
		onexctl := onexctlcmd.NewDefaultOneXCtlCommand()
		genMarkdown(onexctl, "", outDir)
		for _, c := range onexctl.Commands() {
			genMarkdown(c, "onexctl", outDir)
		}
	default:
		fmt.Fprintf(os.Stderr, "Module %s is not supported", module)
		os.Exit(1)
	}
}

func preamble(out *bytes.Buffer, name, short, long string) {
	out.WriteString(`% OneX(1) onex User Manuals
% Eric Paris
% Jan 2015
# NAME
`)
	fmt.Fprintf(out, "%s \\- %s\n\n", name, short)
	fmt.Fprintf(out, "# SYNOPSIS\n")
	fmt.Fprintf(out, "**%s** [OPTIONS]\n\n", name)
	fmt.Fprintf(out, "# DESCRIPTION\n")
	fmt.Fprintf(out, "%s\n\n", long)
}

func printFlags(out io.Writer, flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		format := "**--%s**=%s\n\t%s\n\n"
		if flag.Value.Type() == "string" {
			// put quotes on the value
			format = "**--%s**=%q\n\t%s\n\n"
		}

		// Todo, when we mark a shorthand is deprecated, but specify an empty message.
		// The flag.ShorthandDeprecated is empty as the shorthand is deprecated.
		// Using len(flag.ShorthandDeprecated) > 0 can't handle this, others are ok.
		if !(len(flag.ShorthandDeprecated) > 0) && len(flag.Shorthand) > 0 {
			format = "**-%s**, " + format
			fmt.Fprintf(out, format, flag.Shorthand, flag.Name, flag.DefValue, flag.Usage)
		} else {
			fmt.Fprintf(out, format, flag.Name, flag.DefValue, flag.Usage)
		}
	})
}

func printOptions(out io.Writer, command *cobra.Command) {
	flags := command.NonInheritedFlags()
	if flags.HasFlags() {
		fmt.Fprintf(out, "# OPTIONS\n")
		printFlags(out, flags)
		fmt.Fprintf(out, "\n")
	}
	flags = command.InheritedFlags()
	if flags.HasFlags() {
		fmt.Fprintf(out, "# OPTIONS INHERITED FROM PARENT COMMANDS\n")
		printFlags(out, flags)
		fmt.Fprintf(out, "\n")
	}
}

func genMarkdown(command *cobra.Command, parent, docsDir string) {
	dparent := strings.ReplaceAll(parent, " ", "-")
	name := command.Name()

	dname := name
	if len(parent) > 0 {
		dname = dparent + "-" + name
		name = parent + " " + name
	}

	out := new(bytes.Buffer)

	short, long := command.Short, command.Long
	if len(long) == 0 {
		long = short
	}

	preamble(out, name, short, long)
	printOptions(out, command)

	if len(command.Example) > 0 {
		fmt.Fprintf(out, "# EXAMPLE\n")
		fmt.Fprintf(out, "```\n%s\n```\n", command.Example)
	}

	if len(command.Commands()) > 0 || len(parent) > 0 {
		fmt.Fprintf(out, "# SEE ALSO\n")

		if len(parent) > 0 {
			fmt.Fprintf(out, "**%s(1)**, ", dparent)
		}

		for _, c := range command.Commands() {
			fmt.Fprintf(out, "**%s-%s(1)**, ", dname, c.Name())
			genMarkdown(c, name, docsDir)
		}

		fmt.Fprintf(out, "\n")
	}

	out.WriteString(`
# HISTORY
January 2015, Originally compiled by Eric Paris (eparis at redhat dot com) based on the superproj source material, but hopefully they have been automatically generated since!
`)

	final := mangen.Render(out.Bytes())

	filename := docsDir + dname + ".1"

	outFile, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()

	_, err = outFile.Write(final)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
