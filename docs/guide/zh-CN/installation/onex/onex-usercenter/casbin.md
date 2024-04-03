# Casbin 授权策略添加

onex-usercenter 基于 Casbin 的 RBAC 授权模式。实现的目标：能够指定路径进行授权。

- 支持的模型请参考：https://casbin.org/zh/docs/supported-models
- Model 语法请参考：https://casbin.org/zh/docs/syntax-for-models

## RBAC Model

```sql
insert into casbin_rule values(id,'g', 'colin', 'data2_admin','','','',''); -- colin 用户属于 data2_admin 角色
insert into casbin_rule values(id,'p', 'colin', 'data1', 'read','allow','',''); -- colin 用户对 data1 有 read 权限
insert into casbin_rule values(id,'p', 'alice', 'data2', 'read','allow','',''); 
insert into casbin_rule values(id,'p', 'alice', 'data2', 'read','deny','',''); 
insert into casbin_rule values(id,'p', 'data2_admin', 'data2', 'read','allow','',''); -- data2_admin 角色对 data2 有读权限
insert into casbin_rule values(id,'p', 'data2_admin', 'data2', 'write','allow','',''); -- data2_admin 角色对 data2 有写权限
```

上述 Polic 的测试：
```bash
{"sub":"colin","obj":"data1","act":"read"} - allow
{"sub":"alice","obj":"data2","act":"read"} - deny
{"sub":"bob","obj":"data3","act":"read"} - deny - 默认 deny
{"sub":"root","obj":"data3","act":"read"} - allow - 超级管理员有一切权限
```

## RESTful Model

```sql
insert into casbin_rule values(id,'p','cathy', '/cathy_data', '(GET)|(POST)|(PUT)|(DELETE)');
```

## onex-usercenter 策略定义 

```toml
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.sub == "root"`
```

- policy_effect: 该 Effect 原语表示当至少存在一个决策结果为 `allow` 的匹配规则，且不存在决策结果为 `deny` 的匹配规则时，则最终决策结果为 `allow`。 这时 `allow` 授权和 `deny` 授权同时存在，但是 `deny` 优先。

在 onex-usercenter 中，默认情况下是默认拒绝的，也就是说如果没有任何匹配的策略，则拒绝访问。这是因为 onex-center 的安全设计原则之一是“最小特权原则”，它要求在没有明确授权的情况下应该默认拒绝访问，以保障系统的安全性。
