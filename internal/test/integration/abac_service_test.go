//go:build e2e

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitee.com/flycash/permission-platform/internal/domain"
	"gitee.com/flycash/permission-platform/internal/repository"
	"gitee.com/flycash/permission-platform/internal/repository/dao"
	"gitee.com/flycash/permission-platform/internal/test/integration/ioc/abac"
	testioc "gitee.com/flycash/permission-platform/internal/test/ioc"
	"github.com/ego-component/egorm"
	"github.com/stretchr/testify/suite"
)

type AbacServiceSuite struct {
	suite.Suite
	valRepo        repository.AttributeValueRepository
	definitionRepo repository.AttributeDefinitionRepository
	permissionRepo repository.PermissionRepository
	resourceRepo   repository.ResourceRepository
	policyRepo     repository.PolicyRepo
	db             *egorm.Component
}

func (s *AbacServiceSuite) SetupSuite() {
	db := testioc.InitDBAndTables()
	svc := abac.Init(db)
	s.definitionRepo = svc.DefinitionRepo
	s.valRepo = svc.ValRepo
	s.policyRepo = svc.PolicyRepo
	s.permissionRepo = svc.PermissionRepo
	s.db = db
}

func (s *AbacServiceSuite) clearBizVal(bizId int64) {
	t := s.T()
	t.Helper()
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.Policy{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.PolicyRule{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.SubjectAttributeValue{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.ResourceAttributeValue{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.EnvironmentAttributeValue{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.PermissionPolicy{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.Permission{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.AttributeDefinition{})
	s.db.WithContext(t.Context()).Where("biz_id = ?", bizId).Delete(&dao.Resource{})
}

func (s *AbacServiceSuite) TestAttributeDefinition_Save() {
	ctx := context.Background()
	bizID := int64(2)
	defer s.clearBizVal(bizID)

	// 定义保存方法
	save := func(def domain.AttributeDefinition) (int64, error) {
		return s.definitionRepo.Save(ctx, bizID, def)
	}

	tests := []struct {
		name    string
		before  func() (domain.AttributeDefinition, int64)
		after   func(def domain.AttributeDefinition) domain.AttributeDefinition
		wantErr bool
		check   func(t *testing.T, def domain.AttributeDefinition, id int64)
	}{
		{
			name: "新增属性定义",
			before: func() (domain.AttributeDefinition, int64) {
				return domain.AttributeDefinition{}, 0
			},
			after: func(def domain.AttributeDefinition) domain.AttributeDefinition {
				return domain.AttributeDefinition{
					Name:           "age",
					Description:    "用户年龄",
					DataType:       domain.DataTypeNumber,
					EntityType:     domain.EntityTypeSubject,
					ValidationRule: "^[0-9]+$",
				}
			},
			wantErr: false,
			check: func(t *testing.T, def domain.AttributeDefinition, id int64) {
				s.Greater(id, int64(0))
				found, err := s.definitionRepo.First(ctx, bizID, id)
				s.NoError(err)
				s.Equal(def.Name, found.Name)
				s.Equal(def.Description, found.Description)
				s.Equal(def.DataType, found.DataType)
				s.Equal(def.EntityType, found.EntityType)
				s.Equal(def.ValidationRule, found.ValidationRule)
			},
		},
		{
			name: "更新属性定义",
			before: func() (domain.AttributeDefinition, int64) {
				def := domain.AttributeDefinition{
					Name:           "new_age",
					Description:    "用户年龄",
					DataType:       domain.DataTypeNumber,
					EntityType:     domain.EntityTypeSubject,
					ValidationRule: "^[0-9]+$",
				}
				id, err := save(def)
				s.NoError(err)
				s.Greater(id, int64(0))
				return def, id
			},
			after: func(def domain.AttributeDefinition) domain.AttributeDefinition {
				def.Description = "更新后的用户年龄描述"
				def.DataType = domain.DataTypeString
				def.EntityType = domain.EntityTypeResource
				def.ValidationRule = "^[a-zA-Z]+$"
				return def
			},
			wantErr: false,
			check: func(t *testing.T, def domain.AttributeDefinition, id int64) {
				found, err := s.definitionRepo.First(ctx, bizID, id)
				s.NoError(err)
				s.Equal("更新后的用户年龄描述", found.Description)
				s.Equal(domain.DataTypeString, found.DataType)
				s.Equal(domain.EntityTypeResource, found.EntityType)
				s.Equal("^[a-zA-Z]+$", found.ValidationRule)
				s.Equal(def.Name, found.Name) // 名称不应该改变
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			def, id := tt.before()
			updatedDef := tt.after(def)
			updatedID, err := save(updatedDef)
			if tt.wantErr {
				s.Error(err)
				tt.check(t, def, id)
				return
			}
			s.NoError(err)
			if id > 0 {
				s.Equal(id, updatedID) // 验证是更新而不是新建
			}
			tt.check(t, updatedDef, updatedID)
		})
	}
}

func (s *AbacServiceSuite) TestAttributeDefinition_First() {
	ctx := context.Background()
	bizID := int64(3)
	defer s.clearBizVal(bizID)

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "test_attr",
		Description:    "测试属性",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: ".*",
	}
	id, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(id, int64(0))

	// 测试查询存在的属性定义
	found, err := s.definitionRepo.First(ctx, bizID, id)
	s.NoError(err)
	s.Equal(def.Name, found.Name)
	s.Equal(def.Description, found.Description)
	s.Equal(def.DataType, found.DataType)
	s.Equal(def.EntityType, found.EntityType)
	s.Equal(def.ValidationRule, found.ValidationRule)

	// 测试查询不存在的属性定义
	_, err = s.definitionRepo.First(ctx, bizID, 999999)
	s.Error(err)
}

func (s *AbacServiceSuite) TestAttributeDefinition_Del() {
	ctx := context.Background()
	bizID := int64(4)

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "to_delete",
		Description:    "待删除的属性",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: ".*",
	}
	id, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(id, int64(0))

	// 测试删除存在的属性定义
	err = s.definitionRepo.Del(ctx, bizID, id)
	s.NoError(err)

	// 验证属性定义已被删除
	_, err = s.definitionRepo.First(ctx, bizID, id)
	s.Error(err)

	// 测试删除不存在的属性定义
	err = s.definitionRepo.Del(ctx, bizID, 999999)
	s.NoError(err) // 删除不存在的记录应该不返回错误
}

func (s *AbacServiceSuite) TestAttributeDefinition_Find() {
	ctx := context.Background()
	bizID := int64(5)

	// 清理可能存在的测试数据
	s.db.WithContext(ctx).Where("biz_id = ?", bizID).Delete(&dao.AttributeDefinition{})

	// 创建测试数据
	subjectDef := domain.AttributeDefinition{
		Name:           "subject_attr",
		Description:    "主体属性",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: ".*",
	}
	resourceDef := domain.AttributeDefinition{
		Name:           "resource_attr",
		Description:    "资源属性",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeResource,
		ValidationRule: "^[0-9]+$",
	}
	envDef := domain.AttributeDefinition{
		Name:           "env_attr",
		Description:    "环境属性",
		DataType:       domain.DataTypeBoolean,
		EntityType:     domain.EntityTypeEnvironment,
		ValidationRule: "",
	}

	// 保存测试数据
	_, err := s.definitionRepo.Save(ctx, bizID, subjectDef)
	s.NoError(err)
	_, err = s.definitionRepo.Save(ctx, bizID, resourceDef)
	s.NoError(err)
	_, err = s.definitionRepo.Save(ctx, bizID, envDef)
	s.NoError(err)

	// 测试查询所有属性定义
	bizDef, err := s.definitionRepo.Find(ctx, bizID)
	s.NoError(err)
	s.Equal(bizID, bizDef.BizID)
	s.Len(bizDef.SubjectAttrDefs, 1)
	s.Len(bizDef.ResourceAttrDefs, 1)
	s.Len(bizDef.EnvironmentAttrDefs, 1)

	// 验证主体属性
	s.Equal(subjectDef.Name, bizDef.SubjectAttrDefs[0].Name)
	s.Equal(subjectDef.Description, bizDef.SubjectAttrDefs[0].Description)
	s.Equal(subjectDef.DataType, bizDef.SubjectAttrDefs[0].DataType)
	s.Equal(subjectDef.EntityType, bizDef.SubjectAttrDefs[0].EntityType)
	s.Equal(subjectDef.ValidationRule, bizDef.SubjectAttrDefs[0].ValidationRule)

	// 验证资源属性
	s.Equal(resourceDef.Name, bizDef.ResourceAttrDefs[0].Name)
	s.Equal(resourceDef.Description, bizDef.ResourceAttrDefs[0].Description)
	s.Equal(resourceDef.DataType, bizDef.ResourceAttrDefs[0].DataType)
	s.Equal(resourceDef.EntityType, bizDef.ResourceAttrDefs[0].EntityType)
	s.Equal(resourceDef.ValidationRule, bizDef.ResourceAttrDefs[0].ValidationRule)

	// 验证环境属性
	s.Equal(envDef.Name, bizDef.EnvironmentAttrDefs[0].Name)
	s.Equal(envDef.Description, bizDef.EnvironmentAttrDefs[0].Description)
	s.Equal(envDef.DataType, bizDef.EnvironmentAttrDefs[0].DataType)
	s.Equal(envDef.EntityType, bizDef.EnvironmentAttrDefs[0].EntityType)
	s.Equal(envDef.ValidationRule, bizDef.EnvironmentAttrDefs[0].ValidationRule)

	// 测试查询不存在的业务ID
	_, err = s.definitionRepo.Find(ctx, 999999)
	s.NoError(err) // 应该返回空结果而不是错误
}

func (s *AbacServiceSuite) TestAttributeSubjectValue_Save() {
	ctx := context.Background()
	bizID := int64(6)
	subjectID := int64(1001)
	defer s.clearBizVal(bizID)

	// 定义保存方法
	save := func(val domain.AttributeValue) (int64, error) {
		return s.valRepo.SaveSubjectValue(ctx, bizID, subjectID, val)
	}

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "age",
		Description:    "用户年龄",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: "^[0-9]+$",
	}
	defID, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(defID, int64(0))

	// 更新属性定义ID
	def.ID = defID

	tests := []struct {
		name    string
		before  func() (domain.AttributeValue, int64)
		after   func(val domain.AttributeValue) domain.AttributeValue
		wantErr bool
		check   func(t *testing.T, val domain.AttributeValue, id int64)
	}{
		{
			name: "新增主体属性值",
			before: func() (domain.AttributeValue, int64) {
				return domain.AttributeValue{}, 0
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				return domain.AttributeValue{
					Definition: def,
					Value:      "25",
				}
			},
			wantErr: false,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				s.Greater(id, int64(0))
				// 验证保存的属性值
				subjectObj, err := s.valRepo.FindSubjectValue(ctx, bizID, subjectID)
				s.NoError(err)
				s.Equal(subjectID, subjectObj.ID)
				s.Equal(bizID, subjectObj.BizID)
				s.Len(subjectObj.AttributeValues, 1)
				s.Equal("25", subjectObj.AttributeValues[0].Value)
				s.Equal(defID, subjectObj.AttributeValues[0].Definition.ID)
			},
		},
		{
			name: "更新主体属性值",
			before: func() (domain.AttributeValue, int64) {
				val := domain.AttributeValue{
					Definition: def,
					Value:      "25",
				}
				id, err := save(val)
				s.NoError(err)
				s.Greater(id, int64(0))
				return val, id
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				val.Value = "30"
				return val
			},
			wantErr: false,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				// 验证更新后的属性值
				subjectObj, err := s.valRepo.FindSubjectValue(ctx, bizID, subjectID)
				s.NoError(err)
				s.Equal(subjectID, subjectObj.ID)
				s.Equal(bizID, subjectObj.BizID)
				s.Len(subjectObj.AttributeValues, 1)
				s.Equal("30", subjectObj.AttributeValues[0].Value)
				s.Equal(defID, subjectObj.AttributeValues[0].Definition.ID)
			},
		},
		{
			name: "保存无效的属性值",
			before: func() (domain.AttributeValue, int64) {
				val := domain.AttributeValue{
					Definition: def,
					Value:      "25",
				}
				id, err := save(val)
				s.NoError(err)
				s.Greater(id, int64(0))
				return val, id
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				val.Value = "abc" // 不符合数字类型的验证规则
				return val
			},
			wantErr: true,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				// 验证属性值没有被更新
				subjectObj, err := s.valRepo.FindSubjectValue(ctx, bizID, subjectID)
				s.NoError(err)
				s.Equal(subjectID, subjectObj.ID)
				s.Equal(bizID, subjectObj.BizID)
				s.Len(subjectObj.AttributeValues, 1)
				s.Equal("25", subjectObj.AttributeValues[0].Value) // 应该保持原来的值
				s.Equal(defID, subjectObj.AttributeValues[0].Definition.ID)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			val, id := tt.before()
			updatedVal := tt.after(val)
			updatedID, err := save(updatedVal)
			if tt.wantErr {
				s.Error(err)
				tt.check(t, val, id)
				return
			}
			s.NoError(err)
			if id > 0 {
				s.Equal(id, updatedID) // 验证是更新而不是新建
			}
			tt.check(t, updatedVal, updatedID)
		})
	}
}

