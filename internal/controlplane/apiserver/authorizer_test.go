// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestBuildAuthorizerModes(t *testing.T) {
	informerFactory := informers.NewSharedInformerFactory(fake.NewSimpleClientset(), 0)

	tests := []struct {
		name         string
		modes        []string
		wantErr      bool
		wantResolver bool
		wantAllow    bool // expected decision for an arbitrary request
	}{
		{
			name:      "nil modes defaults to AlwaysAllow",
			modes:     nil,
			wantAllow: true,
		},
		{
			name:      "empty modes defaults to AlwaysAllow",
			modes:     []string{""},
			wantAllow: true,
		},
		{
			name:      "AlwaysAllow",
			modes:     []string{AuthorizationModeAlwaysAllow},
			wantAllow: true,
		},
		{
			name:      "AlwaysDeny",
			modes:     []string{AuthorizationModeAlwaysDeny},
			wantAllow: false,
		},
		{
			name:         "RBAC",
			modes:        []string{AuthorizationModeRBAC},
			wantResolver: true,
		},
		{
			name:    "unsupported mode",
			modes:   []string{"Unknown"},
			wantErr: true,
		},
		{
			name:    "webhook without config file",
			modes:   []string{AuthorizationModeWebhook},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authz, resolver, err := BuildAuthorizer(tt.modes, informerFactory, webhookAuthorizationOptions{})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantResolver && resolver == nil {
				t.Fatalf("expected a non-nil rule resolver")
			}

			// AlwaysAllow/AlwaysDeny decisions are asserted against an arbitrary request.
			if tt.wantResolver {
				return
			}
			decision, _, err := authz.Authorize(context.Background(), &authorizer.AttributesRecord{
				User:            &user.DefaultInfo{Name: "someone"},
				Verb:            "get",
				Namespace:       "default",
				APIGroup:        "",
				Resource:        "pods",
				ResourceRequest: true,
			})
			if err != nil {
				t.Fatalf("unexpected authorize error: %v", err)
			}
			if got := decision == authorizer.DecisionAllow; got != tt.wantAllow {
				t.Fatalf("expected allow=%v, got decision %v", tt.wantAllow, decision)
			}
		})
	}
}

func TestRBACAuthorizer(t *testing.T) {
	ctx := context.Background()

	// Seed RBAC objects covering both Cluster-level and Namespace-level RBAC.
	client := fake.NewSimpleClientset(
		// -- Cluster-level: alice can get/list pods in any namespace.
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: "cr-get-pods"},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"get", "list"}, APIGroups: []string{""}, Resources: []string{"pods"}},
			},
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "crb-alice"},
			RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "cr-get-pods"},
			Subjects:   []rbacv1.Subject{{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: "alice"}},
		},
		// -- Namespace-level: bob can get pods only in ns1.
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: "role-get-pods", Namespace: "ns1"},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"get"}, APIGroups: []string{""}, Resources: []string{"pods"}},
			},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "rb-bob", Namespace: "ns1"},
			RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "role-get-pods"},
			Subjects:   []rbacv1.Subject{{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: "bob"}},
		},
		// -- ClusterRole granted within a single namespace (ClusterRole + namespaced RoleBinding).
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: "cr-get-secrets"},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"get"}, APIGroups: []string{""}, Resources: []string{"secrets"}},
			},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "rb-carol", Namespace: "ns1"},
			RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "cr-get-secrets"},
			Subjects:   []rbacv1.Subject{{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: "carol"}},
		},
	)

	informerFactory := informers.NewSharedInformerFactory(client, 0)
	authz, ruleResolver, err := BuildAuthorizer([]string{AuthorizationModeRBAC}, informerFactory, webhookAuthorizationOptions{})
	if err != nil {
		t.Fatalf("unexpected error building authorizer: %v", err)
	}

	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(),
		informerFactory.Rbac().V1().Roles().Informer().HasSynced,
		informerFactory.Rbac().V1().RoleBindings().Informer().HasSynced,
		informerFactory.Rbac().V1().ClusterRoles().Informer().HasSynced,
		informerFactory.Rbac().V1().ClusterRoleBindings().Informer().HasSynced,
	) {
		t.Fatalf("failed to sync RBAC informers")
	}

	newAttr := func(name string, groups []string, verb, namespace, resource string) *authorizer.AttributesRecord {
		return &authorizer.AttributesRecord{
			User:            &user.DefaultInfo{Name: name, Groups: groups},
			Verb:            verb,
			Namespace:       namespace,
			APIGroup:        "",
			Resource:        resource,
			ResourceRequest: true,
		}
	}

	tests := []struct {
		name      string
		attrs     *authorizer.AttributesRecord
		wantAllow bool
	}{
		{"alice get pods ns1 (cluster binding)", newAttr("alice", nil, "get", "ns1", "pods"), true},
		{"alice get pods ns2 (cluster binding)", newAttr("alice", nil, "get", "ns2", "pods"), true},
		{"alice delete pods ns1 (verb not allowed)", newAttr("alice", nil, "delete", "ns1", "pods"), false},
		{"bob get pods ns1 (namespaced role)", newAttr("bob", nil, "get", "ns1", "pods"), true},
		{"bob get pods ns2 (role out of scope)", newAttr("bob", nil, "get", "ns2", "pods"), false},
		{"carol get secrets ns1 (clusterrole via rolebinding)", newAttr("carol", nil, "get", "ns1", "secrets"), true},
		{"carol get secrets ns2 (clusterrole scoped to ns1)", newAttr("carol", nil, "get", "ns2", "secrets"), false},
		{"dave get pods (no binding)", newAttr("dave", nil, "get", "ns1", "pods"), false},
		{"system:masters always allowed", newAttr("admin", []string{user.SystemPrivilegedGroup}, "delete", "ns1", "pods"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, _, err := authz.Authorize(ctx, tt.attrs)
			if err != nil {
				t.Fatalf("unexpected authorize error: %v", err)
			}
			if got := decision == authorizer.DecisionAllow; got != tt.wantAllow {
				t.Fatalf("expected allow=%v, got decision %v", tt.wantAllow, decision)
			}
		})
	}

	// The RBAC authorizer must also act as the rule resolver.
	if ruleResolver == nil {
		t.Fatalf("expected a non-nil rule resolver for RBAC mode")
	}
	resourceRules, _, incomplete, err := ruleResolver.RulesFor(ctx, &user.DefaultInfo{Name: "alice"}, "")
	if err != nil {
		t.Fatalf("unexpected error resolving rules: %v", err)
	}
	if incomplete || len(resourceRules) == 0 {
		t.Fatalf("expected complete non-empty rules, got incomplete=%v len=%d", incomplete, len(resourceRules))
	}
}

