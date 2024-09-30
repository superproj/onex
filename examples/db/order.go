package main

import (
	"context"
	"fmt"
	"log"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/fakeserver/store"
	"github.com/superproj/onex/internal/fakeserver/store/mysql"
	"github.com/superproj/onex/pkg/db"
	"github.com/superproj/onex/pkg/store/where"
)

func main() {
	optss := &db.MySQLOptions{
		Addr:     "10.37.43.62:3306",
		Username: "onex",
		Password: "onex(#)666",
		Database: "onex",
	}

	db, err := db.NewMySQL(optss)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	store.SetStore(mysql.NewStore(db))

	order := model.OrderM{
		Customer: "colin404",
		Product:  "phone",
		Quantity: 10,
	}
	if err := store.S.Orders().Create(context.TODO(), &order); err != nil {
		panic(err)
	}

	order.Customer = "colin505"
	if err := store.S.Orders().Update(context.TODO(), &order); err != nil {
		panic(err)
	}

	_, orderList, err := store.S.Orders().List(context.TODO(), where.F("customer", "colin505").L(1))
	for _, order := range orderList {
		fmt.Println(order.Customer)
	}

	if err := store.S.Orders().Delete(context.TODO(), where.F("order_id", 1)); err != nil {
		panic(err)
	}
}
