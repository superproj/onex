// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/superproj/onex/internal/nightwatch/dao/model"
)

func newCronJobM(db *gorm.DB, opts ...gen.DOOption) cronJobM {
	_cronJobM := cronJobM{}

	_cronJobM.cronJobMDo.UseDB(db, opts...)
	_cronJobM.cronJobMDo.UseModel(&model.CronJobM{})

	tableName := _cronJobM.cronJobMDo.TableName()
	_cronJobM.ALL = field.NewAsterisk(tableName)
	_cronJobM.ID = field.NewInt64(tableName, "id")
	_cronJobM.CronJobID = field.NewString(tableName, "cronjob_id")
	_cronJobM.UserID = field.NewString(tableName, "user_id")
	_cronJobM.Scope = field.NewString(tableName, "scope")
	_cronJobM.Name = field.NewString(tableName, "name")
	_cronJobM.Description = field.NewString(tableName, "description")
	_cronJobM.Schedule = field.NewString(tableName, "schedule")
	_cronJobM.Status = field.NewField(tableName, "status")
	_cronJobM.ConcurrencyPolicy = field.NewInt32(tableName, "concurrency_policy")
	_cronJobM.Suspend = field.NewInt32(tableName, "suspend")
	_cronJobM.JobTemplate = field.NewField(tableName, "job_template")
	_cronJobM.SuccessHistoryLimit = field.NewInt32(tableName, "success_history_limit")
	_cronJobM.FailedHistoryLimit = field.NewInt32(tableName, "failed_history_limit")
	_cronJobM.CreatedAt = field.NewTime(tableName, "created_at")
	_cronJobM.UpdatedAt = field.NewTime(tableName, "updated_at")

	_cronJobM.fillFieldMap()

	return _cronJobM
}

type cronJobM struct {
	cronJobMDo

	ALL                 field.Asterisk
	ID                  field.Int64  // 主键 ID
	CronJobID           field.String // CronJob ID
	UserID              field.String // 创建人
	Scope               field.String // CronJob 作用域
	Name                field.String // CronJob 名称
	Description         field.String // CronJob 描述
	Schedule            field.String // Quartz 格式的调度时间描述
	Status              field.Field  // CronJob 任务状态
	ConcurrencyPolicy   field.Int32  // 作业处理方式（1 串行，2 并行，3 替换）
	Suspend             field.Int32  // 是否挂起（1 挂起，0 不挂起）
	JobTemplate         field.Field  // Job 模版
	SuccessHistoryLimit field.Int32  // 要保留的成功完成作业的数量。值必须是非负整数
	FailedHistoryLimit  field.Int32  // 要保留的失败完成作业的数量。值必须是非负整数。
	CreatedAt           field.Time   // 创建时间
	UpdatedAt           field.Time   // 更新时间

	fieldMap map[string]field.Expr
}

func (c cronJobM) Table(newTableName string) *cronJobM {
	c.cronJobMDo.UseTable(newTableName)
	return c.updateTableName(newTableName)
}

func (c cronJobM) As(alias string) *cronJobM {
	c.cronJobMDo.DO = *(c.cronJobMDo.As(alias).(*gen.DO))
	return c.updateTableName(alias)
}

func (c *cronJobM) updateTableName(table string) *cronJobM {
	c.ALL = field.NewAsterisk(table)
	c.ID = field.NewInt64(table, "id")
	c.CronJobID = field.NewString(table, "cronjob_id")
	c.UserID = field.NewString(table, "user_id")
	c.Scope = field.NewString(table, "scope")
	c.Name = field.NewString(table, "name")
	c.Description = field.NewString(table, "description")
	c.Schedule = field.NewString(table, "schedule")
	c.Status = field.NewField(table, "status")
	c.ConcurrencyPolicy = field.NewInt32(table, "concurrency_policy")
	c.Suspend = field.NewInt32(table, "suspend")
	c.JobTemplate = field.NewField(table, "job_template")
	c.SuccessHistoryLimit = field.NewInt32(table, "success_history_limit")
	c.FailedHistoryLimit = field.NewInt32(table, "failed_history_limit")
	c.CreatedAt = field.NewTime(table, "created_at")
	c.UpdatedAt = field.NewTime(table, "updated_at")

	c.fillFieldMap()

	return c
}

