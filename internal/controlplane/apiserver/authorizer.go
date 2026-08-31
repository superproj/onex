// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package apiserver

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/authorization/union"
	webhookutil "k8s.io/apiserver/pkg/util/webhook"
	"k8s.io/client-go/informers"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac"
)

const (
	// AuthorizationModeAlwaysAllow authorizes every request. It is the default so
	// that existing clients keep working without credentials.
	AuthorizationModeAlwaysAllow = "AlwaysAllow"
	// AuthorizationModeAlwaysDeny denies every request. It is useful for testing.
	AuthorizationModeAlwaysDeny = "AlwaysDeny"
	// AuthorizationModeRBAC authorizes requests using the RBAC authorizer, driven
	// by Role/RoleBinding/ClusterRole/ClusterRoleBinding objects stored in etcd.
	AuthorizationModeRBAC = "RBAC"
	// AuthorizationModeWebhook delegates SubjectAccessReview decisions to a remote
	// authorizer (e.g. a casbin webhook) reachable over the authorization/v1
	// SubjectAccessReview API.
	AuthorizationModeWebhook = "Webhook"
)

// webhookAuthorizationOptions configures the Webhook authorization mode. The
// remote authorizer is a kubeconfig whose `cluster.server` serves the
// authorization/v1 SubjectAccessReview API.
type webhookAuthorizationOptions struct {
	// ConfigFile is the path to the kubeconfig pointing at the remote authorizer.
	ConfigFile string
	// CacheAuthorizedTTL is how long an allow decision is cached for.
	CacheAuthorizedTTL time.Duration
	// CacheUnauthorizedTTL is how long a deny decision is cached for.
	CacheUnauthorizedTTL time.Duration
}

// defaultAuthorizationWebhookRetryBackoff mirrors kube-apiserver's default for
// the delegating authorization webhook retry loop.
var defaultAuthorizationWebhookRetryBackoff = wait.Backoff{
	Duration: 500 * time.Millisecond,
	Factor:   1.5,
	Jitter:   0.2,
	Steps:    5,
}

// BuildAuthorizer returns an authorizer and rule resolver for the requested,
// ordered list of authorization modes. Unlike the kube-apiserver's delegating
// authorization, this does not require a remote kube cluster to serve
// SubjectAccessReviews: the RBAC authorizer runs locally against the RBAC
// objects watched by informerFactory, and the Webhook mode targets a standalone
// external authorizer (e.g. a casbin service).
//
// system:masters is unconditionally allowed, mirroring the kube-apiserver's
// AlwaysAllowGroups configuration, so that the loopback client, bootstrap
// policy and the cluster administrator keep working even before the bootstrap
// roles are fully applied.
func BuildAuthorizer(authorizationMode []string, informerFactory informers.SharedInformerFactory, webhookOptions webhookAuthorizationOptions) (authorizer.Authorizer, authorizer.RuleResolver, error) {
	// Filter out empty modes (e.g. a trailing comma or an empty flag value), so
	// an empty --authorization-mode keeps falling back to AlwaysAllow.
	modes := make([]string, 0, len(authorizationMode))
	for _, mode := range authorizationMode {
		if mode != "" {
			modes = append(modes, mode)
		}
	}
	if len(modes) == 0 {
		modes = []string{AuthorizationModeAlwaysAllow}
	}

	var authorizers []union.NamedAuthorizer
	var ruleResolver authorizer.RuleResolver

	for _, mode := range modes {
		// Keep cases in sync with the validation in options/validation.go.
		switch mode {
		case AuthorizationModeAlwaysAllow:
			authorizers = append(authorizers, union.NamedAuthorizer{
				AuthorizerName: "AlwaysAllow",
				Authorizer:     authorizerfactory.NewAlwaysAllowAuthorizer(),
			})
		case AuthorizationModeAlwaysDeny:
			authorizers = append(authorizers, union.NamedAuthorizer{
				AuthorizerName: "AlwaysDeny",
				Authorizer:     authorizerfactory.NewAlwaysDenyAuthorizer(),
			})
		case AuthorizationModeRBAC:
			rbacAuthorizer := rbac.New(
				&rbac.RoleGetter{Lister: informerFactory.Rbac().V1().Roles().Lister()},
				&rbac.RoleBindingLister{Lister: informerFactory.Rbac().V1().RoleBindings().Lister()},
				&rbac.ClusterRoleGetter{Lister: informerFactory.Rbac().V1().ClusterRoles().Lister()},
				&rbac.ClusterRoleBindingLister{Lister: informerFactory.Rbac().V1().ClusterRoleBindings().Lister()},
			)
			// system:masters is a privileged group that is always allowed.
			authorizers = append(authorizers,
				union.NamedAuthorizer{
					AuthorizerName: "system:masters",
					Authorizer:     authorizerfactory.NewPrivilegedGroups(user.SystemPrivilegedGroup),
				},
				union.NamedAuthorizer{
					AuthorizerName: "RBAC",
					Authorizer:     rbacAuthorizer,
				},
			)
			// RBACAuthorizer also implements authorizer.RuleResolver, which is used
			// to serve the SelfSubjectAccessReview rules endpoint.
			ruleResolver = rbacAuthorizer
		case AuthorizationModeWebhook:
			webhookAuthorizer, err := newWebhookAuthorizer(webhookOptions)
			if err != nil {
				return nil, nil, err
			}
			authorizers = append(authorizers, union.NamedAuthorizer{
				AuthorizerName: "Webhook",
				Authorizer:     webhookAuthorizer,
			})
			// The returned authorizer is the webhook.WebhookAuthorizer, which also
			// implements authorizer.RuleResolver for SelfSubjectAccessReview.
			ruleResolver, _ = webhookAuthorizer.(authorizer.RuleResolver)
		default:
			return nil, nil, fmt.Errorf("unsupported authorization mode %q", mode)
		}
	}

	authz, err := union.New(authorizers...)
	if err != nil {
		return nil, nil, err
	}
	return authz, ruleResolver, nil
}

// newWebhookAuthorizer builds a delegating authorizer that forwards
// SubjectAccessReview requests to the remote authorizer referenced by
// webhookOptions.ConfigFile.
func newWebhookAuthorizer(webhookOptions webhookAuthorizationOptions) (authorizer.Authorizer, error) {
	if webhookOptions.ConfigFile == "" {
		return nil, fmt.Errorf("--authorization-webhook-config-file must be specified when authorization mode %q is enabled", AuthorizationModeWebhook)
	}
	clientConfig, err := webhookutil.LoadKubeconfig(webhookOptions.ConfigFile, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load authorization webhook kubeconfig: %w", err)
	}
	client, err := authorizationclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build SubjectAccessReview client: %w", err)
	}
	config := authorizerfactory.DelegatingAuthorizerConfig{
		SubjectAccessReviewClient: client,
		AllowCacheTTL:             webhookOptions.CacheAuthorizedTTL,
		DenyCacheTTL:              webhookOptions.CacheUnauthorizedTTL,
		WebhookRetryBackoff:       &defaultAuthorizationWebhookRetryBackoff,
	}
	return config.New()
}
