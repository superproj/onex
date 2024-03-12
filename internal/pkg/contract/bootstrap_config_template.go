// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package contract

import (
	"sync"
)

// BootstrapConfigTemplateContract encodes information about the Cluster API contract for BootstrapConfigTemplate objects
// like KubeadmConfigTemplate, etc.
type BootstrapConfigTemplateContract struct{}

var (
	bootstrapConfigTemplate     *BootstrapConfigTemplateContract
	onceBootstrapConfigTemplate sync.Once
)

// BootstrapConfigTemplate provide access to the information about the Cluster API contract for BootstrapConfigTemplate objects.
func BootstrapConfigTemplate() *BootstrapConfigTemplateContract {
	onceBootstrapConfigTemplate.Do(func() {
		bootstrapConfigTemplate = &BootstrapConfigTemplateContract{}
	})
	return bootstrapConfigTemplate
}

// Template provides access to the template.
func (c *BootstrapConfigTemplateContract) Template() *BootstrapConfigTemplateTemplate {
	return &BootstrapConfigTemplateTemplate{}
}

// BootstrapConfigTemplateTemplate provides a helper struct for working with the template in an BootstrapConfigTemplate.
type BootstrapConfigTemplateTemplate struct{}

// Metadata provides access to the metadata of a template.
func (c *BootstrapConfigTemplateTemplate) Metadata() *Metadata {
	return &Metadata{
		path: Path{"spec", "template", "metadata"},
	}
}