func (c *cronJobM) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := c.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (c *cronJobM) fillFieldMap() {
	c.fieldMap = make(map[string]field.Expr, 15)
	c.fieldMap["id"] = c.ID
	c.fieldMap["cronjob_id"] = c.CronJobID
	c.fieldMap["user_id"] = c.UserID
	c.fieldMap["scope"] = c.Scope
	c.fieldMap["name"] = c.Name
	c.fieldMap["description"] = c.Description
	c.fieldMap["schedule"] = c.Schedule
	c.fieldMap["status"] = c.Status
	c.fieldMap["concurrency_policy"] = c.ConcurrencyPolicy
	c.fieldMap["suspend"] = c.Suspend
	c.fieldMap["job_template"] = c.JobTemplate
	c.fieldMap["success_history_limit"] = c.SuccessHistoryLimit
	c.fieldMap["failed_history_limit"] = c.FailedHistoryLimit
	c.fieldMap["created_at"] = c.CreatedAt
	c.fieldMap["updated_at"] = c.UpdatedAt
}

func (c cronJobM) clone(db *gorm.DB) cronJobM {
	c.cronJobMDo.ReplaceConnPool(db.Statement.ConnPool)
	return c
}

func (c cronJobM) replaceDB(db *gorm.DB) cronJobM {
	c.cronJobMDo.ReplaceDB(db)
	return c
}

type cronJobMDo struct{ gen.DO }

type ICronJobMDo interface {
	gen.SubQuery
	Debug() ICronJobMDo
	WithContext(ctx context.Context) ICronJobMDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ICronJobMDo
	WriteDB() ICronJobMDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ICronJobMDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ICronJobMDo
	Not(conds ...gen.Condition) ICronJobMDo
	Or(conds ...gen.Condition) ICronJobMDo
	Select(conds ...field.Expr) ICronJobMDo
	Where(conds ...gen.Condition) ICronJobMDo
	Order(conds ...field.Expr) ICronJobMDo
	Distinct(cols ...field.Expr) ICronJobMDo
	Omit(cols ...field.Expr) ICronJobMDo
	Join(table schema.Tabler, on ...field.Expr) ICronJobMDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ICronJobMDo
	RightJoin(table schema.Tabler, on ...field.Expr) ICronJobMDo
	Group(cols ...field.Expr) ICronJobMDo
	Having(conds ...gen.Condition) ICronJobMDo
	Limit(limit int) ICronJobMDo
	Offset(offset int) ICronJobMDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ICronJobMDo
	Unscoped() ICronJobMDo
	Create(values ...*model.CronJobM) error
	CreateInBatches(values []*model.CronJobM, batchSize int) error
	Save(values ...*model.CronJobM) error
	First() (*model.CronJobM, error)
	Take() (*model.CronJobM, error)
	Last() (*model.CronJobM, error)
	Find() ([]*model.CronJobM, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.CronJobM, err error)
	FindInBatches(result *[]*model.CronJobM, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.CronJobM) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ICronJobMDo
	Assign(attrs ...field.AssignExpr) ICronJobMDo
	Joins(fields ...field.RelationField) ICronJobMDo
	Preload(fields ...field.RelationField) ICronJobMDo
	FirstOrInit() (*model.CronJobM, error)
	FirstOrCreate() (*model.CronJobM, error)
	FindByPage(offset int, limit int) (result []*model.CronJobM, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ICronJobMDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (c cronJobMDo) Debug() ICronJobMDo {
	return c.withDO(c.DO.Debug())
}

func (c cronJobMDo) WithContext(ctx context.Context) ICronJobMDo {
	return c.withDO(c.DO.WithContext(ctx))
}

func (c cronJobMDo) ReadDB() ICronJobMDo {
	return c.Clauses(dbresolver.Read)
}

func (c cronJobMDo) WriteDB() ICronJobMDo {
	return c.Clauses(dbresolver.Write)
}

func (c cronJobMDo) Session(config *gorm.Session) ICronJobMDo {
	return c.withDO(c.DO.Session(config))
}

func (c cronJobMDo) Clauses(conds ...clause.Expression) ICronJobMDo {
	return c.withDO(c.DO.Clauses(conds...))
}

func (c cronJobMDo) Returning(value interface{}, columns ...string) ICronJobMDo {
	return c.withDO(c.DO.Returning(value, columns...))
}

func (c cronJobMDo) Not(conds ...gen.Condition) ICronJobMDo {
	return c.withDO(c.DO.Not(conds...))
}

func (c cronJobMDo) Or(conds ...gen.Condition) ICronJobMDo {
	return c.withDO(c.DO.Or(conds...))
}

func (c cronJobMDo) Select(conds ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.Select(conds...))
}

func (c cronJobMDo) Where(conds ...gen.Condition) ICronJobMDo {
	return c.withDO(c.DO.Where(conds...))
}

func (c cronJobMDo) Order(conds ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.Order(conds...))
}

