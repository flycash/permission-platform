package repository

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
)

type RBACRepository interface {
	// CheckUserPermission 检查用户是否有对指定权限
	CheckUserPermission(ctx context.Context, bizID, userID int64, permission domain.Permission) (bool, error)

	// Role相关方法

	CreateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	GetRole(ctx context.Context, id int64) (domain.Role, error)
	UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error)
	DeleteRole(ctx context.Context, id int64) error
	ListRoles(ctx context.Context, bizID int64, offset, limit int, roleType string) ([]domain.Role, int, error)

	// Resource相关方法

	CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	GetResource(ctx context.Context, id int64) (domain.Resource, error)
	UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error)
	DeleteResource(ctx context.Context, id int64) error
	ListResources(ctx context.Context, bizID int64, offset, limit int, resourceType, key string) ([]domain.Resource, int, error)
	FindResource(ctx context.Context, bizID int64, key string) (domain.Resource, error)

	// Permission相关方法

	CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	GetPermission(ctx context.Context, id int64) (domain.Permission, error)
	UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error)
	DeletePermission(ctx context.Context, id int64) error
	ListPermissions(ctx context.Context, bizID int64, offset, limit int, resourceType, resourceKey, action string) ([]domain.Permission, int, error)
	FindPermissions(ctx context.Context, bizID int64, resourceKey string, action domain.ActionType) ([]domain.Permission, error)

	// 用户角色相关方法

	GrantUserRole(ctx context.Context, bizID int64, userID int64, roleID int64, startTime, endTime int64) (domain.UserRole, error)
	RevokeUserRole(ctx context.Context, bizID int64, userID int64, roleID int64) error
	ListUserRoles(ctx context.Context, bizID int64, userID int64, offset, limit int) ([]domain.UserRole, int, error)

	// 业务配置相关方法

	GetBusinessConfig(ctx context.Context, id int64) (domain.BusinessConfig, error)
	GetBusinessConfigs(ctx context.Context, ids []int64) (map[int64]domain.BusinessConfig, error)
	SaveBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error)
	DeleteBusinessConfig(ctx context.Context, id int64) error
	ListBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, int, error)

	// 角色权限相关方法

	GrantRolePermission(ctx context.Context, bizID, roleID, permissionID, startTime, endTime int64) (domain.RolePermission, error)
	RevokeRolePermission(ctx context.Context, bizID, roleID, permissionID int64) error
	ListRolePermissions(ctx context.Context, bizID, roleID int64, offset, limit int) ([]domain.RolePermission, int, error)

	// 角色包含关系相关方法

	CreateRoleInclusion(ctx context.Context, bizID, includingRoleID, includedRoleID int64) (domain.RoleInclusion, error)
	GetRoleInclusion(ctx context.Context, id int64) (domain.RoleInclusion, error)
	DeleteRoleInclusion(ctx context.Context, bizID, includingRoleID, includedRoleID int64) error
	ListRoleInclusions(ctx context.Context, bizID, roleID int64, isIncluding bool, offset, limit int) ([]domain.RoleInclusion, int, error)

	// 用户权限相关方法

	GrantUserPermission(ctx context.Context, bizID, userID, permissionID int64, effect string, startTime, endTime int64) (domain.UserPermission, error)
	RevokeUserPermission(ctx context.Context, bizID, userID, permissionID int64) error
	ListUserPermissions(ctx context.Context, bizID, userID int64, offset, limit int, onlyValid bool) ([]domain.UserPermission, int, error)
}

type rbacRepository struct {
	resourceDAO       dao.ResourceDAO
	permissionDAO     dao.PermissionDAO
	roleDAO           dao.RoleDAO
	rolePermissionDAO dao.RolePermissionDAO
	roleInclusionDAO  dao.RoleInclusionDAO
	userPermissionDAO dao.UserPermissionDAO
	userRoleDAO       dao.UserRoleDAO
	businessConfigDAO dao.BusinessConfigDAO
}

// NewRBACRepository 创建RBAC仓储的实例
func NewRBACRepository(
	resourceDAO dao.ResourceDAO,
	permissionDAO dao.PermissionDAO,
	roleDAO dao.RoleDAO,
	rolePermissionDAO dao.RolePermissionDAO,
	roleInclusionDAO dao.RoleInclusionDAO,
	userPermissionDAO dao.UserPermissionDAO,
	userRoleDAO dao.UserRoleDAO,
	businessConfigDAO dao.BusinessConfigDAO,
) RBACRepository {
	return &rbacRepository{
		resourceDAO:       resourceDAO,
		permissionDAO:     permissionDAO,
		roleDAO:           roleDAO,
		rolePermissionDAO: rolePermissionDAO,
		roleInclusionDAO:  roleInclusionDAO,
		userPermissionDAO: userPermissionDAO,
		userRoleDAO:       userRoleDAO,
		businessConfigDAO: businessConfigDAO,
	}
}

