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
| OrderNotFound | 404 |  订单找不到 ，可能是订单不存在或输入的订单标识有误 |
| OrderAlreadyExists | 409 |  订单已存在，无法创建用户 |
| OrderCreateFailed | 541 |  创建订单失败，可能是由于服务器或其他问题导致的创建过程中的错误 |

## 参考

- [错误规范](https://github.com/superproj/zero/blob/master/docs/devel/zh-CN/conversions/errors.md)