func (s *AbacServiceSuite) TestAttributeSubjectValue_Delete() {
	ctx := s.T().Context()
	bizID := int64(7)
	subjectID := int64(1002)
	defer s.clearBizVal(bizID)

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "age",
		Description:    "用户年龄",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: "^[0-9]+$",
	}
	defID, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(defID, int64(0))

	// 更新属性定义ID
	def.ID = defID

	// 保存一个属性值
	val := domain.AttributeValue{
		Definition: def,
		Value:      "25",
	}
	id, err := s.valRepo.SaveSubjectValue(ctx, bizID, subjectID, val)
	s.NoError(err)
	s.Greater(id, int64(0))

	// 测试删除存在的属性值
	err = s.valRepo.DeleteSubjectValue(ctx, id)
	s.NoError(err)

	// 验证属性值已被删除
	var res dao.SubjectAttributeValue
	err = s.db.WithContext(ctx).
		Where("id = ?", id).First(&res).Error
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)
	subjectObj, err := s.valRepo.FindSubjectValue(ctx, bizID, subjectID)
	s.NoError(err)
	s.Equal(subjectID, subjectObj.ID)
	s.Equal(bizID, subjectObj.BizID)
	s.Empty(subjectObj.AttributeValues)

	// 测试删除不存在的属性值
	err = s.valRepo.DeleteSubjectValue(ctx, 999999)
	s.NoError(err) // 删除不存在的记录应该不返回错误
}

