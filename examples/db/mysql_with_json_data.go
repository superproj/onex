package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/superproj/onex/internal/pkg/meta"
	"github.com/superproj/onex/pkg/db"
	"gorm.io/gorm"
)

const TableNameQagentsDataset = "qagents_dataset"

var ErrDatasetSamplesInvalidType = errors.New("invalid type for DatasetSamples")
var ErrTagsInvalidType = errors.New("invalid type for Tags")

// QagentsDataset mapped from table <qagents_dataset>
type Dataset struct {
	ID            int64          `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:true;comment:数据库 ID" json:"id"`              // 数据库 ID
	Name          string         `gorm:"column:name;type:varchar(255);not null;comment:样本集名称" json:"name"`                                   // 样本集名称
	Type          string         `gorm:"column:type;type:varchar(50);not null;comment:样本集类型" json:"type"`                                    // 样本集类型
	Tenant        string         `gorm:"column:tenant;type:varchar(256);not null;index:idx_tenant,priority:1;comment:业务线" json:"tenant"`     // 业务线
	Author        string         `gorm:"column:author;type:varchar(100);not null;comment:创建人名称" json:"author"`                               // 创建人名称
	Description   string         `gorm:"column:description;type:varchar(256);not null;comment:样本集描述" json:"description"`                     // 样本集描述
	Tags          Tags           `gorm:"column:tags;type:json;comment:标签" json:"tags"`                                                       // 标签
	Visibility    int32          `gorm:"column:visibility;type:int;not null;comment:可见性" json:"visibility"`                                  // 可见性
	SecurityLevel int32          `gorm:"column:security_level;type:int;not null;comment:安全等级" json:"security_level"`                         // 安全等级
	Samples       DatasetSamples `gorm:"column:samples;type:json;comment:样本集数据内容" json:"samples"`                                            // 样本集数据内容
	CreatedAt     time.Time      `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
}

// TableName QagentsDataset's table name
func (*Dataset) TableName() string {
	return TableNameQagentsDataset
}

type DatasetSample struct {
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type DatasetSamples []DatasetSample
type Tags []string

// Scan implements the sql Scanner interface
func (d *DatasetSamples) Scan(value interface{}) error {
	if value == nil {
		*d = DatasetSamples{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrDatasetSamplesInvalidType
	}

	return json.Unmarshal(bytes, d)
}

// Value implements the sql Valuer interface
func (d DatasetSamples) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Value implements the sql Valuer interface
func (t Tags) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements the sql Scanner interface
func (t *Tags) Scan(value interface{}) error {
	if value == nil {
		*t = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrTagsInvalidType
	}

	return json.Unmarshal(bytes, t)
}

func main() {
	optss := &db.MySQLOptions{
		Addr:     "10.37.43.62:3306",
		Username: "onex",
		Password: "onex(#)666",
		Database: "experience",
	}

	db, err := db.NewMySQL(optss)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	dataset := &Dataset{
		Name:          "aa",
		Type:          "bb",
		Tenant:        "cc",
		Author:        "cc",
		Description:   "111",
		Tags:          Tags{"11", "22"},
		Visibility:    3,
		SecurityLevel: 4,
		Samples: DatasetSamples{
			DatasetSample{
				Type:    "11",
				Content: "aaa",
			},
		},
	}

	// 插入数据
	result := db.Create(dataset)
	if result.Error != nil {
		log.Fatalf("failed to create dataset: %v", result.Error)
	}

	// 查询数据
	var retrievedDataset Dataset
	if result := db.Where("id = ?", 1).First(&retrievedDataset); result.Error != nil {
		log.Fatalf("failed to retrieve dataset: %v", result.Error)
	}

	log.Printf("Retrieved dataset: %+v\n", retrievedDataset)

	// 输出样本内容
	for _, sample := range retrievedDataset.Samples {
		log.Printf("Sample: %+v\n", sample)
	}

	opts := make([]meta.ListOption, 0)
	opts = append(opts, meta.WithOffset(int64(1)))
	opts = append(opts, meta.WithOffset(int64(2)))
	filters := make(map[string]any, 0)
	filters["name"] = "aa"
	filters["tags"] = []string{"33"}
	//filters["author"] = *req.Author

	opts = append(opts, meta.WithFilter(filters))
	_, ds, err := QueryDataset(db, opts...)
	fmt.Println(len(ds))
}

func QueryDataset(db *gorm.DB, opts ...meta.ListOption) (int64, []*Dataset, error) {
	o := meta.NewListOptions(opts...)

	tags, ok := o.Filters["tags"]
	if ok {
		r, err := json.Marshal(tags)
		if err != nil {
			return 0, nil, err
		}
		db = db.Where("JSON_OVERLAPS(tags->'$', CAST( ? AS JSON))", string(r))
		delete(o.Filters, "tags")
	}

	var ret []*Dataset
	var count int64
	err := db.
		Where(o.Filters).
		Offset(o.Offset).
		Limit(o.Limit).
		Order("id desc").
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).Error
	if err != nil {
		return 0, nil, err
	}

	return count, ret, nil
}