// convertDomainActionToDAOAction 将领域模型的操作类型转换为DAO的操作类型
func (r *rbacRepository) convertDomainActionToDAOAction(action domain.ActionType) dao.ActionType {
	switch action {
	case domain.ActionTypeRead:
		return dao.ActionTypeRead
	case domain.ActionTypeWrite:
		return dao.ActionTypeUpdate
	case domain.ActionTypeCreate:
		return dao.ActionTypeCreate
	case domain.ActionTypeDelete:
		return dao.ActionTypeDelete
	case domain.ActionTypeExecute:
		return dao.ActionTypeExecute
	case domain.ActionTypeExport:
		return dao.ActionTypeExport
	case domain.ActionTypeImport:
		return dao.ActionTypeImport
	default:
		return dao.ActionTypeRead // 默认为读权限
	}
}

func (r *rbacRepository) FindResource(ctx context.Context, bizID int64, key string) (domain.Resource, error) {
	res, err := r.resourceDAO.FindByBizIDAndKey(ctx, bizID, key)
	return r.toResourceDomain(res), err
}

func (r *rbacRepository) FindPermissions(ctx context.Context, bizID int64, resourceKey string, action domain.ActionType) ([]domain.Permission, error) {
	permissions, err := r.permissionDAO.FindPermissions(ctx, bizID, resourceKey, r.convertDomainActionToDAOAction(action))
	if err != nil {
		return nil, err
	}
	list := slice.Map(permissions, func(_ int, src dao.Permission) domain.Permission {
		return r.toPermissionDomain(src)
	})
	return list, nil
}

// CheckUserPermission 检查用户是否具有特定权限
func (r *rbacRepository) CheckUserPermission(ctx context.Context, bizID, userID int64, permission domain.Permission) (bool, error) {
	// 1. 转换操作类型
	daoAction := r.convertDomainActionToDAOAction(permission.Action)
	resourceKey := permission.ResourceKey

	// 2. 检查直接分配给用户的权限
	currentTime := time.Now().UnixMilli()
	userPermissions, err := r.userPermissionDAO.FindValidPermissions(ctx, bizID, userID, currentTime)
	if err != nil {
		return false, err
	}

	for i := range userPermissions {
		// 处理直接分配给用户的权限
		if userPermissions[i].ResourceKey == resourceKey && userPermissions[i].Action == daoAction {
			// 如果有明确的拒绝权限，直接返回无权限
			if userPermissions[i].Effect == dao.EffectTypeDeny {
				return false, nil
			}
			// 如果有明确的允许权限，直接返回有权限
			if userPermissions[i].Effect == dao.EffectTypeAllow {
				return true, nil
			}
		}
	}

	// 3. 获取用户所有有效角色
	userRoles, err := r.userRoleDAO.FindValidRoles(ctx, bizID, userID, currentTime)
	if err != nil {
		return false, err
	}

	roleIDs := make([]int64, 0, len(userRoles))
	for _, userRole := range userRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	// 如果用户没有任何角色，则无权限
	if len(roleIDs) == 0 {
		return false, nil
	}

	// 4. 处理角色继承关系，获取所有角色（包括继承角色）
	allRoleIDs := make(map[int64]struct{})
	for _, roleID := range roleIDs {
		allRoleIDs[roleID] = struct{}{}
		// 获取该角色包含的其他角色
		inclusions, err := r.roleInclusionDAO.FindByIncludingRoleID(ctx, bizID, roleID)
		if err != nil {
			return false, err
		}

		for _, inclusion := range inclusions {
			allRoleIDs[inclusion.IncludedRoleID] = struct{}{}
		}
	}

	// 将map转换为slice
	allRoles := make([]int64, 0, len(allRoleIDs))
	for roleID := range allRoleIDs {
		allRoles = append(allRoles, roleID)
	}

	// 5. 检查角色权限
	for _, roleID := range allRoles {
		rolePermissions, err := r.rolePermissionDAO.FindByRoleID(ctx, bizID, roleID)
		if err != nil {
			return false, err
		}

		for i := range rolePermissions {
			if rolePermissions[i].ResourceKey == resourceKey && rolePermissions[i].Action == daoAction {
				return true, nil
			}
		}
	}

	// 6. 默认无权限
	return false, nil
}

// Role相关方法实现
func (r *rbacRepository) CreateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	create, err := r.roleDAO.Create(ctx, r.toRoleEntity(role))
	if err != nil {
		return domain.Role{}, err
	}
	return r.toRoleDomain(create), nil
}