func (s *AbacServiceSuite) TestAttributeSubjectValue_FindWithDefinition() {
	ctx := context.Background()
	bizID := int64(8)
	subjectID := int64(1003)
	defer s.clearBizVal(bizID)

	// 创建多个属性定义
	ageDef := domain.AttributeDefinition{
		Name:           "age",
		Description:    "用户年龄",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: "^[0-9]+$",
	}
	ageDefID, err := s.definitionRepo.Save(ctx, bizID, ageDef)
	s.NoError(err)
	s.Greater(ageDefID, int64(0))
	ageDef.ID = ageDefID

	nameDef := domain.AttributeDefinition{
		Name:           "name",
		Description:    "用户名称",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeSubject,
		ValidationRule: ".*",
	}
	nameDefID, err := s.definitionRepo.Save(ctx, bizID, nameDef)
	s.NoError(err)
	s.Greater(nameDefID, int64(0))
	nameDef.ID = nameDefID

	// 保存多个属性值
	ageVal := domain.AttributeValue{
		Definition: ageDef,
		Value:      "25",
	}
	_, err = s.valRepo.SaveSubjectValue(ctx, bizID, subjectID, ageVal)
	s.NoError(err)

	nameVal := domain.AttributeValue{
		Definition: nameDef,
		Value:      "张三",
	}
	_, err = s.valRepo.SaveSubjectValue(ctx, bizID, subjectID, nameVal)
	s.NoError(err)

	// 测试查询带定义的属性值
	subjectObj, err := s.valRepo.FindSubjectValueWithDefinition(ctx, bizID, subjectID)
	s.NoError(err)
	s.Equal(subjectID, subjectObj.ID)
	s.Equal(bizID, subjectObj.BizID)
	s.Len(subjectObj.AttributeValues, 2)

	// 验证属性值
	values := make(map[string]string)
	for _, val := range subjectObj.AttributeValues {
		values[val.Definition.Name] = val.Value
	}
	s.Equal("25", values["age"])
	s.Equal("张三", values["name"])

	// 验证属性定义
	defs := make(map[string]domain.AttributeDefinition)
	for _, val := range subjectObj.AttributeValues {
		val.Definition.Ctime = 0
		val.Definition.Utime = 0
		defs[val.Definition.Name] = val.Definition
	}

	s.Equal(ageDef, defs["age"])
	s.Equal(nameDef, defs["name"])
}

