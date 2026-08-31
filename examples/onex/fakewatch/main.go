//go:build ignore

package main

import (
	"github.com/onexstack/onex/cmd/onex-nightwatch/app"
)

import "fmt"

func main() {
	app.NewJobServer().Run()
}