func (r *rbacRepository) GetRole(ctx context.Context, id int64) (domain.Role, error) {
	role, err := r.roleDAO.GetByID(ctx, id)
	if err != nil {
		return domain.Role{}, err
	}
	return r.toRoleDomain(role), nil
}

func (r *rbacRepository) UpdateRole(ctx context.Context, role domain.Role) (domain.Role, error) {
	err := r.roleDAO.Update(ctx, r.toRoleEntity(role))
	if err != nil {
		return domain.Role{}, err
	}

	// 更新成功后重新获取角色信息
	updatedRole, err := r.roleDAO.GetByID(ctx, role.ID)
	if err != nil {
		return domain.Role{}, err
	}
	return r.toRoleDomain(updatedRole), nil
}

func (r *rbacRepository) DeleteRole(ctx context.Context, id int64) error {
	return r.roleDAO.Delete(ctx, id)
}

func (r *rbacRepository) ListRoles(ctx context.Context, bizID int64, offset, limit int, roleType string) ([]domain.Role, int, error) {
	var roles []dao.Role
	var err error

	if roleType != "" {
		roles, err = r.roleDAO.FindByBizIDAndType(ctx, bizID, dao.RoleType(roleType), offset, limit)
	} else {
		roles, err = r.roleDAO.FindByBizID(ctx, bizID, offset, limit)
	}

	if err != nil {
		return nil, 0, err
	}

	// 简化处理，实际项目中可能需要额外查询总数
	count := len(roles)

	domainRoles := make([]domain.Role, 0, len(roles))
	for i := range roles {
		domainRoles = append(domainRoles, r.toRoleDomain(roles[i]))
	}
	return domainRoles, count, nil
}

// Resource相关方法实现
func (r *rbacRepository) CreateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	created, err := r.resourceDAO.Create(ctx, r.toResourceEntity(resource))
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toResourceDomain(created), nil
}

func (r *rbacRepository) GetResource(ctx context.Context, id int64) (domain.Resource, error) {
	resource, err := r.resourceDAO.GetByID(ctx, id)
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toResourceDomain(resource), nil
}

func (r *rbacRepository) UpdateResource(ctx context.Context, resource domain.Resource) (domain.Resource, error) {
	err := r.resourceDAO.Update(ctx, r.toResourceEntity(resource))
	if err != nil {
		return domain.Resource{}, err
	}

	// 更新后重新获取
	updatedResource, err := r.resourceDAO.GetByID(ctx, resource.ID)
	if err != nil {
		return domain.Resource{}, err
	}
	return r.toResourceDomain(updatedResource), nil
}

func (r *rbacRepository) DeleteResource(ctx context.Context, id int64) error {
	return r.resourceDAO.Delete(ctx, id)
}

func (r *rbacRepository) ListResources(ctx context.Context, bizID int64, offset, limit int, resourceType, key string) ([]domain.Resource, int, error) {
	var resources []dao.Resource
	var err error

	switch {
	case resourceType != "" && key != "":
		// 如果需要按照type和key过滤
		resources, err = r.resourceDAO.FindByBizIDAndType(ctx, bizID, resourceType, offset, limit)
		// 在内存中进一步过滤key
		if err == nil && key != "" {
			var filtered []dao.Resource
			for i := range resources {
				if resources[i].Key == key {
					filtered = append(filtered, resources[i])
				}
			}
			resources = filtered
		}
	case resourceType != "":
		// 只按type过滤
		resources, err = r.resourceDAO.FindByBizIDAndType(ctx, bizID, resourceType, offset, limit)
	case key != "":
		// 只按key过滤
		resource, err := r.resourceDAO.FindByBizIDAndKey(ctx, bizID, key)
		if err == nil {
			resources = []dao.Resource{resource}
		}
	default:
		// 不过滤
		resources, err = r.resourceDAO.FindByBizID(ctx, bizID, offset, limit)
	}

	if err != nil {
		return nil, 0, err
	}

	// 结果总数，实际项目中可能需要单独查询
	count := len(resources)

	domainResources := make([]domain.Resource, 0, len(resources))
	for i := range resources {
		domainResources = append(domainResources, r.toResourceDomain(resources[i]))
	}
	return domainResources, count, nil
}

// Permission相关方法实现
func (r *rbacRepository) CreatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	created, err := r.permissionDAO.Create(ctx, r.toPermissionEntity(permission))
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toPermissionDomain(created), nil
}

func (r *rbacRepository) GetPermission(ctx context.Context, id int64) (domain.Permission, error) {
	permission, err := r.permissionDAO.GetByID(ctx, id)
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toPermissionDomain(permission), nil
}

