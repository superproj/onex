package v1

import (
	"github.com/superproj/onex/internal/nightwatch/biz"
	"github.com/superproj/onex/internal/nightwatch/validation"
)

type NightWatchService struct {
	valid *validation.Validator
	biz   biz.IBiz
}

func NewNightWatchService(valid *validation.Validator, biz biz.IBiz) *NightWatchService {
	return &NightWatchService{valid: valid, biz: biz}
}
