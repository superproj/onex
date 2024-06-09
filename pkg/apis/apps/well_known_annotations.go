package apps

const (
	// AnnotationDeletionProtection is the key for the annotation used to enable deletion protection for resources.
	// When this annotation is present on a resource, the resource should not be deleted unless the annotation is removed.
	AnnotationDeletionProtection = "apps.onex.io/deletion-protection"
)
