// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	known "github.com/superproj/onex/internal/pkg/known/usercenter"
	"github.com/superproj/onex/internal/pkg/zid"
	"github.com/superproj/onex/pkg/authn"
)

// Secret status constants.
const (
	StatusSecretDisabled = iota // Status used for disabling a secret.
	StatusSecretNormal          // Status used for enabling a secret.
)

// BeforeCreate runs before creating a SecretM database record and initializes various fields.
func (s *SecretM) BeforeCreate(tx *gorm.DB) (err error) {
	// Supports custom SecretKey, but users need to ensure the uniqueness of the SecretKey themselves.
	// onex-cacheserver will use this feature to set secret.
	if s.SecretID == "" {
		// Generate a new UUID for SecretKey.
		s.SecretID = uuid.New().String()
	}

	// Generate a new UUID for SecretID.
	s.SecretKey = uuid.New().String()

	// Set the default status for the secret as normal.
	s.Status = StatusSecretNormal

	return nil
}

// AfterCreate runs after creating a UserM database record and updates the UserID field.
func (u *UserM) AfterCreate(tx *gorm.DB) (err error) {
	u.UserID = zid.User.New(uint64(u.ID)) // Generate and set a new user ID.

	return tx.Save(u).Error // Save the updated user record to the database.
}

// BeforeCreate runs before creating a UserM database record and initializes various fields.
func (u *UserM) BeforeCreate(tx *gorm.DB) (err error) {
	u.Password, err = authn.Encrypt(u.Password) // Encrypt the user password.
	if err != nil {
		return err // Return error if there's a problem with encryption.
	}

	u.Status = known.UserStatusActive // Set the default status for the user as active.

	return nil
}