func (c cronJobMDo) Distinct(cols ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.Distinct(cols...))
}

func (c cronJobMDo) Omit(cols ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.Omit(cols...))
}

func (c cronJobMDo) Join(table schema.Tabler, on ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.Join(table, on...))
}

func (c cronJobMDo) LeftJoin(table schema.Tabler, on ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.LeftJoin(table, on...))
}

func (c cronJobMDo) RightJoin(table schema.Tabler, on ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.RightJoin(table, on...))
}

func (c cronJobMDo) Group(cols ...field.Expr) ICronJobMDo {
	return c.withDO(c.DO.Group(cols...))
}

func (c cronJobMDo) Having(conds ...gen.Condition) ICronJobMDo {
	return c.withDO(c.DO.Having(conds...))
}

func (c cronJobMDo) Limit(limit int) ICronJobMDo {
	return c.withDO(c.DO.Limit(limit))
}

func (c cronJobMDo) Offset(offset int) ICronJobMDo {
	return c.withDO(c.DO.Offset(offset))
}

func (c cronJobMDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ICronJobMDo {
	return c.withDO(c.DO.Scopes(funcs...))
}

func (c cronJobMDo) Unscoped() ICronJobMDo {
	return c.withDO(c.DO.Unscoped())
}

func (c cronJobMDo) Create(values ...*model.CronJobM) error {
	if len(values) == 0 {
		return nil
	}
	return c.DO.Create(values)
}

func (c cronJobMDo) CreateInBatches(values []*model.CronJobM, batchSize int) error {
	return c.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (c cronJobMDo) Save(values ...*model.CronJobM) error {
	if len(values) == 0 {
		return nil
	}
	return c.DO.Save(values)
}

func (c cronJobMDo) First() (*model.CronJobM, error) {
	if result, err := c.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.CronJobM), nil
	}
}

func (c cronJobMDo) Take() (*model.CronJobM, error) {
	if result, err := c.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.CronJobM), nil
	}
}

func (c cronJobMDo) Last() (*model.CronJobM, error) {
	if result, err := c.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.CronJobM), nil
	}
}

func (c cronJobMDo) Find() ([]*model.CronJobM, error) {
	result, err := c.DO.Find()
	return result.([]*model.CronJobM), err
}

func (c cronJobMDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.CronJobM, err error) {
	buf := make([]*model.CronJobM, 0, batchSize)
	err = c.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (c cronJobMDo) FindInBatches(result *[]*model.CronJobM, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return c.DO.FindInBatches(result, batchSize, fc)
}

func (c cronJobMDo) Attrs(attrs ...field.AssignExpr) ICronJobMDo {
	return c.withDO(c.DO.Attrs(attrs...))
}

func (c cronJobMDo) Assign(attrs ...field.AssignExpr) ICronJobMDo {
	return c.withDO(c.DO.Assign(attrs...))
}

func (c cronJobMDo) Joins(fields ...field.RelationField) ICronJobMDo {
	for _, _f := range fields {
		c = *c.withDO(c.DO.Joins(_f))
	}
	return &c
}

func (c cronJobMDo) Preload(fields ...field.RelationField) ICronJobMDo {
	for _, _f := range fields {
		c = *c.withDO(c.DO.Preload(_f))
	}
	return &c
}

func (c cronJobMDo) FirstOrInit() (*model.CronJobM, error) {
	if result, err := c.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.CronJobM), nil
	}
}

func (c cronJobMDo) FirstOrCreate() (*model.CronJobM, error) {
	if result, err := c.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.CronJobM), nil
	}
}

func (c cronJobMDo) FindByPage(offset int, limit int) (result []*model.CronJobM, count int64, err error) {
	result, err = c.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = c.Offset(-1).Limit(-1).Count()
	return
}

func (c cronJobMDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = c.Count()
	if err != nil {
		return
	}

	err = c.Offset(offset).Limit(limit).Scan(result)
	return
}

func (c cronJobMDo) Scan(result interface{}) (err error) {
	return c.DO.Scan(result)
}

func (c cronJobMDo) Delete(models ...*model.CronJobM) (result gen.ResultInfo, err error) {
	return c.DO.Delete(models)
}

func (c *cronJobMDo) withDO(do gen.Dao) *cronJobMDo {
	c.DO = *do.(*gen.DO)
	return c
}
