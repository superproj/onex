package conversion

import (
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// ConvertToV1SecretReply converts a Secret model to its v1 representation.
func ConvertToV1SecretReply(secretM *model.SecretM) *v1.SecretReply {
	var secret v1.SecretReply
	_ = copier.Copy(&secret, secretM)
	secret.CreatedAt = timestamppb.New(secretM.CreatedAt)
	secret.UpdatedAt = timestamppb.New(secretM.UpdatedAt)
	return &secret
}

// ConvertToV1UserReply converts a User model to its v1 representation.
func ConvertToV1UserReply(userM *model.UserM) *v1.UserReply {
	var user v1.UserReply
	_ = copier.Copy(&user, userM)
	user.CreatedAt = timestamppb.New(userM.CreatedAt)
	user.UpdatedAt = timestamppb.New(userM.UpdatedAt)
	return &user
}
