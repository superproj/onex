// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/pkg/onexx"
	"github.com/superproj/onex/internal/usercenter/store"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockIStore(ctrl)

	type args struct {
		ds store.IStore
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "default",
			args: args{ds: mockStore},
			want: &validator{mockStore},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ds)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validator_ValidateCreateUserRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockIStore(ctrl)
	mockUserStore := store.NewMockUserStore(ctrl)
	mockUserStore.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)
	mockStore.EXPECT().Users().Return(mockUserStore)

	type fields struct {
		ds store.IStore
	}
	type args struct {
		ctx context.Context
		rq  *v1.CreateUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "default",
			fields: fields{mockStore},
			args: args{
				context.Background(), &v1.CreateUserRequest{
					Username: "colin",
					Nickname: "colin",
					Password: "onex(#)666",
					Email:    "colin404@foxmail.com",
					Phone:    "1812884xxxx",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := &validator{
				ds: tt.fields.ds,
			}
			if err := vd.ValidateCreateUserRequest(tt.args.ctx, tt.args.rq); (err != nil) != tt.wantErr {
				t.Errorf("validator.ValidateCreateUserRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validator_ValidateListUserRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		ds store.IStore
	}
	type args struct {
		ctx context.Context
		rq  *v1.ListUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "pass",
			fields: fields{
				ds: nil,
			},
			args: args{
				ctx: onexx.NewUserID(context.Background(), "user-admin"),
				rq: &v1.ListUserRequest{
					Limit:  0,
					Offset: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "fail",
			fields: fields{
				ds: nil,
			},
			args: args{
				ctx: onexx.NewUserID(context.Background(), "user-xxx"),
				rq: &v1.ListUserRequest{
					Limit:  0,
					Offset: 10,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := &validator{
				ds: tt.fields.ds,
			}
			if err := vd.ValidateListUserRequest(tt.args.ctx, tt.args.rq); (err != nil) != tt.wantErr {
				t.Errorf("validator.ValidateListUserRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validator_ValidateCreateSecretRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockIStore(ctrl)
	mockSecretStore := store.NewMockSecretStore(ctrl)
	mockSecretStore.EXPECT().List(gomock.Any(), gomock.Any()).Return(int64(0), nil, nil)
	mockStore.EXPECT().Secrets().Return(mockSecretStore)

	type fields struct {
		ds store.IStore
	}
	type args struct {
		ctx context.Context
		rq  *v1.CreateSecretRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				ds: mockStore,
			},
			args: args{
				ctx: onexx.NewUserID(context.Background(), "user-test"),
				rq: &v1.CreateSecretRequest{
					Name:        "test",
					Expires:     0,
					Description: "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := &validator{
				ds: tt.fields.ds,
			}
			if err := vd.ValidateCreateSecretRequest(tt.args.ctx, tt.args.rq); (err != nil) != tt.wantErr {
				t.Errorf("validator.ValidateCreateSecretRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validator_ValidateAuthRequest(t *testing.T) {
	type fields struct {
		ds store.IStore
	}
	type args struct {
		ctx context.Context
		rq  *v1.AuthRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := &validator{
				ds: tt.fields.ds,
			}
			if err := vd.ValidateAuthRequest(tt.args.ctx, tt.args.rq); (err != nil) != tt.wantErr {
				t.Errorf("validator.ValidateAuthRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validator_ValidateAuthorizeRequest(t *testing.T) {
	type fields struct {
		ds store.IStore
	}
	type args struct {
		ctx context.Context
		rq  *v1.AuthorizeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := &validator{
				ds: tt.fields.ds,
			}
			if err := vd.ValidateAuthorizeRequest(tt.args.ctx, tt.args.rq); (err != nil) != tt.wantErr {
				t.Errorf("validator.ValidateAuthorizeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
