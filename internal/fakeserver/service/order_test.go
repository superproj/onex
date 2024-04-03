// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/superproj/onex/internal/fakeserver/biz"
	"github.com/superproj/onex/internal/fakeserver/biz/order"
	v1 "github.com/superproj/onex/pkg/api/fakeserver/v1"
)

func TestFakeServerService_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	want := &v1.CreateOrderResponse{OrderID: "order-22vtll"}
	mockOrderBiz := order.NewMockOrderBiz(ctrl)
	mockBiz := biz.NewMockIBiz(ctrl)
	mockOrderBiz.EXPECT().Create(gomock.Any(), gomock.Any()).Return(want, nil).Times(1)
	mockBiz.EXPECT().Orders().AnyTimes().Return(mockOrderBiz)

	type fields struct {
		UnimplementedFakeServerServer v1.UnimplementedFakeServerServer
		biz                           biz.IBiz
	}
	type args struct {
		ctx context.Context
		rq  *v1.CreateOrderRequest
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *v1.CreateOrderResponse
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:   "default",
			fields: fields{biz: mockBiz},
			args: args{
				ctx: context.Background(),
				rq: &v1.CreateOrderRequest{
					Customer: "colin",
					Product:  "iphone15",
					Quantity: 1,
				},
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &FakeServerService{
				UnimplementedFakeServerServer: tt.fields.UnimplementedFakeServerServer,
				biz:                           tt.fields.biz,
			}
			got, err := s.CreateOrder(tt.args.ctx, tt.args.rq)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
