// Package product 提供商品领域模型的核心业务逻辑。
//
// 职责：
//   - 商品生命周期管理（创建、上架、下架、删除）
//   - 库存管理（增减、验证）
//   - 商品参数选项管理
//   - 商品数据安全处理（XSS防护）
//
// 业务规则：
//   - 新创建商品默认状态为 pending（待上架）
//   - 库存不能为负数
//   - 只有 online 状态的商品可被下单
//   - 商品名称和描述需经过 XSS 清理
//
// 使用示例：
//
//	// 创建新商品
//	p := product.NewProduct(shopID, "咖啡", 25.00, 100)
//
//	// 检查库存并扣减
//	if p.HasEnoughStock(5) {
//	    p.DecreaseStock(5)
//	}
//
//	// 数据清理后保存
//	p.Sanitize()
package product

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"orderease/models"
	"orderease/utils"
)

// ProductStatus 商品状态
//
// 状态流转：pending -> online -> offline
//   - pending: 待上架，新建商品默认状态
//   - online:  已上架，可被用户下单
//   - offline: 已下架，不可被下单但保留数据
type ProductStatus string

const (
	ProductStatusPending ProductStatus = "pending" // 待上架
	ProductStatusOnline  ProductStatus = "online"  // 已上架
	ProductStatusOffline ProductStatus = "offline" // 已下架
)

// Product 商品聚合根
//
// 作为聚合根，Product 负责：
//   - 管理商品自身的状态和属性
//   - 维护商品参数选项 (OptionCategories) 集合
//   - 提供库存管理和数据安全处理
//
// 约束：
//   - ID 为 0 表示未持久化的新商品
//   - price 单位为元，支持小数
//   - stock 为整数，不能为负数
//   - 持久化前必须调用 Sanitize() 清理数据
type Product struct {
	id               snowflake.ID
	shopID           snowflake.ID
	name             string
	description      string
	price            float64
	stock            int
	imageURL         string
	status           ProductStatus
	optionCategories []models.ProductOptionCategory
	createdAt        time.Time
	updatedAt        time.Time
}