func TestWebhookAuthorizer(t *testing.T) {
	ctx := context.Background()

	// The fake remote authorizer allows only the user "admin", mirroring a
	// minimal casbin-subject-access-review webhook.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apis/authorization.k8s.io/v1/subjectaccessreviews" {
			http.NotFound(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var sar authorizationv1.SubjectAccessReview
		if err := json.Unmarshal(body, &sar); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if sar.Spec.User == "admin" {
			sar.Status.Allowed = true
		} else {
			// The v1.37 SubjectAccessReviewStatus expresses a rejection via
			// `denied` (the legacy `allowed` field is deprecated).
			sar.Status.Denied = true
			sar.Status.Reason = "user is not admin"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&sar)
	}))
	defer srv.Close()

	kubeconfig := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s
  name: webhook
contexts:
- context:
    cluster: webhook
    user: webhook
  name: webhook
current-context: webhook
users:
- name: webhook
  user: {}
`, srv.URL)

	tmp, err := os.CreateTemp(t.TempDir(), "webhook-kubeconfig-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString(kubeconfig); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	informerFactory := informers.NewSharedInformerFactory(fake.NewSimpleClientset(), 0)
	authz, ruleResolver, err := BuildAuthorizer([]string{AuthorizationModeWebhook}, informerFactory, webhookAuthorizationOptions{
		ConfigFile: tmp.Name(),
		// Disable caching so consecutive requests both reach the remote authorizer.
		CacheAuthorizedTTL:   0,
		CacheUnauthorizedTTL: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error building webhook authorizer: %v", err)
	}
	if ruleResolver == nil {
		t.Fatalf("expected a non-nil rule resolver for Webhook mode")
	}

	attr := func(name string) *authorizer.AttributesRecord {
		return &authorizer.AttributesRecord{
			User:            &user.DefaultInfo{Name: name},
			Verb:            "get",
			Namespace:       "default",
			APIGroup:        "",
			Resource:        "pods",
			ResourceRequest: true,
		}
	}

	decision, _, err := authz.Authorize(ctx, attr("admin"))
	if err != nil {
		t.Fatalf("unexpected authorize error: %v", err)
	}
	if decision != authorizer.DecisionAllow {
		t.Fatalf("expected admin to be allowed, got decision %v", decision)
	}

	decision, _, err = authz.Authorize(ctx, attr("bob"))
	if err != nil {
		t.Fatalf("unexpected authorize error: %v", err)
	}
	if decision != authorizer.DecisionDeny {
		t.Fatalf("expected bob to be denied, got decision %v", decision)
	}
}
