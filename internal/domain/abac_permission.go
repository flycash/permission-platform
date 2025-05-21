package domain

type PermissionRequest struct {
	SubjectAttrValues     map[string]string
	ResourceAttrValues    map[string]string
	EnvironmentAttrValues map[string]string
}