// NewProduct 创建新商品（简化版）
//
// 参数：
//   - shopID: 所属店铺ID
//   - name:   商品名称
//   - price:  商品单价（元）
//   - stock:  初始库存数量
//
// 返回：
//   - 状态为 pending 的商品实体
//
// 注意：
//   - ID 为 0，需在持久化时分配
//   - 不包含描述、图片、参数选项，需通过 Setters 设置
//   - 如需一次性设置所有字段，使用 NewProductWithDefaults
func NewProduct(shopID snowflake.ID, name string, price float64, stock int) *Product {
	return &Product{
		shopID:    shopID,
		name:      name,
		price:     price,
		stock:     stock,
		status:    ProductStatusPending,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
}

// NewProductWithDefaults 创建带完整默认值的商品（完整版）
//
// 参数：
//   - shopID:           所属店铺ID
//   - name:             商品名称
//   - price:            商品单价（元）
//   - stock:            初始库存数量
//   - description:      商品描述
//   - imageURL:         商品图片URL
//   - optionCategories: 商品参数选项类别
//
// 返回：
//   - 状态为 pending 的完整商品实体
//
// 与 NewProduct 的区别：
//   - 支持一次性设置所有字段
//   - 减少 Handler 层代码复杂度
//   - 避免多次调用 Setter 方法
func NewProductWithDefaults(shopID snowflake.ID, name string, price float64, stock int,
	description string, imageURL string, optionCategories []models.ProductOptionCategory) *Product {

	return &Product{
		shopID:           shopID,
		name:             name,
		price:            price,
		stock:            stock,
		description:      description,
		imageURL:         imageURL,
		status:           ProductStatusPending, // 默认状态
		optionCategories: optionCategories,
		createdAt:        time.Now(),
		updatedAt:        time.Now(),
	}
}

// Getters
func (p *Product) ID() snowflake.ID {
	return p.id
}

func (p *Product) ShopID() snowflake.ID {
	return p.shopID
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Description() string {
	return p.description
}

func (p *Product) Price() float64 {
	return p.price
}

func (p *Product) Stock() int {
	return p.stock
}

func (p *Product) ImageURL() string {
	return p.imageURL
}

func (p *Product) Status() ProductStatus {
	return p.status
}

func (p *Product) OptionCategories() []models.ProductOptionCategory {
	return p.optionCategories
}

func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

// Setters
func (p *Product) SetID(id snowflake.ID) {
	p.id = id
}

func (p *Product) SetName(name string) {
	p.name = name
}

func (p *Product) SetDescription(desc string) {
	p.description = desc
}

func (p *Product) SetPrice(price float64) {
	p.price = price
}

func (p *Product) SetStock(stock int) {
	p.stock = stock
}

func (p *Product) SetImageURL(url string) {
	p.imageURL = url
}

func (p *Product) SetStatus(status ProductStatus) {
	p.status = status
}

func (p *Product) SetOptionCategories(categories []models.ProductOptionCategory) {
	p.optionCategories = categories
}

func (p *Product) SetCreatedAt(t time.Time) {
	p.createdAt = t
}

func (p *Product) SetUpdatedAt(t time.Time) {
	p.updatedAt = t
}

// ==================== 业务方法 ====================

// IsOnline 判断商品是否已上架
//
// 使用场景：
//   - 前端商品列表过滤
//   - 下单时验证商品可售状态
//   - 商品搜索条件
func (p *Product) IsOnline() bool {
	return p.status == ProductStatusOnline
}

// IsOffline 判断商品是否已下架
func (p *Product) IsOffline() bool {
	return p.status == ProductStatusOffline
}

// IsPending 判断商品是否待上架
func (p *Product) IsPending() bool {
	return p.status == ProductStatusPending
}

// InStock 判断是否有库存
func (p *Product) InStock() bool {
	return p.stock > 0
}

// HasEnoughStock 判断库存是否足够
//
// 参数：
//   - quantity: 需要的数量
//
// 返回：
//   - true:  库存充足（stock >= quantity）
//   - false: 库存不足
//
// 使用时机：
//   - 下单前验证库存
//   - 购物车数量调整时
//   - 库存预警判断
func (p *Product) HasEnoughStock(quantity int) bool {
	return p.stock >= quantity
}

// DecreaseStock 减少库存（下单时调用）
//
// 参数：
//   - quantity: 减少的数量
//
// 安全机制：
//   - 如果库存不足，不执行扣减（静默处理）
//   - 业务层应先调用 HasEnoughStock 验证
//
// 相关方法：
//   - HasEnoughStock: 扣减前验证
//   - IncreaseStock:  恢复库存（取消订单时）
func (p *Product) DecreaseStock(quantity int) {
	if p.stock >= quantity {
		p.stock -= quantity
	}
}

// IncreaseStock 增加库存（取消订单时调用）
//
// 参数：
//   - quantity: 增加的数量
//
// 使用场景：
//   - 订单取消时恢复库存
//   - 退货入库时
//   - 库存补货时
//
// 注意：此方法不检查上限，允许无限增加
func (p *Product) IncreaseStock(quantity int) {
	p.stock += quantity
}

// Sanitize 清理商品数据，防止 XSS 攻击
//
// 清理内容：
//   - name:        HTML 标签转义
//   - description: HTML 标签转义
//
// 使用时机：
//   - 持久化前必须调用
//   - 从用户输入更新字段后
//
// 注意：
//   - 图片 URL 验证由 Media Service 处理
//   - 价格、库存等数值类型无需清理
func (p *Product) Sanitize() {
	p.name = utils.SanitizeString(p.name)
	p.description = utils.SanitizeString(p.description)
	// 图片验证已在 Media Service 处理
}

// ToModel 转换为持久化模型
func (p *Product) ToModel() *models.Product {
	return &models.Product{
		ID:               p.id,
		ShopID:           p.shopID,
		Name:             p.name,
		Description:      p.description,
		Price:            p.price,
		Stock:            p.stock,
		ImageURL:         p.imageURL,
		Status:           string(p.status),
		OptionCategories: p.optionCategories,
		CreatedAt:        p.createdAt,
		UpdatedAt:        p.updatedAt,
	}
}

// ProductFromModel 从持久化模型创建领域实体
func ProductFromModel(model *models.Product) *Product {
	return &Product{
		id:               model.ID,
		shopID:           model.ShopID,
		name:             model.Name,
		description:      model.Description,
		price:            model.Price,
		stock:            model.Stock,
		imageURL:         model.ImageURL,
		status:           ProductStatus(model.Status),
		optionCategories: model.OptionCategories,
		createdAt:        model.CreatedAt,
		updatedAt:        model.UpdatedAt,
	}
}
