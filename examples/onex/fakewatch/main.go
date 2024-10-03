package main

import (
	"github.com/superproj/onex/cmd/onex-nightwatch/app"
)

import "fmt"

func main() {
	app.NewJobServer().Run()
}