func (r *rbacRepository) UpdatePermission(ctx context.Context, permission domain.Permission) (domain.Permission, error) {
	err := r.permissionDAO.Update(ctx, r.toPermissionEntity(permission))
	if err != nil {
		return domain.Permission{}, err
	}

	// 更新后重新获取
	updatedPermission, err := r.permissionDAO.GetByID(ctx, permission.ID)
	if err != nil {
		return domain.Permission{}, err
	}
	return r.toPermissionDomain(updatedPermission), nil
}

func (r *rbacRepository) DeletePermission(ctx context.Context, id int64) error {
	return r.permissionDAO.Delete(ctx, id)
}

func (r *rbacRepository) ListPermissions(ctx context.Context, bizID int64, offset, limit int, resourceType, resourceKey, action string) ([]domain.Permission, int, error) {
	var permissions []dao.Permission
	var err error

	// 根据不同参数组合选择不同的查询方法
	switch {
	case resourceType != "":
		permissions, err = r.permissionDAO.FindByBizIDAndResourceType(ctx, bizID, resourceType, offset, limit)
	case resourceKey != "":
		permissions, err = r.permissionDAO.FindByBizIDAndResourceKey(ctx, bizID, resourceKey, offset, limit)
	case action != "":
		permissions, err = r.permissionDAO.FindByBizIDAndAction(ctx, bizID, dao.ActionType(action), offset, limit)
	default:
		permissions, err = r.permissionDAO.FindByBizID(ctx, bizID, offset, limit)
	}

	if err != nil {
		return nil, 0, err
	}

	// 结果总数
	count := len(permissions)

	domainPermissions := make([]domain.Permission, 0, len(permissions))
	for i := range permissions {
		domainPermissions = append(domainPermissions, r.toPermissionDomain(permissions[i]))
	}
	return domainPermissions, count, nil
}

// 用户角色相关方法实现
func (r *rbacRepository) GrantUserRole(ctx context.Context, bizID, userID, roleID, startTime, endTime int64) (domain.UserRole, error) {
	userRole := dao.UserRole{
		BizID:     bizID,
		UserID:    userID,
		RoleID:    roleID,
		StartTime: startTime,
		EndTime:   endTime,
	}
	created, err := r.userRoleDAO.Create(ctx, userRole)
	if err != nil {
		return domain.UserRole{}, err
	}
	return r.toUserRoleDomain(created), nil
}

func (r *rbacRepository) RevokeUserRole(ctx context.Context, bizID, userID, roleID int64) error {
	return r.userRoleDAO.DeleteByUserIDAndRoleID(ctx, bizID, userID, roleID)
}

func (r *rbacRepository) ListUserRoles(ctx context.Context, bizID, userID int64, offset, limit int) ([]domain.UserRole, int, error) {
	// 首先查找用户所有角色
	userRoles, err := r.userRoleDAO.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// 过滤出指定bizID的角色
	var filteredRoles []dao.UserRole
	for i := range userRoles {
		if userRoles[i].BizID == bizID {
			filteredRoles = append(filteredRoles, userRoles[i])
		}
	}

	// 处理分页
	startIndex := offset
	endIndex := offset + limit
	if startIndex >= len(filteredRoles) {
		return []domain.UserRole{}, 0, nil
	}
	if endIndex > len(filteredRoles) {
		endIndex = len(filteredRoles)
	}

	// 总数
	count := len(filteredRoles)

	// 转换为domain模型
	paginatedRoles := filteredRoles[startIndex:endIndex]
	domainUserRoles := make([]domain.UserRole, 0, len(paginatedRoles))
	for i := range paginatedRoles {
		domainUserRoles = append(domainUserRoles, r.toUserRoleDomain(paginatedRoles[i]))
	}

	return domainUserRoles, count, nil
}

func (r *rbacRepository) toRoleEntity(role domain.Role) dao.Role {
	return dao.Role{
		ID:          role.ID,
		BizID:       role.BizID,
		Type:        dao.RoleType(role.Type),
		Name:        role.Name,
		Description: role.Description,
		StartTime:   role.StartTime,
		EndTime:     role.EndTime,
		Ctime:       role.Ctime,
		Utime:       role.Utime,
	}
}

func (r *rbacRepository) toRoleDomain(role dao.Role) domain.Role {
	var roleType domain.RoleType
	switch role.Type {
	case "system":
		roleType = domain.RoleTypeSystem
	case "custom":
		roleType = domain.RoleTypeCustom
	case "temporary":
		roleType = domain.RoleTypeTemporary
	default:
		roleType = domain.RoleTypeCustom
	}

	return domain.Role{
		ID:          role.ID,
		BizID:       role.BizID,
		Type:        roleType,
		Name:        role.Name,
		Description: role.Description,
		StartTime:   role.StartTime,
		EndTime:     role.EndTime,
		Ctime:       role.Ctime,
		Utime:       role.Utime,
	}
}

