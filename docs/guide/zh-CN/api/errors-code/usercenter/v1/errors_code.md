# 错误码

**注意：** 该组件错误码列表，由 `protoc --go-errors-code_out=paths=source_relative:./docs/guide/zh-CN/api/errors-code` 命令生成，不要对此文件做任何更改。

## 错误说明

如果请求结果返回格式如下：
```json
{
  "metadata": {},
  "message": "User already exists",
  "reason": "UserAlreadyExists",
  "code": 409
}
```

则表示调用 API 接口失败，可能需要客户端进行相应的错误处理。

## 错误码列表

当前组件支持的错误码列表如下：

| Reason | HTTP Status Code | Description |
| :----: | :--------------: | ----------- |
| UserLoginFailed | 401 |  用户登录失败，身份验证未通过 |
| UserAlreadyExists | 409 |  用户已存在，无法创建用户 |
| UserNotFound | 404 |  用户未找到，可能是用户不存在或输入的用户标识有误 |
| UserCreateFailed | 541 |  创建用户失败，可能是由于服务器或其他问题导致的创建过程中的错误 |
| UserOperationForbidden | 403 |  用户操作被禁止，可能是由于权限不足或其他安全限制导致的 |
| SecretReachMaxCount | 400 |  密钥达到最大数量限制，无法继续创建新密钥 |
| SecretNotFound | 404 |  密钥未找到，可能是由于密钥不存在或输入的密钥标识有误 |
| SecretCreateFailed | 541 |  创建密钥失败，可能是由于服务器或其他问题导致的创建过程中的错误 |

## 参考

- [错误规范](https://github.com/superproj/zero/blob/master/docs/devel/zh-CN/conversions/errors.md)