func (s *AbacServiceSuite) TestAttributeResourceValue_Save() {
	ctx := context.Background()
	bizID := int64(9)
	resourceID := int64(2001)
	defer s.clearBizVal(bizID)

	// 定义保存方法
	save := func(val domain.AttributeValue) (int64, error) {
		return s.valRepo.SaveResourceValue(ctx, bizID, resourceID, val)
	}

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "size",
		Description:    "文件大小",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeResource,
		ValidationRule: "^[0-9]+$",
	}
	defID, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(defID, int64(0))
	def.ID = defID

	tests := []struct {
		name    string
		before  func() (domain.AttributeValue, int64)
		after   func(val domain.AttributeValue) domain.AttributeValue
		wantErr bool
		check   func(t *testing.T, val domain.AttributeValue, id int64)
	}{
		{
			name: "新增资源属性值",
			before: func() (domain.AttributeValue, int64) {
				return domain.AttributeValue{}, 0
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				return domain.AttributeValue{
					Definition: def,
					Value:      "1024",
				}
			},
			wantErr: false,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				s.Greater(id, int64(0))
				resourceObj, err := s.valRepo.FindResourceValue(ctx, bizID, resourceID)
				s.NoError(err)
				s.Equal(resourceID, resourceObj.ID)
				s.Equal(bizID, resourceObj.BizID)
				s.Len(resourceObj.AttributeValues, 1)
				s.Equal("1024", resourceObj.AttributeValues[0].Value)
				s.Equal(defID, resourceObj.AttributeValues[0].Definition.ID)
			},
		},
		{
			name: "更新资源属性值",
			before: func() (domain.AttributeValue, int64) {
				val := domain.AttributeValue{
					Definition: def,
					Value:      "1024",
				}
				id, err := save(val)
				s.NoError(err)
				s.Greater(id, int64(0))
				return val, id
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				val.Value = "2048"
				return val
			},
			wantErr: false,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				resourceObj, err := s.valRepo.FindResourceValue(ctx, bizID, resourceID)
				s.NoError(err)
				s.Equal(resourceID, resourceObj.ID)
				s.Equal(bizID, resourceObj.BizID)
				s.Len(resourceObj.AttributeValues, 1)
				s.Equal("2048", resourceObj.AttributeValues[0].Value)
				s.Equal(defID, resourceObj.AttributeValues[0].Definition.ID)
			},
		},
		{
			name: "保存无效的资源属性值",
			before: func() (domain.AttributeValue, int64) {
				val := domain.AttributeValue{
					Definition: def,
					Value:      "1024",
				}
				id, err := save(val)
				s.NoError(err)
				s.Greater(id, int64(0))
				return val, id
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				val.Value = "abc" // 不符合数字类型的验证规则
				return val
			},
			wantErr: true,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				resourceObj, err := s.valRepo.FindResourceValue(ctx, bizID, resourceID)
				s.NoError(err)
				s.Equal(resourceID, resourceObj.ID)
				s.Equal(bizID, resourceObj.BizID)
				s.Len(resourceObj.AttributeValues, 1)
				s.Equal("1024", resourceObj.AttributeValues[0].Value) // 应该保持原来的值
				s.Equal(defID, resourceObj.AttributeValues[0].Definition.ID)
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			val, id := tt.before()
			updatedVal := tt.after(val)
			updatedID, err := save(updatedVal)
			if tt.wantErr {
				s.Error(err)
				tt.check(t, val, id)
				return
			}
			s.NoError(err)
			if id > 0 {
				s.Equal(id, updatedID)
			}
			tt.check(t, updatedVal, updatedID)
		})
	}
}