func (r *rbacRepository) toResourceEntity(resource domain.Resource) dao.Resource {
	return dao.Resource{
		ID:          resource.ID,
		BizID:       resource.BizID,
		Type:        resource.Type,
		Key:         resource.Key,
		Name:        resource.Name,
		Description: resource.Description,
		Ctime:       resource.Ctime,
		Utime:       resource.Utime,
	}
}

func (r *rbacRepository) toResourceDomain(resource dao.Resource) domain.Resource {
	return domain.Resource{
		ID:          resource.ID,
		BizID:       resource.BizID,
		Type:        resource.Type,
		Key:         resource.Key,
		Name:        resource.Name,
		Description: resource.Description,
		Ctime:       resource.Ctime,
		Utime:       resource.Utime,
	}
}

//nolint:dupl // 虽然结构相似，但执行的是互逆的转换操作
func (r *rbacRepository) toPermissionEntity(permission domain.Permission) dao.Permission {
	var action dao.ActionType
	switch permission.Action {
	case domain.ActionTypeCreate:
		action = dao.ActionTypeCreate
	case domain.ActionTypeRead:
		action = dao.ActionTypeRead
	case domain.ActionTypeWrite:
		action = dao.ActionTypeUpdate
	case domain.ActionTypeDelete:
		action = dao.ActionTypeDelete
	case domain.ActionTypeExecute:
		action = dao.ActionTypeExecute
	case domain.ActionTypeExport:
		action = dao.ActionTypeExport
	case domain.ActionTypeImport:
		action = dao.ActionTypeImport
	default:
		action = dao.ActionTypeRead
	}

	return dao.Permission{
		ID:           permission.ID,
		BizID:        permission.BizID,
		Name:         permission.Name,
		Description:  permission.Description,
		ResourceID:   permission.ResourceID,
		ResourceType: permission.ResourceType,
		ResourceKey:  permission.ResourceKey,
		Action:       action,
		Ctime:        permission.Ctime,
		Utime:        permission.Utime,
	}
}

//nolint:dupl // 虽然结构相似，但执行的是互逆的转换操作
func (r *rbacRepository) toPermissionDomain(permission dao.Permission) domain.Permission {
	var action domain.ActionType
	switch permission.Action {
	case dao.ActionTypeCreate:
		action = domain.ActionTypeCreate
	case dao.ActionTypeRead:
		action = domain.ActionTypeRead
	case dao.ActionTypeUpdate:
		action = domain.ActionTypeWrite
	case dao.ActionTypeDelete:
		action = domain.ActionTypeDelete
	case dao.ActionTypeExecute:
		action = domain.ActionTypeExecute
	case dao.ActionTypeExport:
		action = domain.ActionTypeExport
	case dao.ActionTypeImport:
		action = domain.ActionTypeImport
	default:
		action = domain.ActionTypeRead
	}

	return domain.Permission{
		ID:           permission.ID,
		BizID:        permission.BizID,
		Name:         permission.Name,
		Description:  permission.Description,
		ResourceID:   permission.ResourceID,
		ResourceType: permission.ResourceType,
		ResourceKey:  permission.ResourceKey,
		Action:       action,
		Ctime:        permission.Ctime,
		Utime:        permission.Utime,
	}
}

func (r *rbacRepository) toUserRoleDomain(userRole dao.UserRole) domain.UserRole {
	var roleType domain.RoleType
	switch userRole.RoleType {
	case "system":
		roleType = domain.RoleTypeSystem
	case "custom":
		roleType = domain.RoleTypeCustom
	case "temporary":
		roleType = domain.RoleTypeTemporary
	default:
		roleType = domain.RoleTypeCustom
	}

	return domain.UserRole{
		ID:        userRole.ID,
		BizID:     userRole.BizID,
		UserID:    userRole.UserID,
		RoleID:    userRole.RoleID,
		RoleName:  userRole.RoleName,
		RoleType:  roleType,
		StartTime: userRole.StartTime,
		EndTime:   userRole.EndTime,
		Ctime:     userRole.Ctime,
		Utime:     userRole.Utime,
	}
}

// 业务配置相关方法实现
func (r *rbacRepository) GetBusinessConfig(ctx context.Context, id int64) (domain.BusinessConfig, error) {
	config, err := r.businessConfigDAO.GetByID(ctx, id)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return r.toBusinessConfigDomain(config), nil
}

func (r *rbacRepository) GetBusinessConfigs(ctx context.Context, ids []int64) (map[int64]domain.BusinessConfig, error) {
	configMap, err := r.businessConfigDAO.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]domain.BusinessConfig, len(configMap))
	for id, config := range configMap {
		result[id] = r.toBusinessConfigDomain(config)
	}
	return result, nil
}

