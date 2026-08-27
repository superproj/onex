// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/onexstack/onexstack/pkg/db"
	"github.com/spf13/pflag"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// 帮助信息文本.
const helpText = `Usage: main [flags] arg [arg...]

This is a pflag example.

Flags:
`

// Querier 定义了数据库查询接口.
type Querier interface {
	// FilterWithNameAndRole 按名称和角色查询记录
	FilterWithNameAndRole(name string) ([]gen.T, error)
}

// GenerateConfig 保存代码生成的配置.
type GenerateConfig struct {
	ModelPackagePath string
	GenerateFunc     func(g *gen.Generator)
}

// 预定义的生成配置.
var generateConfigs = map[string]GenerateConfig{
	"uc":    {ModelPackagePath: "internal/usercenter/model", GenerateFunc: GenerateUserCenterModels},
	"gw":    {ModelPackagePath: "internal/gateway/model", GenerateFunc: GenerateGatewayModels},
	"fs":    {ModelPackagePath: "internal/fakeserver/model", GenerateFunc: GenerateFakeServerModels},
	"other": {ModelPackagePath: "internal/gateway/model", GenerateFunc: GenerateOtherModels},
	"nw":    {ModelPackagePath: "internal/nightwatch/model", GenerateFunc: GenerateNightWatchModels},
}

// 命令行参数.
var (
	addr       = pflag.StringP("addr", "a", "127.0.0.1:3306", "MySQL host address.")
	username   = pflag.StringP("username", "u", "onex", "Username to connect to the database.")
	password   = pflag.StringP("password", "p", "onex(#)666", "Password to use when connecting to the database.")
	database   = pflag.StringP("db", "d", "onex", "Database name to connect to.")
	modelPath  = pflag.String("model-pkg-path", "", "Generated model code's package name.")
	components = pflag.StringSlice("component", []string{"uc", "gw", "nw"}, "Generated model code's for specified component.")
	help       = pflag.BoolP("help", "h", false, "Show this help message.")

	usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
)

func main() {
	// 设置自定义的使用说明函数
	pflag.Usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
	pflag.Parse()

	// 如果设置了帮助标志，则显示帮助信息并退出
	if *help {
		pflag.Usage()
		return
	}

	// 初始化数据库连接
	dbInstance, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 处理组件并生成代码
	for _, component := range *components {
		processComponent(component, dbInstance)
	}
}

// initializeDatabase 创建并返回一个数据库连接.
func initializeDatabase() (*gorm.DB, error) {
	dbOptions := &db.MySQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *database,
	}

	// 创建并返回数据库连接
	return db.NewMySQL(dbOptions)
}

// processComponent 处理单个组件以生成代码.
func processComponent(component string, dbInstance *gorm.DB) {
	config, ok := generateConfigs[component]
	if !ok {
		log.Printf("Component '%s' not found in configuration. Skipping.", component)
		return
	}

	// 解析模型包路径
	modelPkgPath := resolveModelPackagePath(config.ModelPackagePath)

	// 创建生成器实例
	generator := createGenerator(modelPkgPath)
	generator.UseDB(dbInstance)

	// 应用自定义生成器选项
	applyGeneratorOptions(generator)

	// 使用指定的函数生成模型
	config.GenerateFunc(generator)

	// 执行代码生成
	generator.Execute()
}

// resolveModelPackagePath 确定模型生成的包路径.
func resolveModelPackagePath(defaultPath string) string {
	if *modelPath != "" {
		return *modelPath
	}

	return filepath.Join(rootDir(), defaultPath)
}

// createGenerator 初始化并返回一个新的生成器实例.
func createGenerator(packagePath string) *gen.Generator {
	return gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		ModelPkgPath:      packagePath,
		WithUnitTest:      true,
		FieldNullable:     true,  // 对于数据库中可空的字段，使用指针类型。
		FieldSignable:     false, // 禁用无符号属性以提高兼容性。
		FieldWithIndexTag: false, // 不包含 GORM 的索引标签。
		FieldWithTypeTag:  false, // 不包含 GORM 的类型标签。
	})
}

// applyGeneratorOptions 设置自定义生成器选项.
func applyGeneratorOptions(g *gen.Generator) {
	// 为特定字段自定义 GORM 标签
	g.WithOpts(
		gen.FieldGORMTag("createdAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
		gen.FieldGORMTag("updatedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
	)
}

func GenerateUserCenterModels(g *gen.Generator) {
	g.GenerateModelAs("usercenter_user", "UserM")
	g.GenerateModelAs("usercenter_secret", "SecretM")
}

func GenerateGatewayModels(g *gen.Generator) {
	g.GenerateModelAs("gateway_chain", "ChainM")
	g.GenerateModelAs("gateway_minerset", "MinerSetM")
	g.GenerateModelAs("gateway_miner", "MinerM")
}

func GenerateFakeServerModels(g *gen.Generator) {
	g.GenerateModelAs("fakeserver_order", "OrderM")
}

func GenerateOtherModels(g *gen.Generator) {
	g.GenerateModelAs("locks", "LocksM")
	g.GenerateModelAs("ratelimit_leader_election", "LeaderElectionM")
}

func GenerateNightWatchModels(g *gen.Generator) {
	g.GenerateModelAs(
		"nightwatch_cronjob",
		"CronJobM",
		gen.FieldRename("cronjob_id", "CronJobID"),
		gen.FieldType("job_template", "*JobM"),
		gen.FieldType("status", "CronJobStatus"),
	)
	g.GenerateModelAs(
		"nightwatch_job",
		"JobM",
		gen.FieldRename("cronjob_id", "CronJobID"),
		gen.FieldType("params", "*JobParams"),
		gen.FieldType("results", "*JobResults"),
		gen.FieldType("conditions", "*JobConditions"),
	)
}

func rootDir() string {
	// 获取当前源文件路径
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Error retrieving file info")
	}

	// 获取当前文件所在的目录
	dir := filepath.Dir(file)

	// 绝对路径
	absPath, err := filepath.Abs(dir + "../../../")
	if err != nil {
		log.Fatalf("Error getting absolute directory path: %v", err)
	}
	return absPath
}
