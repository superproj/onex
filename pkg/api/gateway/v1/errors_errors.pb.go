// Code generated by protoc-gen-go-errors. DO NOT EDIT.

package v1

import (
	fmt "fmt"
	errors "github.com/go-kratos/kratos/v2/errors"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
const _ = errors.SupportPackageIsVersion1

// 用户登录失败，可能是用户名或密码不对
func IsUserLoginFailed(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_UserLoginFailed.String() && e.Code == 401
}

// 用户登录失败，可能是用户名或密码不对
func ErrorUserLoginFailed(format string, args ...interface{}) *errors.Error {
	return errors.New(401, ErrorReason_UserLoginFailed.String(), fmt.Sprintf(format, args...))
}

// 用户已存在错误
func IsUserAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_UserAlreadyExists.String() && e.Code == 409
}

// 用户已存在错误
func ErrorUserAlreadyExists(format string, args ...interface{}) *errors.Error {
	return errors.New(409, ErrorReason_UserAlreadyExists.String(), fmt.Sprintf(format, args...))
}

// 用户未找到错误
func IsUserNotFound(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_UserNotFound.String() && e.Code == 404
}

// 用户未找到错误
func ErrorUserNotFound(format string, args ...interface{}) *errors.Error {
	return errors.New(404, ErrorReason_UserNotFound.String(), fmt.Sprintf(format, args...))
}

// 创建用户失败错误
func IsUserCreateFailed(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_UserCreateFailed.String() && e.Code == 541
}

// 创建用户失败错误
func ErrorUserCreateFailed(format string, args ...interface{}) *errors.Error {
	return errors.New(541, ErrorReason_UserCreateFailed.String(), fmt.Sprintf(format, args...))
}