func (r *rbacRepository) SaveBusinessConfig(ctx context.Context, config domain.BusinessConfig) (domain.BusinessConfig, error) {
	daoConfig := r.toBusinessConfigEntity(config)
	saved, err := r.businessConfigDAO.SaveConfig(ctx, daoConfig)
	if err != nil {
		return domain.BusinessConfig{}, err
	}
	return r.toBusinessConfigDomain(saved), nil
}

func (r *rbacRepository) DeleteBusinessConfig(ctx context.Context, id int64) error {
	return r.businessConfigDAO.Delete(ctx, id)
}

func (r *rbacRepository) ListBusinessConfigs(ctx context.Context, offset, limit int) ([]domain.BusinessConfig, int, error) {
	configs, err := r.businessConfigDAO.Find(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	count := len(configs)
	result := make([]domain.BusinessConfig, 0, count)
	for i := range configs {
		result = append(result, r.toBusinessConfigDomain(configs[i]))
	}
	return result, count, nil
}

func (r *rbacRepository) toBusinessConfigEntity(config domain.BusinessConfig) dao.BusinessConfig {
	return dao.BusinessConfig{
		ID:        config.ID,
		Name:      config.Name,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		RateLimit: config.RateLimit,
		Token:     config.Token,
		Ctime:     config.Ctime,
		Utime:     config.Utime,
	}
}

func (r *rbacRepository) toBusinessConfigDomain(config dao.BusinessConfig) domain.BusinessConfig {
	return domain.BusinessConfig{
		ID:        config.ID,
		Name:      config.Name,
		OwnerID:   config.OwnerID,
		OwnerType: config.OwnerType,
		RateLimit: config.RateLimit,
		Token:     config.Token,
		Ctime:     config.Ctime,
		Utime:     config.Utime,
	}
}

// 角色权限相关方法实现
func (r *rbacRepository) GrantRolePermission(ctx context.Context, bizID, roleID, permissionID, startTime, endTime int64) (domain.RolePermission, error) {
	// 获取角色信息
	role, err := r.roleDAO.GetByID(ctx, roleID)
	if err != nil {
		return domain.RolePermission{}, err
	}

	// 获取权限信息
	permission, err := r.permissionDAO.GetByID(ctx, permissionID)
	if err != nil {
		return domain.RolePermission{}, err
	}

	rolePermission := dao.RolePermission{
		BizID:        bizID,
		RoleID:       roleID,
		PermissionID: permissionID,
		RoleName:     role.Name,
		RoleType:     role.Type,
		ResourceType: permission.ResourceType,
		ResourceKey:  permission.ResourceKey,
		Action:       permission.Action,
	}

	created, err := r.rolePermissionDAO.Create(ctx, rolePermission)
	if err != nil {
		return domain.RolePermission{}, err
	}

	result := r.toRolePermissionDomain(created)
	// 单独设置StartTime和EndTime，因为dao结构体中没有这些字段
	result.StartTime = startTime
	result.EndTime = endTime

	return result, nil
}

func (r *rbacRepository) RevokeRolePermission(ctx context.Context, bizID, roleID, permissionID int64) error {
	return r.rolePermissionDAO.DeleteByRoleIDAndPermissionID(ctx, bizID, roleID, permissionID)
}

func (r *rbacRepository) ListRolePermissions(ctx context.Context, bizID, roleID int64, offset, limit int) ([]domain.RolePermission, int, error) {
	// 查找角色所有权限
	rolePermissions, err := r.rolePermissionDAO.FindByRoleID(ctx, bizID, roleID)
	if err != nil {
		return nil, 0, err
	}

	// 总数
	count := len(rolePermissions)

	// 处理分页
	startIndex := offset
	endIndex := offset + limit
	if startIndex >= count {
		return []domain.RolePermission{}, 0, nil
	}
	if endIndex > count {
		endIndex = count
	}

	// 转换为domain模型
	paginatedPermissions := rolePermissions[startIndex:endIndex]
	domainRolePermissions := make([]domain.RolePermission, 0, len(paginatedPermissions))
	for i := range paginatedPermissions {
		domainRolePermissions = append(domainRolePermissions, r.toRolePermissionDomain(paginatedPermissions[i]))
	}

	return domainRolePermissions, count, nil
}

func (r *rbacRepository) toRolePermissionDomain(rolePermission dao.RolePermission) domain.RolePermission {
	// 获取权限名称，这里简化处理
	permissionName := ""

	return domain.RolePermission{
		ID:             rolePermission.ID,
		BizID:          rolePermission.BizID,
		RoleID:         rolePermission.RoleID,
		PermissionID:   rolePermission.PermissionID,
		PermissionName: permissionName,
		PermissionType: rolePermission.ResourceType,
		StartTime:      0, // dao结构体中没有这些字段，由上层设置
		EndTime:        0,
	}
}

// 角色包含关系相关方法实现
func (r *rbacRepository) CreateRoleInclusion(ctx context.Context, bizID, includingRoleID, includedRoleID int64) (domain.RoleInclusion, error) {
	// 获取包含者角色信息
	includingRole, err := r.roleDAO.GetByID(ctx, includingRoleID)
	if err != nil {
		return domain.RoleInclusion{}, err
	}

	// 获取被包含角色信息
	includedRole, err := r.roleDAO.GetByID(ctx, includedRoleID)
	if err != nil {
		return domain.RoleInclusion{}, err
	}

	roleInclusion := dao.RoleInclusion{
		BizID:             bizID,
		IncludingRoleID:   includingRoleID,
		IncludingRoleType: includingRole.Type,
		IncludingRoleName: includingRole.Name,
		IncludedRoleID:    includedRoleID,
		IncludedRoleType:  includedRole.Type,
		IncludedRoleName:  includedRole.Name,
	}
	created, err := r.roleInclusionDAO.Create(ctx, roleInclusion)
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	return r.toRoleInclusionDomain(created), nil
}

func (r *rbacRepository) GetRoleInclusion(ctx context.Context, id int64) (domain.RoleInclusion, error) {
	roleInclusion, err := r.roleInclusionDAO.GetByID(ctx, id)
	if err != nil {
		return domain.RoleInclusion{}, err
	}
	return r.toRoleInclusionDomain(roleInclusion), nil
}

func (r *rbacRepository) DeleteRoleInclusion(ctx context.Context, bizID, includingRoleID, includedRoleID int64) error {
	return r.roleInclusionDAO.Delete(ctx, bizID, includingRoleID, includedRoleID)
}

func (r *rbacRepository) ListRoleInclusions(ctx context.Context, bizID, roleID int64, isIncluding bool, offset, limit int) ([]domain.RoleInclusion, int, error) {
	var roleInclusions []dao.RoleInclusion
	var err error

	// 根据是否是包含关系查询
	if isIncluding {
		roleInclusions, err = r.roleInclusionDAO.FindByIncludingRoleID(ctx, bizID, roleID)
	} else {
		// 这里应该使用FindByIncludedRoleID方法，但目前DAO接口中没有这个方法
		// 暂时使用变通方法，实际项目中应添加该方法
		roleInclusions, err = r.roleInclusionDAO.FindByBizIDAndRoleID(ctx, bizID, roleID)
		// 过滤出roleID是被包含角色的记录
		var filteredInclusions []dao.RoleInclusion
		for _, inclusion := range roleInclusions {
			if inclusion.IncludedRoleID == roleID {
				filteredInclusions = append(filteredInclusions, inclusion)
			}
		}
		roleInclusions = filteredInclusions
	}

	if err != nil {
		return nil, 0, err
	}

	// 处理分页
	startIndex := offset
	endIndex := offset + limit
	if startIndex >= len(roleInclusions) {
		return []domain.RoleInclusion{}, 0, nil
	}
	if endIndex > len(roleInclusions) {
		endIndex = len(roleInclusions)
	}

	// 总数
	count := len(roleInclusions)

	// 转换为domain模型
	paginatedInclusions := roleInclusions[startIndex:endIndex]
	domainRoleInclusions := make([]domain.RoleInclusion, 0, len(paginatedInclusions))
	for i := range paginatedInclusions {
		domainRoleInclusions = append(domainRoleInclusions, r.toRoleInclusionDomain(paginatedInclusions[i]))
	}

	return domainRoleInclusions, count, nil
}

func (r *rbacRepository) toRoleInclusionDomain(roleInclusion dao.RoleInclusion) domain.RoleInclusion {
	var includingRoleType, includedRoleType domain.RoleType

	switch roleInclusion.IncludingRoleType {
	case dao.RoleTypeSystem:
		includingRoleType = domain.RoleTypeSystem
	case dao.RoleTypeCustom:
		includingRoleType = domain.RoleTypeCustom
	case dao.RoleTypeTemporary:
		includingRoleType = domain.RoleTypeTemporary
	default:
		includingRoleType = domain.RoleTypeCustom
	}

	switch roleInclusion.IncludedRoleType {
	case dao.RoleTypeSystem:
		includedRoleType = domain.RoleTypeSystem
	case dao.RoleTypeCustom:
		includedRoleType = domain.RoleTypeCustom
	case dao.RoleTypeTemporary:
		includedRoleType = domain.RoleTypeTemporary
	default:
		includedRoleType = domain.RoleTypeCustom
	}

	return domain.RoleInclusion{
		ID:                roleInclusion.ID,
		BizID:             roleInclusion.BizID,
		IncludingRoleID:   roleInclusion.IncludingRoleID,
		IncludingRoleType: includingRoleType,
		IncludingRoleName: roleInclusion.IncludingRoleName,
		IncludedRoleID:    roleInclusion.IncludedRoleID,
		IncludedRoleType:  includedRoleType,
		IncludedRoleName:  roleInclusion.IncludedRoleName,
		IsIncluding:       true, // 默认为包含关系
		Ctime:             roleInclusion.Ctime,
		Utime:             roleInclusion.Utime,
	}
}

// 用户权限相关方法实现
func (r *rbacRepository) GrantUserPermission(ctx context.Context, bizID, userID, permissionID int64, effect string, startTime, endTime int64) (domain.UserPermission, error) {
	// 获取权限信息
	permission, err := r.permissionDAO.GetByID(ctx, permissionID)
	if err != nil {
		return domain.UserPermission{}, err
	}

	// 获取资源信息
	resource, err := r.resourceDAO.GetByID(ctx, permission.ResourceID)
	if err != nil {
		return domain.UserPermission{}, err
	}

	userPermission := dao.UserPermission{
		BizID:          bizID,
		UserID:         userID,
		PermissionID:   permissionID,
		PermissionName: permission.Name,
		ResourceType:   resource.Type,
		ResourceKey:    resource.Key,
		ResourceName:   resource.Name,
		Action:         permission.Action,
		StartTime:      startTime,
		EndTime:        endTime,
		Effect:         dao.EffectType(effect),
	}
	created, err := r.userPermissionDAO.Create(ctx, userPermission)
	if err != nil {
		return domain.UserPermission{}, err
	}
	return r.toUserPermissionDomain(created), nil
}

func (r *rbacRepository) RevokeUserPermission(ctx context.Context, bizID, userID, permissionID int64) error {
	return r.userPermissionDAO.DeleteByUserIDAndPermissionID(ctx, bizID, userID, permissionID)
}

func (r *rbacRepository) ListUserPermissions(ctx context.Context, bizID, userID int64, offset, limit int, onlyValid bool) ([]domain.UserPermission, int, error) {
	var userPermissions []dao.UserPermission
	var err error

	if onlyValid {
		// 获取当前有效的权限
		currentTime := time.Now().UnixMilli()
		userPermissions, err = r.userPermissionDAO.FindValidPermissions(ctx, bizID, userID, currentTime)
	} else {
		// 获取所有权限
		userPermissions, err = r.userPermissionDAO.FindByBizIDAndUserID(ctx, bizID, userID)
	}

	if err != nil {
		return nil, 0, err
	}

	// 处理分页
	startIndex := offset
	endIndex := offset + limit
	if startIndex >= len(userPermissions) {
		return []domain.UserPermission{}, 0, nil
	}
	if endIndex > len(userPermissions) {
		endIndex = len(userPermissions)
	}

	// 总数
	count := len(userPermissions)

	// 转换为domain模型
	paginatedPermissions := userPermissions[startIndex:endIndex]
	domainUserPermissions := make([]domain.UserPermission, 0, len(paginatedPermissions))
	for i := range paginatedPermissions {
		domainUserPermissions = append(domainUserPermissions, r.toUserPermissionDomain(paginatedPermissions[i]))
	}

	return domainUserPermissions, count, nil
}

func (r *rbacRepository) toUserPermissionDomain(userPermission dao.UserPermission) domain.UserPermission {
	var action domain.ActionType
	switch userPermission.Action {
	case dao.ActionTypeCreate:
		action = domain.ActionTypeCreate
	case dao.ActionTypeRead:
		action = domain.ActionTypeRead
	case dao.ActionTypeUpdate:
		action = domain.ActionTypeWrite
	case dao.ActionTypeDelete:
		action = domain.ActionTypeDelete
	case dao.ActionTypeExecute:
		action = domain.ActionTypeExecute
	case dao.ActionTypeExport:
		action = domain.ActionTypeExport
	case dao.ActionTypeImport:
		action = domain.ActionTypeImport
	default:
		action = domain.ActionTypeRead
	}

	return domain.UserPermission{
		ID:             userPermission.ID,
		BizID:          userPermission.BizID,
		UserID:         userPermission.UserID,
		PermissionID:   userPermission.PermissionID,
		PermissionName: userPermission.PermissionName,
		ResourceType:   userPermission.ResourceType,
		ResourceKey:    userPermission.ResourceKey,
		ResourceName:   userPermission.ResourceName,
		Action:         action,
		StartTime:      userPermission.StartTime,
		EndTime:        userPermission.EndTime,
		Effect:         string(userPermission.Effect),
		OnlyValid:      false, // 默认为false
		Ctime:          userPermission.Ctime,
		Utime:          userPermission.Utime,
	}
}