func (s *AbacServiceSuite) TestAttributeResourceValue_Delete() {
	ctx := context.Background()
	bizID := int64(10)
	resourceID := int64(2002)
	defer s.clearBizVal(bizID)

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "size",
		Description:    "文件大小",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeResource,
		ValidationRule: "^[0-9]+$",
	}
	defID, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(defID, int64(0))
	def.ID = defID

	// 保存一个属性值
	val := domain.AttributeValue{
		Definition: def,
		Value:      "1024",
	}
	id, err := s.valRepo.SaveResourceValue(ctx, bizID, resourceID, val)
	s.NoError(err)

	// 测试删除存在的属性值
	err = s.valRepo.DeleteResourceValue(ctx, id)
	s.NoError(err)

	// 验证属性值已被删除
	var res dao.ResourceAttributeValue
	err = s.db.WithContext(ctx).
		Where("id = ?", id).First(&res).Error
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)

	// 测试删除不存在的属性值
	err = s.valRepo.DeleteResourceValue(ctx, 999999)
	s.NoError(err)
}

func (s *AbacServiceSuite) TestAttributeResourceValue_FindWithDefinition() {
	ctx := context.Background()
	bizID := int64(11)
	resourceID := int64(2003)
	defer s.clearBizVal(bizID)

	// 创建多个属性定义
	sizeDef := domain.AttributeDefinition{
		Name:           "size",
		Description:    "文件大小",
		DataType:       domain.DataTypeNumber,
		EntityType:     domain.EntityTypeResource,
		ValidationRule: "^[0-9]+$",
	}
	sizeDefID, err := s.definitionRepo.Save(ctx, bizID, sizeDef)
	s.NoError(err)
	s.Greater(sizeDefID, int64(0))
	sizeDef.ID = sizeDefID

	typeDef := domain.AttributeDefinition{
		Name:           "type",
		Description:    "文件类型",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeResource,
		ValidationRule: ".*",
	}
	typeDefID, err := s.definitionRepo.Save(ctx, bizID, typeDef)
	s.NoError(err)
	s.Greater(typeDefID, int64(0))
	typeDef.ID = typeDefID

	// 保存多个属性值
	sizeVal := domain.AttributeValue{
		Definition: sizeDef,
		Value:      "1024",
	}
	_, err = s.valRepo.SaveResourceValue(ctx, bizID, resourceID, sizeVal)
	s.NoError(err)

	typeVal := domain.AttributeValue{
		Definition: typeDef,
		Value:      "pdf",
	}
	_, err = s.valRepo.SaveResourceValue(ctx, bizID, resourceID, typeVal)
	s.NoError(err)

	// 测试查询带定义的属性值
	resourceObj, err := s.valRepo.FindResourceValueWithDefinition(ctx, bizID, resourceID)
	s.NoError(err)
	s.Equal(resourceID, resourceObj.ID)
	s.Equal(bizID, resourceObj.BizID)
	s.Len(resourceObj.AttributeValues, 2)

	// 验证属性值
	values := make(map[string]string)
	for _, val := range resourceObj.AttributeValues {
		val.Ctime = 0
		val.Utime = 0
		values[val.Definition.Name] = val.Value
	}
	s.Equal("1024", values["size"])
	s.Equal("pdf", values["type"])

	// 验证属性定义
	defs := make(map[string]domain.AttributeDefinition)
	for _, val := range resourceObj.AttributeValues {
		val.Definition.Ctime = 0
		val.Definition.Utime = 0
		defs[val.Definition.Name] = val.Definition

	}
	s.Equal(sizeDef, defs["size"])
	s.Equal(typeDef, defs["type"])
}

func (s *AbacServiceSuite) TestAttributeEnvironmentValue_Delete() {
	ctx := context.Background()
	bizID := int64(13)
	defer s.clearBizVal(bizID)

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "time",
		Description:    "访问时间",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeEnvironment,
		ValidationRule: ".*",
	}
	defID, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(defID, int64(0))
	def.ID = defID

	// 保存一个属性值
	val := domain.AttributeValue{
		Definition: def,
		Value:      "2024-01-01",
	}
	id, err := s.valRepo.SaveEnvironmentValue(ctx, bizID, val)
	s.NoError(err)

	// 测试删除存在的属性值
	err = s.valRepo.DeleteEnvironmentValue(ctx, id)
	s.NoError(err)

	// 验证属性值已被删除
	var res dao.EnvironmentAttributeValue
	err = s.db.WithContext(ctx).
		Where("id = ?", id).First(&res).Error
	assert.Equal(s.T(), gorm.ErrRecordNotFound, err)

	// 测试删除不存在的属性值
	err = s.valRepo.DeleteEnvironmentValue(ctx, 999999)
	s.NoError(err)
}

