# Traefik 概念、原理

Traefik 有 3 种常见的资源类型 IngressRouteTCP、Ingress、IngressRoute，区别如下：

- Ingress：用于HTTP/HTTPS路由的配置。可以通过Ingress规则将外部流量路由到不同的Service或Pod中。
- IngressRoute：用于HTTP/HTTPS路由的高级配置，支持更多的匹配规则和路由策略。
- IngressRouteTCP：用于TCP流量的配置，可以将TCP流量路由到不同的Service或Pod中。

简而言之，IngressRouteTCP是用于TCP流量的配置，而Ingress和IngressRoute都是用于HTTP/HTTPS路由的配置，其中IngressRoute支持更多的高级配置。
