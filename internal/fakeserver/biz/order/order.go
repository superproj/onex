// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package order

//go:generate mockgen -destination mock_order.go -package order github.com/superproj/onex/internal/fakeserver/biz/order OrderBiz

import (
	"context"
	"errors"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/jinzhu/copier"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/fakeserver/model"
	"github.com/superproj/onex/internal/fakeserver/store"
	v1 "github.com/superproj/onex/pkg/api/fakeserver/v1"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
)

type OrderBiz interface {
	Create(ctx context.Context, rq *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error)
	List(ctx context.Context, rq *v1.ListOrderRequest) (*v1.ListOrderResponse, error)
	Get(ctx context.Context, rq *v1.GetOrderRequest) (*v1.OrderReply, error)
	Update(ctx context.Context, rq *v1.UpdateOrderRequest) error
	Delete(ctx context.Context, rq *v1.DeleteOrderRequest) error
}

type orderBiz struct {
	ds store.IStore
}

var _ OrderBiz = (*orderBiz)(nil)

func New(ds store.IStore) *orderBiz {
	return &orderBiz{ds: ds}
}

func (b *orderBiz) Create(ctx context.Context, rq *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error) {
	var orderM model.OrderM
	_ = copier.Copy(&orderM, rq)

	if err := b.ds.Orders().Create(ctx, &orderM); err != nil {
		return nil, v1.ErrorOrderCreateFailed("create order failed: %v", err)
	}

	return &v1.CreateOrderResponse{OrderID: orderM.OrderID}, nil
}

func (b *orderBiz) Update(ctx context.Context, rq *v1.UpdateOrderRequest) error {
	orderM, err := b.ds.Orders().Get(ctx, where.F("order_id", rq.OrderID))
	if err != nil {
		return err
	}

	if rq.Customer != nil {
		orderM.Customer = *rq.Customer
	}

	if rq.Product != nil {
		orderM.Product = *rq.Product
	}

	if rq.Quantity != nil {
		orderM.Quantity = *rq.Quantity
	}

	return b.ds.Orders().Update(ctx, orderM)
}

// Delete 是 OrderBiz 接口中 `Delete` 方法的实现.
func (b *orderBiz) Delete(ctx context.Context, rq *v1.DeleteOrderRequest) error {
	if err := b.ds.Orders().Delete(ctx, where.F("order_id", rq.OrderID)); err != nil {
		return err
	}

	return nil
}

func (b *orderBiz) Get(ctx context.Context, rq *v1.GetOrderRequest) (*v1.OrderReply, error) {
	orderM, err := b.ds.Orders().Get(ctx, where.F("order_id", rq.OrderID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorOrderNotFound(err.Error())
		}

		return nil, err
	}

	var order v1.OrderReply
	_ = copier.Copy(&order, orderM)
	order.CreatedAt = timestamppb.New(orderM.CreatedAt)
	order.UpdatedAt = timestamppb.New(orderM.UpdatedAt)

	return &order, nil
}

func (b *orderBiz) List(ctx context.Context, rq *v1.ListOrderRequest) (*v1.ListOrderResponse, error) {
	whr := where.NewWhere(where.WithOffset(rq.Offset), where.WithLimit(rq.Limit))
	count, orderList, err := b.ds.Orders().List(ctx, whr)
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list orders from storage")
		return nil, err
	}

	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)
	// 使用 goroutine 提高接口性能
	for _, order := range orderList {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				var o v1.OrderReply
				_ = copier.Copy(&o, order)
				m.Store(order.ID, &v1.OrderReply{
					OrderID:   order.OrderID,
					Customer:  order.Customer,
					Product:   order.Product,
					Quantity:  order.Quantity,
					CreatedAt: timestamppb.New(order.CreatedAt),
					UpdatedAt: timestamppb.New(order.UpdatedAt),
				})

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.C(ctx).Errorw(err, "Failed to wait all function calls returned")
		return nil, err
	}

	// The following code block is used to maintain the consistency of query order.
	orders := make([]*v1.OrderReply, 0, len(orderList))
	for _, item := range orderList {
		order, _ := m.Load(item.ID)
		orders = append(orders, order.(*v1.OrderReply))
	}

	log.C(ctx).Debugw("Get orders from backend storage", "count", len(orders))

	return &v1.ListOrderResponse{TotalCount: count, Orders: orders}, nil
}

// ListWithWorkerPool retrieves a list of all orders from the database use workerpool package.
// Concurrency limits can effectively protect downstream services and control the resource
// consumption of components.
func (b *orderBiz) ListWithWorkerPool(ctx context.Context, rq *v1.ListOrderRequest) (*v1.ListOrderResponse, error) {
	whr := where.NewWhere(where.WithOffset(rq.Offset), where.WithLimit(rq.Limit))
	count, orderList, err := b.ds.Orders().List(ctx, whr)
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list orders from storage")
		return nil, err
	}

	var m sync.Map
	wp := workerpool.New(100)

	// Use goroutine to improve interface performance
	for _, order := range orderList {
		wp.Submit(func() {
			var o v1.OrderReply
			// Here simulates a time-consuming concurrent logic.
			_ = copier.Copy(&o, order)
			m.Store(order.ID, &v1.OrderReply{
				OrderID:   order.OrderID,
				Customer:  order.Customer,
				Product:   order.Product,
				Quantity:  order.Quantity,
				CreatedAt: timestamppb.New(order.CreatedAt),
				UpdatedAt: timestamppb.New(order.UpdatedAt),
			})

			return
		})
	}

	wp.StopWait()

	// The following code block is used to maintain the consistency of query order.
	orders := make([]*v1.OrderReply, 0, len(orderList))
	for _, item := range orderList {
		order, _ := m.Load(item.ID)
		orders = append(orders, order.(*v1.OrderReply))
	}

	log.C(ctx).Debugw("Get orders from backend storage", "count", len(orders))

	return &v1.ListOrderResponse{TotalCount: count, Orders: orders}, nil
}

// ListWithAnts retrieves a list of all orders from the database use ants package.
// Concurrency limits can effectively protect downstream services and control the
// resource consumption of components.
func (b *orderBiz) ListWithAnts(ctx context.Context, rq *v1.ListOrderRequest) (*v1.ListOrderResponse, error) {
	whr := where.NewWhere(where.WithOffset(rq.Offset), where.WithLimit(rq.Limit))
	count, orderList, err := b.ds.Orders().List(ctx, whr)
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list orders from storage")
		return nil, err
	}

	var m sync.Map
	var wg sync.WaitGroup
	pool, _ := ants.NewPool(100)
	defer pool.Release()

	// Use goroutine to improve interface performance
	for _, order := range orderList {
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()

			var o v1.OrderReply
			// Here simulates a time-consuming concurrent logic.
			_ = copier.Copy(&o, order)
			m.Store(order.ID, &v1.OrderReply{
				OrderID:   order.OrderID,
				Customer:  order.Customer,
				Product:   order.Product,
				Quantity:  order.Quantity,
				CreatedAt: timestamppb.New(order.CreatedAt),
				UpdatedAt: timestamppb.New(order.UpdatedAt),
			})

			return
		})
	}

	wg.Wait()

	// The following code block is used to maintain the consistency of query order.
	orders := make([]*v1.OrderReply, 0, len(orderList))
	for _, item := range orderList {
		order, _ := m.Load(item.ID)
		orders = append(orders, order.(*v1.OrderReply))
	}

	log.C(ctx).Debugw("Get orders from backend storage", "count", len(orders))

	return &v1.ListOrderResponse{TotalCount: count, Orders: orders}, nil
}