func (s *AbacServiceSuite) TestAttributeEnvironmentValue_FindWithDefinition() {
	ctx := context.Background()
	bizID := int64(14)
	defer s.clearBizVal(bizID)

	// 创建多个属性定义
	timeDef := domain.AttributeDefinition{
		Name:           "time",
		Description:    "访问时间",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeEnvironment,
		ValidationRule: ".*",
	}
	timeDefID, err := s.definitionRepo.Save(ctx, bizID, timeDef)
	s.NoError(err)
	s.Greater(timeDefID, int64(0))
	timeDef.ID = timeDefID

	ipDef := domain.AttributeDefinition{
		Name:           "ip",
		Description:    "访问IP",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeEnvironment,
		ValidationRule: ".*",
	}
	ipDefID, err := s.definitionRepo.Save(ctx, bizID, ipDef)
	s.NoError(err)
	s.Greater(ipDefID, int64(0))
	ipDef.ID = ipDefID

	// 保存多个属性值
	timeVal := domain.AttributeValue{
		Definition: timeDef,
		Value:      "2024-01-01",
	}
	_, err = s.valRepo.SaveEnvironmentValue(ctx, bizID, timeVal)
	s.NoError(err)

	ipVal := domain.AttributeValue{
		Definition: ipDef,
		Value:      "192.168.1.1",
	}
	_, err = s.valRepo.SaveEnvironmentValue(ctx, bizID, ipVal)
	s.NoError(err)

	// 测试查询带定义的属性值
	envObj, err := s.valRepo.FindEnvironmentValueWithDefinition(ctx, bizID)
	s.NoError(err)

	// 验证属性值
	values := make(map[string]string)
	for _, val := range envObj.AttributeValues {
		values[val.Definition.Name] = val.Value
	}
	s.Equal("2024-01-01", values["time"])
	s.Equal("192.168.1.1", values["ip"])

	// 验证属性定义
	defs := make(map[string]domain.AttributeDefinition)
	for _, val := range envObj.AttributeValues {
		val.Definition.Ctime = 0
		val.Definition.Utime = 0
		defs[val.Definition.Name] = val.Definition
	}
	s.Equal(timeDef, defs["time"])
	s.Equal(ipDef, defs["ip"])
}

func (s *AbacServiceSuite) TestAttributeEnvironmentValue_Save() {
	ctx := s.T().Context()
	bizID := int64(12)
	defer s.clearBizVal(bizID)

	// 定义保存方法
	save := func(val domain.AttributeValue) (int64, error) {
		return s.valRepo.SaveEnvironmentValue(ctx, bizID, val)
	}

	// 先创建一个属性定义
	def := domain.AttributeDefinition{
		Name:           "time",
		Description:    "访问时间",
		DataType:       domain.DataTypeString,
		EntityType:     domain.EntityTypeEnvironment,
		ValidationRule: ".*",
	}
	defID, err := s.definitionRepo.Save(ctx, bizID, def)
	s.NoError(err)
	s.Greater(defID, int64(0))
	def.ID = defID
	tests := []struct {
		name    string
		before  func() (domain.AttributeValue, int64)
		after   func(val domain.AttributeValue) domain.AttributeValue
		wantErr bool
		check   func(t *testing.T, val domain.AttributeValue, id int64)
	}{
		{
			name: "新增环境属性值",
			before: func() (domain.AttributeValue, int64) {
				return domain.AttributeValue{}, 0
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				return domain.AttributeValue{
					Definition: def,
					Value:      "2024-01-01",
				}
			},
			wantErr: false,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				s.Greater(id, int64(0))
				envObj, err := s.valRepo.FindEnvironmentValue(ctx, bizID)
				s.NoError(err)
				s.Equal(bizID, envObj.BizID)
				s.Len(envObj.AttributeValues, 1)
				s.Equal("2024-01-01", envObj.AttributeValues[0].Value)
				s.Equal(defID, envObj.AttributeValues[0].Definition.ID)
			},
		},
		{
			name: "更新环境属性值",
			before: func() (domain.AttributeValue, int64) {
				val := domain.AttributeValue{
					Definition: def,
					Value:      "2024-01-01",
				}
				id, err := save(val)
				s.NoError(err)
				s.Greater(id, int64(0))
				return val, id
			},
			after: func(val domain.AttributeValue) domain.AttributeValue {
				val.Value = "2024-01-02"
				return val
			},
			wantErr: false,
			check: func(t *testing.T, val domain.AttributeValue, id int64) {
				envObj, err := s.valRepo.FindEnvironmentValue(ctx, bizID)
				s.NoError(err)
				s.Equal(bizID, envObj.BizID)
				s.Len(envObj.AttributeValues, 1)
				s.Equal("2024-01-02", envObj.AttributeValues[0].Value)
				s.Equal(defID, envObj.AttributeValues[0].Definition.ID)
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			val, id := tt.before()

			updatedVal := tt.after(val)
			updatedID, err := save(updatedVal)
			if tt.wantErr {
				s.Error(err)
				tt.check(t, val, id)
				return
			}
			s.NoError(err)
			if id > 0 {
				s.Equal(id, updatedID)
			}
			tt.check(t, updatedVal, updatedID)
		})
	}
}

