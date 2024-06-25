// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

import (
	"context"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/jinzhu/copier"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/pkg/meta"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/log"
)

const (
	defaultMaxWorkers = 100
)

// List retrieves a list of all users from the database.
func (b *userBiz) List(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.Users().List(ctx, meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list users from storage")
		return nil, err
	}

	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)
	// Use goroutine to improve interface performance
	for _, user := range list {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				count, _, err := b.ds.Secrets().List(ctx, user.UserID)
				if err != nil {
					log.C(ctx).Errorw(err, "Failed to list secrets")
					return err
				}

				u := ModelToReply(user)
				u.Secrets = count
				m.Store(user.ID, u)

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.C(ctx).Errorw(err, "Failed to wait all function calls returned")
		return nil, err
	}

	// The following code block is used to maintain the consistency of query order.
	users := make([]*v1.UserReply, 0, len(list))
	for _, item := range list {
		user, _ := m.Load(item.ID)
		users = append(users, user.(*v1.UserReply))
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Users: users}, nil
}

// ListWithWorkerPool retrieves a list of all users from the database use workerpool package.
// Concurrency limits can effectively protect downstream services and control the resource
// consumption of components.
func (b *userBiz) ListWithWorkerPool(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.Users().List(ctx, meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list users from storage")
		return nil, err
	}

	var m sync.Map
	wp := workerpool.New(defaultMaxWorkers)
	// Use goroutine to improve interface performance
	for _, user := range list {
		wp.Submit(func() {
			count, _, err := b.ds.Secrets().List(ctx, user.UserID)
			if err != nil {
				log.C(ctx).Errorw(err, "Failed to list secrets")
				return
			}

			u := ModelToReply(user)
			u.Secrets = count
			m.Store(user.ID, u)

			return
		})
	}

	wp.StopWait()

	// The following code block is used to maintain the consistency of query order.
	users := make([]*v1.UserReply, 0, len(list))
	for _, item := range list {
		user, _ := m.Load(item.ID)
		users = append(users, user.(*v1.UserReply))
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Users: users}, nil
}

// ListWithAnts retrieves a list of all users from the database use ants package.
// Concurrency limits can effectively protect downstream services and control the
// resource consumption of components.
func (b *userBiz) ListWithAnts(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.Users().List(ctx, meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list users from storage")
		return nil, err
	}

	var m sync.Map
	var wg sync.WaitGroup
	pool, _ := ants.NewPool(100)
	defer pool.Release()

	// Use goroutine to improve interface performance
	for _, user := range list {
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()

			count, _, err := b.ds.Secrets().List(ctx, user.UserID)
			if err != nil {
				log.C(ctx).Errorw(err, "Failed to list secrets")
				return
			}

			u := ModelToReply(user)
			u.Secrets = count
			m.Store(user.ID, u)

			return
		})
	}

	wg.Wait()

	// The following code block is used to maintain the consistency of query order.
	users := make([]*v1.UserReply, 0, len(list))
	for _, item := range list {
		user, _ := m.Load(item.ID)
		users = append(users, user.(*v1.UserReply))
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Users: users}, nil
}

// ListWithBadPerformance is a poor performance implementation of List.
func (b *userBiz) ListWithBadPerformance(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.Users().List(ctx, meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list users from storage")
		return nil, err
	}

	users := make([]*v1.UserReply, 0)
	for _, item := range list {
		var u v1.UserReply
		_ = copier.Copy(&u, &item)

		count, _, err := b.ds.Secrets().List(ctx, item.UserID)
		if err != nil {
			log.C(ctx).Errorw(err, "Failed to list secrets")
			return nil, err
		}

		u.CreatedAt = timestamppb.New(item.CreatedAt)
		u.UpdatedAt = timestamppb.New(item.UpdatedAt)
		u.Password = "******"
		u.Secrets = count
		users = append(users, &u)
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Users: users}, nil
}
