// Package resourceclean is used to delete resoures from database which is already deleted from onex-apiserver.

// Sometimes resources are deleted from the onex-apiserver, but due to reasons such as deployments,
// component failures, etc., the sync controller fails to remove the resources from the database,
// resulting in residual dirty data in the database. In such cases, the clean controller is needed
// to delete the residual resources from the database.

// Alternatively, you have another option: disable the clean controller, set the deletedAt field
// during synchronization to maintain historical resource data.

package resourceclean // import "github.com/superproj/onex/internal/controller/resourceclean"
