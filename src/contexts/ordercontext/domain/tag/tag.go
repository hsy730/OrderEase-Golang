package tag

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
)

var (
	ErrTagNameRequired = errors.New("标签名称不能为空")
	ErrTagTooLong      = errors.New("标签名称不能超过50个字符")
	ErrTagColorInvalid = errors.New("颜色值无效，必须为7位十六进制格式（如 #FF5733）")
)

type TagID int

// Tag 标签实体
//
// 作为聚合根，Tag 负责：
//   - 管理标签自身的数据完整性
//   - 维护与商品的关联关系（通过 Repository）
//   - 执行删除前的业务规则验证
type Tag struct {
	id          TagID
	shopID      snowflake.ID
	name        string
	description string
	color       string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewTag(shopID snowflake.ID, name, description, color string) (*Tag, error) {
	t := &Tag{
		shopID:      shopID,
		name:        name,
		description: description,
		color:       color,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Tag) ID() TagID                { return t.id }
func (t *Tag) SetID(id TagID)            { t.id = id }
func (t *Tag) ShopID() snowflake.ID     { return t.shopID }
func (t *Tag) Name() string             { return t.name }
func (t *Tag) SetName(name string)      { t.name = name; t.updatedAt = time.Now() }
func (t *Tag) Description() string     { return t.description }
func (t *Tag) SetDescription(desc string) { t.description = desc; t.updatedAt = time.Now() }
func (t *Tag) Color() string            { return t.color }
func (t *Tag) SetColor(color string)    { t.color = color; t.updatedAt = time.Now() }
func (t *Tag) CreatedAt() time.Time     { return t.createdAt }
func (t *Tag) UpdatedAt() time.Time     { return t.updatedAt }

// Validate 验证标签数据有效性
func (t *Tag) Validate() error {
	if len(t.name) == 0 {
		return ErrTagNameRequired
	}
	if len(t.name) > 50 {
		return ErrTagTooLong
	}
	if t.color != "" && !isValidColor(t.color) {
		return ErrTagColorInvalid
	}
	return nil
}

// CanBeDeleted 检查标签是否可以删除（业务规则：无关联商品时才可删除）
func (t *Tag) CanBeDelete(associatedProductCount int64) error {
	if associatedProductCount > 0 {
		return fmt.Errorf("该标签已关联 %d 个商品", associatedProductCount)
	}
	return nil
}

// UpdateInfo 更新标签信息
func (t *Tag) UpdateInfo(name, description, color string) error {
	if name != "" {
		if len(name) > 50 {
			return ErrTagTooLong
		}
		t.name = name
	}
	if description != "" {
		t.description = description
	}
	if color != "" {
		if !isValidColor(color) {
			return ErrTagColorInvalid
		}
		t.color = color
	}
	t.updatedAt = time.Now()
	return nil
}

func isValidColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for _, c := range color[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