func (s *AbacServiceSuite) TestPolicy_Save() {
	ctx := s.T().Context()
	bizID := int64(15)
	defer s.clearBizVal(bizID)

	tests := []struct {
		name    string
		before  func() (domain.Policy, int64)
		after   func(policy domain.Policy) domain.Policy
		wantErr bool
		check   func(t *testing.T, policy domain.Policy, id int64)
	}{
		{
			name: "新增策略",
			before: func() (domain.Policy, int64) {
				return domain.Policy{}, 0
			},
			after: func(policy domain.Policy) domain.Policy {
				return domain.Policy{
					BizID:       bizID,
					ExecuteType: "priority",
					Name:        "测试策略",
					Description: "这是一个测试策略",
				}
			},
			wantErr: false,
			check: func(t *testing.T, policy domain.Policy, id int64) {
				s.Greater(id, int64(0))
				savedPolicy, err := s.policyRepo.First(ctx, bizID, id)
				s.NoError(err)
				s.Equal(bizID, savedPolicy.BizID)
				s.Equal("测试策略", savedPolicy.Name)
				s.Equal("这是一个测试策略", savedPolicy.Description)
				s.Equal("priority", string(savedPolicy.ExecuteType))
			},
		},
		{
			name: "更新策略描述",
			before: func() (domain.Policy, int64) {
				policy := domain.Policy{
					BizID:       bizID,
					Name:        "测试策略1",
					Description: "这是一个测试策略",
				}
				id, err := s.policyRepo.Save(ctx, policy)
				s.NoError(err)
				s.Greater(id, int64(0))
				return policy, id
			},
			after: func(policy domain.Policy) domain.Policy {
				policy.Description = "这是更新后的策略描述"
				return policy
			},
			wantErr: false,
			check: func(t *testing.T, policy domain.Policy, id int64) {
				savedPolicy, err := s.policyRepo.First(ctx, bizID, id)
				s.NoError(err)
				s.Equal(bizID, savedPolicy.BizID)
				s.Equal("测试策略1", savedPolicy.Name)             // 名称保持不变
				s.Equal("这是更新后的策略描述", savedPolicy.Description) // 描述已更新
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			policy, id := tt.before()
			updatedPolicy := tt.after(policy)
			updatedID, err := s.policyRepo.Save(ctx, updatedPolicy)
			if tt.wantErr {
				s.Error(err)
				tt.check(t, policy, id)
				return
			}
			s.NoError(err)
			if id > 0 {
				s.Equal(id, updatedID) // 验证是更新而不是新建
			}
			tt.check(t, updatedPolicy, updatedID)
		})
	}
}

func (s *AbacServiceSuite) TestPolicy_Delete() {
	ctx := s.T().Context()
	bizID := int64(16)
	defer s.clearBizVal(bizID)

	// 创建测试策略
	policy := domain.Policy{
		BizID:       bizID,
		Name:        "待删除策略",
		Description: "这是一个待删除的策略",
	}
	id, err := s.policyRepo.Save(ctx, policy)
	s.NoError(err)
	s.Greater(id, int64(0))

	// 添加策略规则
	rule := domain.PolicyRule{
		AttrDef: domain.AttributeDefinition{
			ID: 1,
		},
		Operator: ">",
		Value:    "18",
	}
	ruleID, err := s.policyRepo.SaveRule(ctx, bizID, id, rule)
	s.NoError(err)
	s.Greater(ruleID, int64(0))

	// 添加权限策略关联
	permissionID := int64(1001)
	err = s.policyRepo.SavePermissionPolicy(ctx, bizID, id, permissionID, "allow")
	s.NoError(err)

	// 验证关联数据已创建
	var permissionPolicy dao.PermissionPolicy
	err = s.db.WithContext(ctx).
		Where("biz_id = ? AND policy_id = ? AND permission_id = ?", bizID, id, permissionID).
		First(&permissionPolicy).Error
	s.NoError(err)

	var policyRule dao.PolicyRule
	err = s.db.WithContext(ctx).
		Where("id = ?", ruleID).
		First(&policyRule).Error
	s.NoError(err)

	// 测试删除策略
	err = s.policyRepo.Delete(ctx, bizID, id)
	s.NoError(err)

	// 验证策略已被删除
	_, err = s.policyRepo.First(ctx, bizID, id)
	s.Error(err)

	// 验证权限策略关联已被删除
	var count int64
	err = s.db.WithContext(ctx).
		Model(&dao.PermissionPolicy{}).
		Where("biz_id = ? AND policy_id = ?", bizID, id).
		Count(&count).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), count)

	// 验证策略规则已被删除
	err = s.db.WithContext(ctx).
		Where("id = ?", ruleID).
		First(&policyRule).Error
	s.Equal(gorm.ErrRecordNotFound, err)

	// 测试删除不存在的策略
	err = s.policyRepo.Delete(ctx, bizID, 999999)
	s.NoError(err) // 删除不存在的记录应该不返回错误
}

