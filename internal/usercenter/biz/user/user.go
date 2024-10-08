// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

//go:generate mockgen -self_package github.com/superproj/onex/internal/usercenter/biz/user -destination mock_user.go -package user github.com/superproj/onex/internal/usercenter/biz/user UserBiz

import (
	"context"
	"errors"
	"regexp"
	"sync"

	"github.com/jinzhu/copier"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/pkg/onexx"
	validationutil "github.com/superproj/onex/internal/pkg/util/validation"
	"github.com/superproj/onex/internal/usercenter/conversion"
	"github.com/superproj/onex/internal/usercenter/model"
	"github.com/superproj/onex/internal/usercenter/store"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/authn"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
)

const (
	defaultMaxWorkers = 100 // Default maximum number of workers for concurrent operations.
)

// UserBiz defines methods used to handle user requests.
type UserBiz interface {
	// Create creates a new user based on the provided request.
	Create(ctx context.Context, rq *v1.CreateUserRequest) (*v1.UserReply, error)

	// Update updates an existing user based on the provided request.
	Update(ctx context.Context, rq *v1.UpdateUserRequest) error

	// Delete removes a user based on the provided request.
	Delete(ctx context.Context, rq *v1.DeleteUserRequest) error

	// Get retrieves a user by username based on the provided request.
	Get(ctx context.Context, rq *v1.GetUserRequest) (*v1.UserReply, error)

	// List retrieves a list of all users based on the provided request.
	List(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error)

	UserExpansion
}

// UserExpansion defines additional methods for user operations.
type UserExpansion interface {
	// UpdatePassword updates the password for a user based on the provided request.
	UpdatePassword(ctx context.Context, rq *v1.UpdatePasswordRequest) error
}

// userBiz is the concrete implementation of the UserBiz interface.
type userBiz struct {
	ds store.IStore // Data store for persistent storage operations.
}

// Ensure userBiz implements the UserBiz interface.
var _ UserBiz = (*userBiz)(nil)

// New creates a new instance of userBiz with the provided data store.
func New(ds store.IStore) *userBiz {
	return &userBiz{ds: ds}
}

