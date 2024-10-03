// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"gorm.io/gen"
	"path/filepath"

	"github.com/superproj/onex/pkg/db"
)

const helpText = `Usage: main [flags] arg [arg...]

This is a pflag example.

Flags:
`

type Querier interface {
	// SELECT * FROM @@table WHERE name = @name AND role = @role
	FilterWithNameAndRole(name string) ([]gen.T, error)
}

type GenerateConfig struct {
	OutPath      string
	GenerateFunc func(g *gen.Generator)
}

var generateConfigs = map[string]GenerateConfig{
	"uc": GenerateConfig{
		OutPath:      "../../internal/usercenter/dao/query",
		GenerateFunc: ForUserCenter,
	},
	"api": GenerateConfig{
		OutPath:      "../../internal/gateway/dao/query",
		GenerateFunc: ForGateway,
	},
	"nw": GenerateConfig{
		OutPath:      "../../internal/nightwatch/dao/query",
		GenerateFunc: ForNightWatch,
	},
}

var (
	addr     = pflag.StringP("addr", "a", "127.0.0.1:3306", "MySQL host address.")
	username = pflag.StringP("username", "u", "onex", "Username to connect to the database.")
	password = pflag.StringP("password", "p", "onex(#)666", "Password to use when connecting to the database.")
	dbname   = pflag.StringP("db", "d", "onex", "Database name to connect to.")

	// outPath   = pflag.String("outpath", "./store", "generated gorm query code's path.").
	modelPath  = pflag.String("model-pkg-path", "./model", "Generated model code's package name.")
	components = pflag.StringSlice("component", []string{"uc", "api", "nw"}, "Generated model code's for specified component.")
	help       = pflag.BoolP("help", "h", false, "Show this help message.")

	usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
)

func main() {
	pflag.Usage = usage
	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	dbOptions := &db.MySQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *dbname,
	}

	dbIns, err := db.NewMySQL(dbOptions)
	if err != nil {
		panic(err)
	}

	// if you want to query without context constrain, set mode gen.WithoutContext ###
	fn := func(absPath string) *gen.Generator {
		return gen.NewGenerator(gen.Config{
			// OutPath:      *outPath,
			// OutFile:      filepath.Base(*outPath) + ".go",
			Mode:    gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
			OutPath: absPath,
			//ModelPkgPath: "../../internal/apiserver/dao/model", // 默认情况下会跟随OutPath参数，在同目录下生成
			//ModelPkgPath:      *modelPath,
			WithUnitTest:      true,
			FieldNullable:     true, // 对于数据库表中nullable的数据，在生成代码中自动对应为指针类型
			FieldWithIndexTag: true, // 从数据库同步的表结构代码包含gorm的index tag
			FieldWithTypeTag:  true, // 同步的表结构代码包含gorm的type tag(数据库中对应数据类型)

		})
	}

	// reuse the database connection in Project or create a connection here
	// if you want to use GenerateModel/GenerateModelAs, UseDB is necessary or it will panic

	for _, comp := range *components {
		config, ok := generateConfigs[comp]
		if !ok {
			continue
		}

		abs, _ := filepath.Abs(config.OutPath)

		g := fn(abs)
		g.UseDB(dbIns)
		config.GenerateFunc(g)
		g.Execute()
	}

	// execute the action of code generation
}

func ForUserCenter(g *gen.Generator) {
	g.GenerateModelAs("uc_user", "UserM", gen.FieldIgnore("placeholder"))
	g.GenerateModelAs("uc_secret", "SecretM", gen.FieldIgnore("placeholder"))
}

func ForGateway(g *gen.Generator) {
	g.GenerateModelAs("api_chain", "ChainM", gen.FieldIgnore("placeholder"))
	g.GenerateModelAs("api_minerset", "MinerSetM", gen.FieldIgnore("placeholder"))
	g.GenerateModelAs("api_miner", "MinerM", gen.FieldIgnore("placeholder"))
	// g.ApplyInterface(func(Querier) {}, model.MinerModel{})
}

func ForNightWatch(g *gen.Generator) {
	// 以下GenerateModel获得变量需放入ApplyBasic/ApplyInterface方法才会生效
	cronJob := g.GenerateModelAs(
		"nw_cronjob",
		"CronJobM",
		gen.FieldRename("cronjob_id", "CronJobID"),
		gen.FieldType("job_template", "*JobM"),
		gen.FieldType("status", "*CronJobStatus"),
	)
	job := g.GenerateModelAs(
		"nw_job",
		"JobM",
		gen.FieldRename("cronjob_id", "CronJobID"),
		gen.FieldType("params", "*JobParams"),
		gen.FieldType("results", "*JobResults"),
		gen.FieldType("conditions", "*JobConditions"),
	)
	g.ApplyBasic(cronJob, job)
}
