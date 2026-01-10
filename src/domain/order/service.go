package order

import (
	"orderease/domain/product"
	"orderease/domain/shared"
)

// ProductFinder 商品查找器接口（依赖反转）
// 领域层通过此接口获取商品信息，而不直接依赖 Repository
type ProductFinder interface {
	FindProduct(id shared.ID) (*product.Product, error)
	FindOption(id shared.ID) (*product.ProductOption, error)
	FindOptionCategory(id shared.ID) (*product.ProductOptionCategory, error)
}

// OrderDomainService 订单领域服务
// 处理跨聚合的业务逻辑
type OrderDomainService struct{}

// NewOrderDomainService 创建订单领域服务
func NewOrderDomainService() *OrderDomainService {
	return &OrderDomainService{}
}

// ValidateOrderCreation 验证订单创建
func (s *OrderDomainService) ValidateOrderCreation(ord *Order, finder ProductFinder) error {
	// 先验证商品
	if err := ord.ValidateItems(finder); err != nil {
		return err
	}

	// 再计算价格
	if err := ord.CalculateTotal(finder); err != nil {
		return err
	}

	return nil
}