// Create creates a new user based on the provided request.
func (b *userBiz) Create(ctx context.Context, rq *v1.CreateUserRequest) (*v1.UserReply, error) {
	var userM model.UserM
	_ = copier.Copy(&userM, rq) // Copy request data to the User model.

	// Start a transaction for creating the user and secret.
	err := b.ds.TX(ctx, func(ctx context.Context) error {
		// Attempt to create the user in the data store.
		if err := b.ds.Users().Create(ctx, &userM); err != nil {
			// Handle duplicate entry error for username.
			match, _ := regexp.MatchString("Duplicate entry '.*' for key 'username'", err.Error())
			if match {
				return v1.ErrorUserAlreadyExists("user %q already exists", userM.Username)
			}
			return v1.ErrorUserCreateFailed("create user failed: %s", err.Error())
		}

		// Create a secret for the newly created user.
		secretM := &model.SecretM{
			UserID:      userM.UserID,
			Name:        "generated",
			Expires:     0,
			Description: "automatically generated when user is created",
		}
		if err := b.ds.Secrets().Create(ctx, secretM); err != nil {
			return v1.ErrorSecretCreateFailed("create secret failed: %s", err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err // Return any error from the transaction.
	}

	return conversion.ConvertToV1UserReply(&userM), nil // Convert and return the created user.
}

// Update updates an existing user based on the provided request.
func (b *userBiz) Update(ctx context.Context, rq *v1.UpdateUserRequest) error {
	whr := where.F("username", rq.Username) // Create a query filter for the username.
	// Limit access to authorized users only.
	if !validationutil.IsAdminUser(onexx.FromUserID(ctx)) {
		whr = whr.T(ctx)
	}

	// Retrieve the user from the data store.
	userM, err := b.ds.Users().Get(ctx, whr)
	if err != nil {
		return err // Return any error encountered.
	}

	// Update fields if provided in the request.
	if rq.Nickname != nil {
		userM.Nickname = *rq.Nickname
	}
	if rq.Email != nil {
		userM.Email = *rq.Email
	}
	if rq.Phone != nil {
		userM.Phone = *rq.Phone
	}

	return b.ds.Users().Update(ctx, userM) // Update the user in the data store.
}

// UpdatePassword updates the password for a user based on the provided request.
func (b *userBiz) UpdatePassword(ctx context.Context, rq *v1.UpdatePasswordRequest) error {
	// Retrieve the user by username.
	userM, err := b.ds.Users().Get(ctx, where.T(ctx).F("username", rq.Username))
	if err != nil {
		return err // Return any error encountered.
	}

	// Compare the old password with the stored password.
	if err := authn.Compare(userM.Password, rq.OldPassword); err != nil {
		return v1.ErrorUserLoginFailed("password incorrect") // Return an error if the old password is incorrect.
	}
	// Encrypt the new password.
	userM.Password, _ = authn.Encrypt(rq.NewPassword)

	return b.ds.Users().Update(ctx, userM) // Update the user's password in the data store.
}

// Delete removes a user based on the provided request.
func (b *userBiz) Delete(ctx context.Context, rq *v1.DeleteUserRequest) error {
	whr := where.F("username", rq.Username) // Create a query filter for the username.
	// Limit access to authorized users only.
	if !validationutil.IsAdminUser(onexx.FromUserID(ctx)) {
		whr = whr.T(ctx)
	}

	return b.ds.Users().Delete(ctx, whr) // Delete the user from the data store.
}

// Get retrieves a user by username based on the provided request.
func (b *userBiz) Get(ctx context.Context, rq *v1.GetUserRequest) (*v1.UserReply, error) {
	whr := where.F("username", rq.Username) // Create a query filter for the username.
	// Limit access to authorized users only.
	if !validationutil.IsAdminUser(onexx.FromUserID(ctx)) {
		whr = whr.T(ctx)
	}

	// Retrieve the user from the data store.
	userM, err := b.ds.Users().Get(ctx, whr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorUserNotFound(err.Error()) // Return an error if the user is not found.
		}
		return nil, err // Return any other error encountered.
	}

	return conversion.ConvertToV1UserReply(userM), nil // Convert and return the found user.
}

// List retrieves a list of all users from the database.
func (b *userBiz) List(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	// Retrieve the total count and list of users from the data store.
	count, userList, err := b.ds.Users().List(ctx, where.P(int(rq.Offset), int(rq.Limit)))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list users from storage")
		return nil, err // Return any error encountered.
	}

	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx) // Create a new error group for concurrent processing.
	// Use goroutines to improve performance when retrieving secrets for users.
	for _, user := range userList {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err() // Return error if the context is done.
			default:
				// Retrieve the count of secrets for each user.
				count, _, err := b.ds.Secrets().List(ctx, where.F("user_id", user.UserID))
				if err != nil {
					log.C(ctx).Errorw(err, "Failed to list secrets")
					return err // Return any error encountered.
				}

				u := conversion.ConvertToV1UserReply(user) // Convert user to response format.
				u.Secrets = count                          // Set the secret count for the user.
				m.Store(user.ID, u)                        // Store the user response in the map.

				return nil
			}
		})
	}

	// Wait for all goroutines to finish.
	if err := eg.Wait(); err != nil {
		log.C(ctx).Errorw(err, "Failed to wait for all function calls to return")
		return nil, err // Return any error encountered.
	}

	// Maintain the consistency of query order while preparing the final user list.
	users := make([]*v1.UserReply, 0, len(userList))
	for _, item := range userList {
		user, _ := m.Load(item.ID)                  // Load the user from the map.
		users = append(users, user.(*v1.UserReply)) // Append the user to the final list.
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Users: users}, nil // Return the response with all retrieved users.
}

// ListWithBadPerformance is a poor performance implementation of List.
func (b *userBiz) ListWithBadPerformance(ctx context.Context, rq *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	// Retrieve the total count and list of users from the data store.
	count, userList, err := b.ds.Users().List(ctx, where.P(int(rq.Offset), int(rq.Limit)))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list users from storage")
		return nil, err // Return any error encountered.
	}

	users := make([]*v1.UserReply, 0) // Initialize the final user response list.
	for _, item := range userList {
		var u v1.UserReply
		_ = copier.Copy(&u, &item) // Copy user data to the response format.

		// Retrieve the count of secrets for each user.
		count, _, err := b.ds.Secrets().List(ctx, where.F("user_id", item.UserID))
		if err != nil {
			log.C(ctx).Errorw(err, "Failed to list secrets")
			return nil, err // Return any error encountered.
		}

		// Set additional user data.
		u.CreatedAt = timestamppb.New(item.CreatedAt)
		u.UpdatedAt = timestamppb.New(item.UpdatedAt)
		u.Password = "******"     // Mask the password in the reply.
		u.Secrets = count         // Set the secret count for the user.
		users = append(users, &u) // Append the user to the final response list.
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Users: users}, nil // Return the response with all retrieved users.
}
