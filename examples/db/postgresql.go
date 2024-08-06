package main

import (
	"log"

	"github.com/superproj/onex/pkg/db"
)

// User 模型
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Email string `gorm:"unique;not null"`
}

func main() {
	opts := &db.PostgreSQLOptions{
		Addr:     "10.37.43.62:5432",
		Username: "easyai",
		Password: "easyai(#)666",
		Database: "easyai",
	}

	db, err := db.NewPostgreSQL(opts)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 插入数据
	newUser := User{Name: "John Doe", Email: "john.doe@example.com"}
	result := db.Create(&newUser)
	if result.Error != nil {
		log.Fatalf("failed to create user: %v", result.Error)
	}

	log.Printf("User created: %v\n", newUser)

	// 查询数据
	var users []User
	db.Find(&users)
	log.Printf("Users: %v\n", users)

	// 更新数据
	db.Model(&newUser).Update("Email", "john.new@example.com")

	// 删除数据
	db.Delete(&newUser)
}