func (s *AbacServiceSuite) TestPolicy_SavePermissionPolicy() {
	ctx := s.T().Context()
	bizID := int64(21)
	defer s.clearBizVal(bizID)

	// 创建测试策略
	policy := domain.Policy{
		BizID:       bizID,
		Name:        "测试策略",
		Description: "这是一个测试策略",
	}
	id, err := s.policyRepo.Save(ctx, policy)
	s.NoError(err)
	s.Greater(id, int64(0))

	// 测试保存权限策略关联
	permissionID := int64(1003)
	err = s.policyRepo.SavePermissionPolicy(ctx, bizID, id, permissionID, domain.EffectAllow)
	s.NoError(err)

	// 验证关联已保存
	var res dao.PermissionPolicy
	err = s.db.WithContext(ctx).Where("biz_id = ? AND policy_id = ?", bizID, id).First(&res).Error
	require.NoError(s.T(), err)
	s.Equal(string(domain.EffectAllow), res.Effect)
}

func (s *AbacServiceSuite) TestPolicy_FindPolicies() {
	ctx := s.T().Context()
	bizID := int64(22)
	defer s.clearBizVal(bizID)

	// 创建多个测试策略
	for i := 1; i <= 5; i++ {
		policy := domain.Policy{
			BizID:       bizID,
			Name:        fmt.Sprintf("策略%d", i),
			Description: fmt.Sprintf("这是策略%d", i),
		}
		_, err := s.policyRepo.Save(ctx, policy)
		s.NoError(err)
	}

	// 测试分页查询
	total, policies, err := s.policyRepo.FindPolicies(ctx, bizID, 0, 3)
	s.NoError(err)
	s.Equal(int64(5), total)
	s.Len(policies, 3)

	// 验证第一页数据
	s.Equal("策略1", policies[0].Name)
	s.Equal("策略2", policies[1].Name)
	s.Equal("策略3", policies[2].Name)

	// 测试第二页
	total, policies, err = s.policyRepo.FindPolicies(ctx, bizID, 3, 3)
	s.NoError(err)
	s.Equal(int64(5), total)
	s.Len(policies, 2)

	// 验证第二页数据
	s.Equal("策略4", policies[0].Name)
	s.Equal("策略5", policies[1].Name)

	// 测试查询不存在的业务ID
	total, policies, err = s.policyRepo.FindPolicies(ctx, 999999, 0, 3)
	s.NoError(err)
	s.Equal(int64(0), total)
	s.Empty(policies)
}

func (s *AbacServiceSuite) TestPolicy_First() {
	ctx := s.T().Context()
	bizID := int64(17)
	defer s.clearBizVal(bizID)

	// 创建测试策略
	policy := domain.Policy{
		BizID:       bizID,
		Name:        "测试策略",
		Description: "这是一个测试策略",
	}
	id, err := s.policyRepo.Save(ctx, policy)
	s.NoError(err)
	s.Greater(id, int64(0))

	// 添加策略规则
	rule := domain.PolicyRule{
		AttrDef: domain.AttributeDefinition{
			ID: 1,
		},
		Operator: ">",
		Value:    "18",
	}
	ruleID, err := s.policyRepo.SaveRule(ctx, bizID, id, rule)
	s.NoError(err)
	s.Greater(ruleID, int64(0))
	// 添加权限策略关联
	permissionID := int64(1001)
	err = s.policyRepo.SavePermissionPolicy(ctx, bizID, id, permissionID, domain.EffectAllow)
	s.NoError(err)

	// 测试查询策略
	foundPolicy, err := s.policyRepo.First(ctx, bizID, id)
	s.NoError(err)
	s.Equal(bizID, foundPolicy.BizID)
	s.Equal("测试策略", foundPolicy.Name)
	s.Equal("这是一个测试策略", foundPolicy.Description)
	s.Equal("logic", string(foundPolicy.ExecuteType))

	// 验证策略规则
	s.Len(foundPolicy.Rules, 1)
	s.Equal(ruleID, foundPolicy.Rules[0].ID)
	s.Equal(int64(1), foundPolicy.Rules[0].AttrDef.ID)
	s.Equal(domain.RuleOperator(">"), foundPolicy.Rules[0].Operator)
	s.Equal("18", foundPolicy.Rules[0].Value)

	// 验证权限策略关联
	var permissionPolicy dao.PermissionPolicy
	err = s.db.WithContext(ctx).
		Where("biz_id = ? AND policy_id = ? AND permission_id = ?", bizID, id, permissionID).
		First(&permissionPolicy).Error
	s.NoError(err)
	s.Equal(string(domain.EffectAllow), permissionPolicy.Effect)
}

func TestAbacServiceSuite(t *testing.T) {
	suite.Run(t, new(AbacServiceSuite))
}
