// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/errors"

	"github.com/superproj/onex/internal/pkg/core"
	known "github.com/superproj/onex/internal/pkg/known/toyblc"
)

type authPair struct {
	value string
	user  string
}

type authPairs []authPair

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}
	for _, pair := range a {
		if subtle.ConstantTimeCompare([]byte(pair.value), []byte(authValue)) == 1 {
			return pair.user, true
		}
	}
	return "", false
}

// Authn returns a Basic HTTP Authorization middleware. It takes as argument a map[string]string where
// the key is the user name and the value is the password.
func BasicAuth(accounts map[string]string) gin.HandlerFunc {
	realm := "Basic realm=" + strconv.Quote("Authorization Required")
	pairs := processAccounts(accounts)

	return func(c *gin.Context) {
		// Search user in the slice of allowed credentials
		user, found := pairs.searchCredential(c.Request.Header.Get("Authorization"))
		if !found {
			// Credentials doesn't match, we return 401 and abort handlers chain.
			c.Header("WWW-Authenticate", realm)
			core.WriteResponse(
				c,
				errors.Unauthorized("UNAUTHORIZED", "The username or password is incorrect"),
				nil,
			)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// The user credentials was found, set user's id to key known.UsernameKey in this context,
		// the user's id canbe read later using `c.MustGet(known.UsernameKey)``.
		c.Set(known.UsernameKey, user)
		c.Next()
	}
}

func processAccounts(accounts map[string]string) authPairs {
	length := len(accounts)
	pairs := make(authPairs, 0, length)
	for user, password := range accounts {
		value := authorizationHeader(user, password)
		pairs = append(pairs, authPair{
			value: value,
			user:  user,
		})
	}
	return pairs
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}
