package domain

type PermissionRequest struct {
	SubjectAttrs     map[string]string
	ResourceAttrs    map[string]string
	EnvironmentAttrs map[string]string
}
