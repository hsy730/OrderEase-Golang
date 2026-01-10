# 电商系统DDD战略设计方案

## 文档审查摘要

**审查日期**：2026-01-10
**审查结果**：第一阶段单体服务重构方案存在多处与实际情况不符，需要调整优化

### 主要发现
1. **目录结构重组声明不实**：文档声称代码已按上下文重组，但实际代码库仍保持原有的DDD四层架构，未进行大规模目录重组
2. **共享内核常量未统一**：文档描述常量已在shared包定义，实际常量分散在models目录中
3. **防腐层实现时机过早**：在单体架构中立即实现完整防腐层属于过度设计，应移至第二阶段
4. **技术准备顺序不合理**：应先引入依赖注入框架，再解决依赖类型错误，然后建立统一错误处理
5. **数据库前缀策略缺乏必要性评估**：当前表名简洁清晰，通过ShopID等字段已实现逻辑隔离，前缀策略可能增加复杂度
6. **服务依赖类型错误**：OrderService和ProductService中存在服务依赖类型错误
7. **领域模型贫血**：业务逻辑集中在Application Service而非Domain Entity

### 修复建议
- 修正文档与实际情况不符的部分
- 优先修复服务依赖类型错误
- 引入Wire依赖注入框架
- 统一常量到共享内核
- 增强领域模型，迁移业务逻辑到实体
- 将防腐层实现移至第二阶段（服务拆分准备）
- 制定具体测试策略和覆盖率目标

## 目录
- [1. 上下文映射图](#1-上下文映射图)
- [2. 限界上下文识别](#2-限界上下文识别)
- [3. 共享内核](#3-共享内核)
- [4. 上下文映射关系](#4-上下文映射关系)
- [5. 防腐层设计](#5-防腐层设计)
- [6. 微服务拆分路线图](#6-微服务拆分路线图)
- [7. 服务间通信方式](#7-服务间通信方式)
- [8. 数据一致性策略](#8-数据一致性策略)
- [9. 实施计划总结](#9-实施计划总结)
- [10. 重构审查总结](#10-重构审查总结)

---

## 1. 上下文映射图

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         共享内核 (Shared Kernel)                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  ID (Snowflake) │  │   Price      │  │  常量定义    │  │  通用工具     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
        ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
        │  用户上下文      │ │  店铺上下文      │ │  商品上下文      │
        │  User Context    │ │  Shop Context    │ │ Product Context  │
        │                  │ │                  │ │                  │
        │  聚合根:        │ │  聚合根:        │ │  聚合根:        │
        │  - User         │ │  - Shop         │ │  - Product       │
        │                  │ │  - Tag          │ │  - OptionCategory │
        │  实体:          │ │                  │ │  - Option        │
        │  - User         │ │  实体:          │ │                  │
        │                  │ │  - Shop         │ │  实体:          │
        │  值对象:        │ │  - Tag          │ │  - Product       │
        │  - UserRole     │ │                  │ │  - OptionCategory │
        │  - UserType     │ │  值对象:        │ │  - Option        │
        │                  │ │  - OrderStatusFlow│ │                  │
        │  领域服务:      │ │                  │ │  值对象:        │
        │  - 用户认证      │ │  领域服务:      │ │  - ProductStatus │
        │                  │ │  - 店铺管理      │ │                  │
        │  仓储:          │ │  - 订单流转配置  │ │  领域服务:      │
        │  - UserRepository│ │                  │ │  - 商品管理      │
        │                  │ │  仓储:          │ │  - 库存管理      │
        │  应用服务:      │ │  - ShopRepository│ │                  │
        │  - UserService   │ │  - TagRepository │ │  仓储:          │
        │                  │ │                  │ │  - ProductRepository│
        │  接口:          │ │  应用服务:      │ │  - OptionCategoryRepository│
        │  - UserAPI      │ │  - ShopService  │ │  - OptionRepository│
        │                  │ │  - TagService   │ │  - ProductTagRepository│
        └──────────────────┘ └──────────────────┘ │                  │
                    │               │               │  应用服务:      │
                    │               │               │  - ProductService│
                    │               │               │                  │
                    │               │               │  接口:          │
                    │               │               │  - ProductAPI    │
                    │               │               └──────────────────┘
                    │               │                       │
                    │               │                       │
                    │               └───────────┬───────────┘
                    │                           │
                    │                           ▼
                    │               ┌──────────────────┐
                    │               │  订单上下文      │
                    │               │  Order Context   │
                    │               │                  │
                    │               │  聚合根:        │
                    │               │  - Order         │
                    │               │                  │
                    │               │  实体:          │
                    │               │  - Order         │
                    │               │  - OrderItem     │
                    │               │  - OrderItemOption│
                    │               │  - OrderStatusLog │
                    │               │                  │
                    │               │  值对象:        │
                    │               │  - OrderStatus   │
                    │               │  - OrderStatusFlow│
                    │               │                  │
                    │               │  领域服务:      │
                    │               │  - 订单创建      │
                    │               │  - 状态流转      │
                    │               │                  │
                    │               │  仓储:          │
                    │               │  - OrderRepository│
                    │               │  - OrderItemRepository│
                    │               │  - OrderItemOptionRepository│
                    │               │  - OrderStatusLogRepository│
                    │               │                  │
                    │               │  应用服务:      │
                    │               │  - OrderService  │
                    │               │                  │
                    │               │  接口:          │
                    │               │  - OrderAPI      │
                    │               └──────────────────┘
                    │
                    ▼
        ┌──────────────────────────────────────────────────────────────┐
        │                    防腐层 (ACL)                             │
        │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐│
        │  │ ProductAdapter │  │  ShopAdapter  │  │  UserAdapter  ││
        │  │ (商品适配器)   │  │ (店铺适配器)   │  │ (用户适配器)   ││
        │  └────────────────┘  └────────────────┘  └────────────────┘│
        └──────────────────────────────────────────────────────────────┘
```

---

## 2. 限界上下文识别

基于代码分析，识别出以下四个限界上下文：

### 2.1 用户上下文

**职责范围**：用户身份管理、认证授权、用户信息维护

**核心聚合**：
- `User` - 用户聚合根

**实体**：
- `User` - 用户实体

**值对象**：
- `UserRole` - 用户角色（private_user, public_user）
- `UserType` - 用户类型（delivery, pickup, system）

**领域服务**：
- 用户认证服务
- 用户注册服务

**仓储接口**：
```go
type UserRepository interface {
    Save(user *User) error
    FindByID(id shared.ID) (*User, error)
    FindByName(name string) (*User, error)
    FindAll(page, pageSize int) ([]User, int64, error)
    Delete(id shared.ID) error
    Update(user *User) error
    Exists(id shared.ID) (bool, error)
}
```

**应用服务**：
- `UserService` - 用户应用服务

**接口**：
- `UserAPI` - 用户管理API

**独立性评估**：⭐⭐⭐⭐⭐（高度独立）
- 业务逻辑自包含
- 依赖关系简单
- 变更频率低
- 可独立部署和扩展

---

### 2.2 店铺上下文

**职责范围**：店铺管理、标签管理、订单流转配置

**核心聚合**：
- `Shop` - 店铺聚合根
- `Tag` - 标签聚合根

**实体**：
- `Shop` - 店铺实体
- `Tag` - 标签实体

**值对象**：
- `OrderStatusFlow` - 订单流转状态配置

**领域服务**：
- 店铺管理服务
- 标签管理服务
- 订单流转配置服务

**仓储接口**：
```go
type ShopRepository interface {
    Save(shop *Shop) error
    FindByID(id uint64) (*Shop, error)
    FindByName(name string) (*Shop, error)
    FindByOwnerUsername(username string) (*Shop, error)
    FindAll(page, pageSize int, search string) ([]Shop, int64, error)
    Delete(id uint64) error
    Update(shop *Shop) error
    Exists(id uint64) (bool, error)
}

type TagRepository interface {
    Save(tag *Tag) error
    FindByID(id int) (*Tag, error)
    FindByShopID(shopID uint64) ([]Tag, error)
    Delete(id int) error
    Update(tag *Tag) error
}
```

**应用服务**：
- `ShopService` - 店铺应用服务
- `TagService` - 标签应用服务

**接口**：
- `ShopAPI` - 店铺管理API
- `TagAPI` - 标签管理API

**独立性评估**：⭐⭐⭐⭐（相对独立）
- 业务边界清晰
- 包含标签管理
- 订单流转配置独立
- 与商品上下文有依赖关系

---

### 2.3 商品上下文

**职责范围**：商品管理、库存管理、商品选项配置

**核心聚合**：
- `Product` - 商品聚合根
- `ProductOptionCategory` - 商品选项类别聚合根
- `ProductOption` - 商品选项聚合根

**实体**：
- `Product` - 商品实体
- `ProductOptionCategory` - 商品选项类别实体
- `ProductOption` - 商品选项实体

**值对象**：
- `ProductStatus` - 商品状态（pending, online, offline）

**领域服务**：
- 商品管理服务
- 库存管理服务
- 商品选项配置服务

**仓储接口**：
```go
type ProductRepository interface {
    Save(product *Product) error
    FindByID(id shared.ID) (*Product, error)
    FindByIDAndShopID(id shared.ID, shopID uint64) (*Product, error)
    FindByShopID(shopID uint64, page, pageSize int, search string, excludeOffline bool) ([]Product, int64, error)
    FindByIDs(ids []shared.ID) ([]Product, error)
    Delete(id shared.ID) error
    Update(product *Product) error
    CountByProductID(productID shared.ID) (int64, error)
}

type ProductOptionCategoryRepository interface {
    Save(category *ProductOptionCategory) error
    FindByID(id shared.ID) (*ProductOptionCategory, error)
    FindByProductID(productID shared.ID) ([]ProductOptionCategory, error)
    DeleteByProductID(productID shared.ID) error
}

type ProductOptionRepository interface {
    Save(option *ProductOption) error
    FindByID(id shared.ID) (*ProductOption, error)
    FindByCategoryID(categoryID shared.ID) ([]ProductOption, error)
    DeleteByCategoryID(categoryID shared.ID) error
}

type ProductTagRepository interface {
    Save(productID shared.ID, tagID int) error
    FindByProductID(productID shared.ID) ([]int, error)
    FindByTagID(tagID int) ([]shared.ID, error)
    DeleteByProductID(productID shared.ID) error
}
```

**应用服务**：
- `ProductService` - 商品应用服务

**接口**：
- `ProductAPI` - 商品管理API

**独立性评估**：⭐⭐⭐（依赖店铺）
- 业务边界清晰
- 依赖店铺上下文（ShopID）
- 被订单上下文依赖
- 库存管理复杂度高

---

### 2.4 订单上下文

**职责范围**：订单生命周期管理、订单状态流转、订单历史记录

**核心聚合**：
- `Order` - 订单聚合根

**实体**：
- `Order` - 订单实体
- `OrderItem` - 订单项实体
- `OrderItemOption` - 订单项选项实体
- `OrderStatusLog` - 订单状态日志实体

**值对象**：
- `OrderStatus` - 订单状态
- `OrderStatusFlow` - 订单流转状态配置
- `OrderStatusConfig` - 订单状态配置
- `OrderStatusTransition` - 订单状态转换

**领域服务**：
- 订单创建服务
- 订单状态流转服务
- 订单历史记录服务

**仓储接口**：
```go
type OrderRepository interface {
    Save(order *Order) error
    FindByID(id shared.ID) (*Order, error)
    FindByIDAndShopID(id shared.ID, shopID uint64) (*Order, error)
    FindByShopID(shopID uint64, page, pageSize int) ([]Order, int64, error)
    FindByUserID(userID shared.ID, shopID uint64, page, pageSize int) ([]Order, int64, error)
    FindUnfinishedByShopID(shopID uint64, flow OrderStatusFlow, page, pageSize int) ([]Order, int64, error)
    Search(shopID uint64, userID string, statuses []OrderStatus, startTime, endTime time.Time, page, pageSize int) ([]Order, int64, error)
    Delete(id shared.ID) error
    Update(order *Order) error
}

type OrderItemRepository interface {
    Save(item *OrderItem) error
    FindByOrderID(orderID shared.ID) ([]OrderItem, error)
    DeleteByOrderID(orderID shared.ID) error
}

type OrderItemOptionRepository interface {
    Save(option *OrderItemOption) error
    FindByOrderItemID(orderItemID shared.ID) ([]OrderItemOption, error)
    DeleteByOrderItemID(orderItemID shared.ID) error
}

type OrderStatusLogRepository interface {
    Save(log *OrderStatusLog) error
    FindByOrderID(orderID shared.ID) ([]OrderStatusLog, error)
    DeleteByOrderID(orderID shared.ID) error
}
```

**应用服务**：
- `OrderService` - 订单应用服务

**接口**：
- `OrderAPI` - 订单管理API

**独立性评估**：⭐⭐（强依赖其他上下文）
- 依赖用户上下文（UserID）
- 依赖店铺上下文（ShopID）
- 依赖商品上下文（ProductID）
- 业务复杂度高
- 状态流转逻辑复杂

---

## 3. 共享内核

共享内核是多个上下文之间共享的领域模型部分，需要保持一致性。

> **📌 当前状态说明**
> 
> **已实现部分**：
> - ID类型：已在`domain/shared/id.go`中定义
> - Price类型：已在`domain/shared/price.go`中定义
> 
> **待统一部分**：
> - 订单状态常量：当前分散在`models/order.go`中，需迁移到`domain/shared/constants.go`
> - 商品状态常量：当前分散在`models/product.go`中，需迁移到`domain/shared/constants.go`
> - 通用工具：当前部分工具分散在utils包，需评估是否纳入共享内核
> 
> **第一阶段目标**：将分散的常量统一到`domain/shared/constants.go`，形成真正的共享内核。

### 3.1 基础类型

#### ID类型
```go
package shared

import (
    "github.com/bwmarrin/snowflake"
)

type ID snowflake.ID

func NewID() ID {
    return ID(snowflake.ID(0))
}

func (id ID) Value() snowflake.ID {
    return snowflake.ID(id)
}

func (id ID) String() string {
    return snowflake.ID(id).String()
}

func (id ID) IsZero() bool {
    return id == 0
}

func ParseIDFromString(s string) (ID, error) {
    id, err := snowflake.ParseString(s)
    return ID(id), err
}

func ParseIDFromUint64(u uint64) ID {
    return ID(u)
}

func (id ID) ToUint64() uint64 {
    return uint64(id)
}
```

#### Price类型
```go
package shared

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "strconv"
)

type Price float64

func (p Price) String() string {
    return fmt.Sprintf("%.2f", p)
}

func (p *Price) Scan(value interface{}) error {
    switch v := value.(type) {
    case float64:
        *p = Price(v)
    case int64:
        *p = Price(float64(v))
    case []uint8:
        if f, err := strconv.ParseFloat(string(v), 64); err == nil {
            *p = Price(f)
        } else {
            return fmt.Errorf("failed to parse Price from string: %v", err)
        }
    default:
        return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *Price", value)
    }
    return nil
}

func (p Price) Value() (driver.Value, error) {
    return float64(p), nil
}

func (p *Price) UnmarshalJSON(data []byte) error {
    var value interface{}
    if err := json.Unmarshal(data, &value); err != nil {
        return err
    }

    switch v := value.(type) {
    case float64:
        *p = Price(v)
    case float32:
        *p = Price(float64(v))
    case string:
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            *p = Price(f)
        } else {
            return fmt.Errorf("invalid price format: %s", v)
        }
    case int:
        *p = Price(float64(v))
    case int64:
        *p = Price(float64(v))
    case int32:
        *p = Price(float64(v))
    default:
        return fmt.Errorf("invalid price type: %T", value)
    }
    return nil
}

func (p Price) ToFloat64() float64 {
    return float64(p)
}

func NewPrice(value float64) Price {
    return Price(value)
}

func (p Price) Add(other Price) Price {
    return p + other
}

func (p Price) Multiply(quantity int) Price {
    return p * Price(quantity)
}

func (p Price) IsZero() bool {
    return p == 0
}

func (p Price) IsPositive() bool {
    return p > 0
}
```

### 3.2 常量定义

#### 订单状态常量
```go
const (
    OrderStatusPending  OrderStatus = 1  // 待处理
    OrderStatusAccepted OrderStatus = 2  // 已接单
    OrderStatusRejected OrderStatus = 3  // 已拒绝
    OrderStatusShipped  OrderStatus = 4  // 已发货
    OrderStatusComplete OrderStatus = 10 // 已完成
    OrderStatusCanceled OrderStatus = -1 // 已取消
)
```

#### 商品状态常量
```go
const (
    ProductStatusPending ProductStatus = "pending" // 待上架
    ProductStatusOnline  ProductStatus = "online"  // 已上架
    ProductStatusOffline ProductStatus = "offline" // 已下架
)
```

### 3.3 通用工具

#### ID生成器
```go
package utils

import (
    "github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func InitSnowflake(nodeID int64) error {
    var err error
    node, err = snowflake.NewNode(nodeID)
    return err
}

func GenerateSnowflakeID() int64 {
    return node.Generate().Int64()
}
```

### 3.4 共享理由

1. **全局唯一性**：ID需要在所有上下文中保持唯一
2. **类型一致性**：Price的计算逻辑需要在所有上下文中保持一致
3. **状态一致性**：订单状态和商品状态需要在所有上下文中保持一致
4. **避免重复**：避免在每个上下文中重复实现相同的功能
5. **降低维护成本**：共享内核的修改只需要在一处进行

---

## 4. 上下文映射关系

### 4.1 映射关系图

```
┌─────────────┐
│ 用户上下文   │ ◄─────────────────────┐
└─────────────┘                       │
        │                             │
        │ Customer/Supplier            │
        │ (用户是订单的客户)            │
        ▼                             │
┌─────────────┐                       │
│ 订单上下文   │ ◄─────────────────────┤
└─────────────┘                       │
        │                             │
        │ OHS (Open Host Service)     │
        │ (商品服务对外开放)            │
        ▲                             │
┌─────────────┐                       │
│ 商品上下文   │ ◄─────────────────────┤
└─────────────┘                       │
        │                             │
        │ ACL (Anti-Corruption Layer)  │
        │ (防腐层保护订单上下文)        │
        ▼                             │
┌─────────────┐                       │
│ 店铺上下文   │ ◄─────────────────────┘
└─────────────┘
```

### 4.2 映射关系说明

| 关系类型 | 源上下文 | 目标上下文 | 说明 | 方向 |
|---------|---------|-----------|------|------|
| **Customer/Supplier** | 用户上下文 | 订单上下文 | 订单依赖用户信息，用户是订单的客户 | 用户 → 订单 |
| **OHS** | 商品上下文 | 订单上下文 | 商品上下文作为开放主机服务，订单上下文可以查询商品信息 | 商品 → 订单 |
| **ACL** | 订单上下文 | 商品上下文 | 订单上下文通过防腐层调用商品服务，避免直接依赖商品内部模型 | 订单 → 商品 |
| **Partnership** | 店铺上下文 | 商品上下文 | 店铺和商品是伙伴关系，商品属于店铺 | 店铺 ↔ 商品 |
| **Partnership** | 店铺上下文 | 订单上下文 | 店铺和订单是伙伴关系，订单属于店铺 | 店铺 ↔ 订单 |

### 4.3 关系类型详解

#### Customer/Supplier（客户/供应商关系）
- **定义**：一个上下文向另一个上下文提供服务
- **场景**：用户上下文为订单上下文提供用户信息
- **实现**：订单上下文通过用户ID查询用户信息
- **特点**：单向依赖，用户上下文不依赖订单上下文

#### OHS（Open Host Service，开放主机服务）
- **定义**：一个上下文提供标准化的服务接口供其他上下文使用
- **场景**：商品上下文提供商品查询接口，订单上下文查询商品信息
- **实现**：定义REST API或gRPC接口
- **特点**：接口稳定，版本化管理

#### ACL（Anti-Corruption Layer，防腐层）
- **定义**：隔离外部上下文的影响，保护内部上下文
- **场景**：订单上下文调用商品服务时使用防腐层
- **实现**：DTO转换、数据映射、错误处理
- **特点**：解耦依赖，隔离变化

#### Partnership（伙伴关系）
- **定义**：两个上下文紧密合作，相互依赖
- **场景**：店铺和商品、店铺和订单
- **实现**：共享数据模型或通过接口调用
- **特点**：双向依赖，需要协调变更

---

## 5. 防腐层设计

> **📌 阶段调整说明**
> 
> 根据架构审查结果，原计划在第一阶段实现的完整防腐层（ACL）已调整至**第二阶段（服务拆分准备）**实施。
> 
> **调整理由**：
> 1. **避免过度设计**：在单体架构中立即实现完整防腐层会增加不必要的抽象层和代码复杂度
> 2. **开发效率考量**：所有跨上下文调用都需要额外的DTO转换，降低开发效率
> 3. **实施优先级**：第一阶段应优先解决架构硬伤（服务依赖类型错误）和基础设施准备（Wire DI框架）
> 
> **第一阶段替代方案**：
> - 明确定义上下文接口契约（接口文档或代码契约）
> - 减少隐式依赖，强化上下文边界
> - 为第二阶段实现防腐层做技术准备
> 
> **第二阶段计划**：
> - 完整实现防腐层架构（DTO、Mapper、Client）
> - 实现服务间通信的容错机制（重试、熔断、降级）
> - 建立统一的错误处理和数据转换标准

### 5.1 防腐层架构图

```
订单上下文中的防腐层结构：

┌─────────────────────────────────────────────────┐
│            订单应用服务层                      │
│         (Order Application Service)            │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│              防腐层 (ACL)                    │
│  ┌──────────────┐  ┌──────────────┐         │
│  │ ProductDTO   │  │  ShopDTO     │         │
│  │ (商品数据传输) │  │ (店铺数据传输) │         │
│  └──────────────┘  └──────────────┘         │
│  ┌──────────────┐  ┌──────────────┐         │
│  │ ProductMapper│  │  ShopMapper  │         │
│  │ (商品映射器)   │  │ (店铺映射器)   │         │
│  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│           外部服务调用层                      │
│  ┌──────────────┐  ┌──────────────┐         │
│  │ ProductClient│  │  ShopClient  │         │
│  │ (商品服务客户端) │  │ (店铺服务客户端) │         │
│  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────┘
```

### 5.2 防腐层职责

1. **数据转换**：将外部服务的DTO转换为订单上下文的领域模型
2. **隔离依赖**：隔离外部服务的变化，保护订单上下文
3. **缓存优化**：缓存外部数据，减少远程调用
4. **错误处理**：统一处理外部服务的异常
5. **重试机制**：实现调用失败的重试逻辑
6. **熔断降级**：实现服务熔断和降级策略

### 5.3 防腐层实现示例

#### ProductDTO（商品数据传输对象）
```go
package dto

type ProductDTO struct {
    ID               shared.ID                    `json:"id"`
    ShopID           uint64                      `json:"shop_id"`
    Name             string                      `json:"name"`
    Description      string                      `json:"description"`
    Price            shared.Price                 `json:"price"`
    Stock            int                         `json:"stock"`
    ImageURL         string                      `json:"image_url"`
    Status           string                      `json:"status"`
    OptionCategories []ProductOptionCategoryDTO  `json:"option_categories"`
}

type ProductOptionCategoryDTO struct {
    ID           shared.ID                `json:"id"`
    ProductID    shared.ID               `json:"product_id"`
    Name         string                  `json:"name"`
    IsRequired   bool                    `json:"is_required"`
    IsMultiple   bool                    `json:"is_multiple"`
    DisplayOrder int                     `json:"display_order"`
    Options      []ProductOptionDTO      `json:"options"`
}

type ProductOptionDTO struct {
    ID              shared.ID `json:"id"`
    CategoryID      shared.ID `json:"category_id"`
    Name            string    `json:"name"`
    PriceAdjustment float64   `json:"price_adjustment"`
    DisplayOrder    int       `json:"display_order"`
    IsDefault       bool      `json:"is_default"`
}
```

#### ProductMapper（商品映射器）
```go
package acl

import (
    "orderease/domain/order"
    "orderease/domain/product"
    "orderease/application/dto"
)

type ProductMapper struct{}

func NewProductMapper() *ProductMapper {
    return &ProductMapper{}
}

func (m *ProductMapper) DTOToDomain(dto *dto.ProductDTO) *product.Product {
    optionCategories := make([]product.ProductOptionCategory, len(dto.OptionCategories))
    for i, catDTO := range dto.OptionCategories {
        options := make([]product.ProductOption, len(catDTO.Options))
        for j, optDTO := range catDTO.Options {
            options[j] = product.ProductOption{
                ID:              optDTO.ID,
                CategoryID:      optDTO.CategoryID,
                Name:            optDTO.Name,
                PriceAdjustment: optDTO.PriceAdjustment,
                DisplayOrder:    optDTO.DisplayOrder,
                IsDefault:       optDTO.IsDefault,
            }
        }

        optionCategories[i] = product.ProductOptionCategory{
            ID:           catDTO.ID,
            ProductID:    catDTO.ProductID,
            Name:         catDTO.Name,
            IsRequired:   catDTO.IsRequired,
            IsMultiple:   catDTO.IsMultiple,
            DisplayOrder: catDTO.DisplayOrder,
            Options:      options,
        }
    }

    return &product.Product{
        ID:               dto.ID,
        ShopID:           dto.ShopID,
        Name:             dto.Name,
        Description:      dto.Description,
        Price:            dto.Price,
        Stock:            dto.Stock,
        ImageURL:         dto.ImageURL,
        Status:           product.ProductStatus(dto.Status),
        OptionCategories: optionCategories,
    }
}

func (m *ProductMapper) DomainToOrderItem(prod *product.Product, quantity int) order.OrderItem {
    return order.OrderItem{
        ProductID:         prod.ID,
        Quantity:          quantity,
        Price:             prod.Price,
        ProductName:       prod.Name,
        ProductDescription: prod.Description,
        ProductImageURL:   prod.ImageURL,
    }
}
```

#### ProductClient（商品服务客户端）
```go
package acl

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "orderease/application/dto"
    "orderease/domain/shared"
)

type ProductClient interface {
    GetProduct(ctx context.Context, productID shared.ID) (*dto.ProductDTO, error)
    GetProducts(ctx context.Context, shopID uint64, page, pageSize int) (*dto.ProductListDTO, error)
    DecreaseStock(ctx context.Context, productID shared.ID, quantity int) error
    IncreaseStock(ctx context.Context, productID shared.ID, quantity int) error
}

type HTTPProductClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

func NewHTTPProductClient(baseURL string) *HTTPProductClient {
    return &HTTPProductClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        timeout: 30 * time.Second,
    }
}

func (c *HTTPProductClient) GetProduct(ctx context.Context, productID shared.ID) (*dto.ProductDTO, error) {
    url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, productID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
    }

    var dto dto.ProductDTO
    if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    return &dto, nil
}

func (c *HTTPProductClient) DecreaseStock(ctx context.Context, productID shared.ID, quantity int) error {
    url := fmt.Sprintf("%s/api/v1/products/%s/stock", c.baseURL, productID)

    reqBody := map[string]interface{}{
        "quantity": quantity,
        "action":  "decrease",
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return fmt.Errorf("序列化请求体失败: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("创建请求失败: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
    }

    return nil
}
```

#### ShopClient（店铺服务客户端）
```go
package acl

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "orderease/application/dto"
)

type ShopClient interface {
    GetShop(ctx context.Context, shopID uint64) (*dto.ShopDTO, error)
    GetShopTags(ctx context.Context, shopID uint64) ([]dto.TagDTO, error)
}

type HTTPShopClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

func NewHTTPShopClient(baseURL string) *HTTPShopClient {
    return &HTTPShopClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        timeout: 30 * time.Second,
    }
}

func (c *HTTPShopClient) GetShop(ctx context.Context, shopID uint64) (*dto.ShopDTO, error) {
    url := fmt.Sprintf("%s/api/v1/shops/%d", c.baseURL, shopID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
    }

    var dto dto.ShopDTO
    if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    return &dto, nil
}

func (c *HTTPShopClient) GetShopTags(ctx context.Context, shopID uint64) ([]dto.TagDTO, error) {
    url := fmt.Sprintf("%s/api/v1/shops/%d/tags", c.baseURL, shopID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
    }

    var tags []dto.TagDTO
    if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    return tags, nil
}
```

---

## 6. 微服务拆分路线图

### 6.1 第一阶段：单体架构优化（当前阶段）

**目标**：强化现有DDD四层架构，建立清晰的上下文边界，为未来服务拆分奠定基础

**时间规划**：1-2个月

#### 架构现状分析

当前代码库具备良好的DDD四层架构基础，已体现上下文边界，无需大规模目录重组：

```
OrderEase-Golang/src/
├── domain/                    # 领域层（已体现上下文边界）
│   ├── user/                 # 用户上下文
│   ├── shop/                 # 店铺上下文
│   ├── product/              # 商品上下文
│   ├── order/                # 订单上下文
│   └── shared/               # 共享内核（ID、Price类型已提取）
├── application/              # 应用层
│   ├── dto/
│   ├── services/
│   └── interfaces/
├── infrastructure/           # 基础设施层
│   ├── persistence/
│   ├── logging/
│   └── config/
└── interfaces/               # 接口层
    ├── http/
    ├── middleware/
    └── handlers/
```

**架构优势**：
- ✅ DDD四层架构清晰，已按上下文划分domain包
- ✅ 通过domain/user、domain/shop、domain/product、domain/order体现上下文边界
- ✅ 无需创建user-context/、shop-context/等额外目录层级
- ✅ 现有结构简洁，易于维护和迁移

**核心问题识别**：
1. **服务依赖类型错误**（架构硬伤）
   - OrderService中的userRepo参数类型错误
   - ProductService中的orderItemRepo参数类型错误
   - 需立即修复，影响代码可维护性
2. **领域模型贫血**（设计问题）
   - 业务逻辑集中在Application Service中
   - Domain Entity未充分体现业务规则
   - 需逐步迁移业务逻辑到领域层
3. **常量分散在models目录，未统一到共享内核**
   - 订单状态、商品状态等常量分散在models/order.go、models/product.go
   - 需迁移到domain/shared/constants.go形成真正的共享内核
4. **跨上下文直接依赖，缺乏接口契约**
   - 上下文间直接调用，缺乏明确的接口定义
   - 需明确定义接口契约，为第二阶段防腐层做准备

#### 关键任务（调整后优先级）

- [ ] **修复服务依赖类型错误**（最高优先级）
  - OrderService中的userRepo参数类型应为user.UserRepository
  - ProductService中的orderItemRepo参数类型应为order.OrderItemRepository
- [ ] **引入Wire依赖注入框架**（基础设施准备）
  - 建立统一的依赖注入机制
  - 解耦组件间的直接依赖
- [ ] **统一常量到共享内核**
  - 将订单状态、商品状态等常量迁移到domain/shared/constants.go
  - 形成真正的共享内核，确保上下文间一致性
- [ ] **增强领域模型，迁移业务逻辑到实体**
  - 将业务逻辑从Application Service迁移到Domain Entity
  - 实现领域服务处理跨聚合业务逻辑
  - 强化聚合根的业务规则约束
- [ ] **建立统一错误处理机制**
  - 定义领域错误类型和API错误响应
  - 实现跨上下文调用的容错机制
- [ ] **明确定义上下文接口契约**
  - 通过接口文档或代码契约定义跨上下文调用规范
  - 为第二阶段实现防腐层做准备
- [ ] **制定具体测试策略**（单元测试覆盖率≥80%）
  - 核心领域逻辑重点测试
  - 跨上下文交互集成测试
  - 使用接口mock隔离上下文依赖

#### 数据库策略调整

> **📌 策略调整说明**
> 
> 原方案建议使用user_、shop_等表前缀区分上下文，经审查评估后决定**暂不实施**。
> 
> **调整理由**：
> 1. **现有表名简洁清晰**：users、shops、products、orders等表名直观易懂
> 2. **逻辑隔离已充分**：通过ShopID、UserID等字段已实现上下文间的逻辑隔离
> 3. **避免增加复杂度**：表前缀会增加SQL编写和调试复杂度，可能破坏现有查询和关联关系
> 4. **降低迁移成本**：避免不必要的数据迁移和代码修改
> 
> **当前策略**：
> - 保持现有表名结构不变
> - 通过ShopID、UserID等字段实现上下文逻辑隔离
> - 在代码层面通过domain包划分体现上下文边界
> 
> **未来考虑**：
> - 仅当有强烈的多租户物理隔离需求时，再评估表前缀或分库分表方案
> - 在服务拆分阶段（第二阶段后）根据实际需求重新评估数据库隔离策略

#### 技术准备（调整后顺序）

- [ ] **引入Wire依赖注入框架**（第1个月）
  - 提供基础设施支撑
  - 解耦组件依赖
- [ ] **修复服务依赖类型错误**（第1个月）
  - 立即解决代码中的硬伤
- [ ] **建立统一错误处理机制**（第1-2个月）
  - 定义领域错误类型
  - 实现API错误响应
- [ ] **实现请求追踪（Trace ID）**（第2个月）
  - 请求链路追踪
  - 日志关联分析
- [ ] **领域事件机制**（可选，根据团队能力决定）
  - 创建domain/events包
  - 实现事件发布/订阅基础设施

---

### 6.2 第二阶段：服务拆分准备（3-6个月）

**目标**：为微服务拆分做好技术准备

**时间规划**：3-4个月

#### 拆分优先级

##### 第一批：用户服务（User Service）

**拆分理由**：
- ✅ 业务独立性强
- ✅ 依赖关系简单
- ✅ 变更频率低
- ✅ 可以独立扩展

**拆分范围**：
- 用户管理
- 用户认证
- 用户权限

**服务接口**：
```
POST   /api/v1/users/register
POST   /api/v1/users/login
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}
DELETE /api/v1/users/{id}
GET    /api/v1/users
```

**数据迁移**：
- 迁移 `user_users` 表
- 迁移 `user_tokens` 表
- 迁移相关索引和约束

**服务配置**：
```yaml
server:
  port: 8001
  name: user-service

database:
  host: localhost
  port: 3306
  name: user_db
  user: root
  password: password

redis:
  host: localhost
  port: 6379
  db: 0

jwt:
  secret: your-secret-key
  expire: 24h
```

---

##### 第二批：店铺服务（Shop Service）

**拆分理由**：
- ✅ 相对独立
- ✅ 包含标签管理
- ✅ 订单流转配置

**拆分范围**：
- 店铺管理
- 标签管理
- 订单流转配置

**服务接口**：
```
POST   /api/v1/shops
GET    /api/v1/shops/{id}
PUT    /api/v1/shops/{id}
DELETE /api/v1/shops/{id}
GET    /api/v1/shops
POST   /api/v1/shops/{id}/tags
GET    /api/v1/shops/{id}/tags
PUT    /api/v1/shops/{id}/tags/{tagId}
DELETE /api/v1/shops/{id}/tags/{tagId}
```

**数据迁移**：
- 迁移 `shop_shops` 表
- 迁移 `shop_tags` 表
- 迁移 `product_tags` 表（标签关联）

**服务配置**：
```yaml
server:
  port: 8002
  name: shop-service

database:
  host: localhost
  port: 3306
  name: shop_db
  user: root
  password: password
```

---

##### 第三批：商品服务（Product Service）

**拆分理由**：
- ⚠️ 依赖店铺服务
- ⚠️ 被订单服务依赖
- ✅ 业务边界清晰

**拆分范围**：
- 商品管理
- 库存管理
- 商品选项配置

**服务接口**：
```
POST   /api/v1/products
GET    /api/v1/products/{id}
PUT    /api/v1/products/{id}
DELETE /api/v1/products/{id}
GET    /api/v1/shops/{shopId}/products
PATCH  /api/v1/products/{id}/stock
PATCH  /api/v1/products/{id}/status
```

**数据迁移**：
- 迁移 `product_products` 表
- 迁移 `product_option_categories` 表
- 迁移 `product_options` 表
- 迁移 `product_tags` 表（商品关联）

**服务间通信**：
- 同步调用店铺服务验证shopId
- 发布库存变更事件

**服务配置**：
```yaml
server:
  port: 8003
  name: product-service

database:
  host: localhost
  port: 3306
  name: product_db
  user: root
  password: password

shop_service:
  url: http://localhost:8002

event_bus:
  type: kafka
  brokers:
    - localhost:9092
  topic: product-events
```

---

##### 第四批：订单服务（Order Service）

**拆分理由**：
- ⚠️ 强依赖其他服务
- ⚠️ 业务复杂度高
- ✅ 核心业务流程

**拆分范围**：
- 订单管理
- 订单状态流转
- 订单历史记录

**服务接口**：
```
POST   /api/v1/orders
GET    /api/v1/orders/{id}
PATCH  /api/v1/orders/{id}/status
DELETE /api/v1/orders/{id}
GET    /api/v1/shops/{shopId}/orders
GET    /api/v1/users/{userId}/orders
GET    /api/v1/shops/{shopId}/orders/unfinished
```

**数据迁移**：
- 迁移 `order_orders` 表
- 迁移 `order_items` 表
- 迁移 `order_item_options` 表
- 迁移 `order_status_logs` 表

**服务间通信**：
- 同步调用商品服务获取商品信息
- 同步调用商品服务扣减库存
- 同步调用店铺服务验证shopId
- 同步调用用户服务验证userId
- 发布订单状态变更事件

**服务配置**：
```yaml
server:
  port: 8004
  name: order-service

database:
  host: localhost
  port: 3306
  name: order_db
  user: root
  password: password

product_service:
  url: http://localhost:8003

shop_service:
  url: http://localhost:8002

user_service:
  url: http://localhost:8001

event_bus:
  type: kafka
  brokers:
    - localhost:9092
  topics:
    order_events: order-events
    product_events: product-events
```

#### 技术准备

- [ ] 搭建API网关（Kong/Nginx/Traefik）
- [ ] 实现服务注册发现（Consul/Etcd）
- [ ] 引入配置中心（Apollo/Nacos）
- [ ] 实现分布式追踪（Jaeger/Zipkin）
- [ ] 搭建监控告警系统（Prometheus/Grafana）
- [ ] 引入日志收集系统（ELK/Loki）

---

### 6.3 第三阶段：微服务架构落地（6-12个月）

**目标**：完成微服务架构的全面落地

**时间规划**：6-8个月

#### 最终架构

```
┌─────────────────────────────────────────────────────────┐
│                    API Gateway                        │
│              (Kong / Nginx / Traefik)               │
│                                                      │
│  功能：                                                │
│  - 路由转发                                          │
│  - 负载均衡                                          │
│  - 认证授权                                          │
│  - 限流熔断                                          │
│  - 请求日志                                          │
└─────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ 用户服务      │ │ 店铺服务      │ │ 商品服务      │
│ User Service │ │ Shop Service │ │Product Service│
│              │ │              │ │              │
│ Port: 8001   │ │ Port: 8002   │ │ Port: 8003   │
│              │ │              │ │              │
│ - 用户管理    │ │ - 店铺管理    │ │ - 商品管理    │
│ - 用户认证    │ │ - 标签管理    │ │ - 库存管理    │
│ - 用户权限    │ │ - 订单流转    │ │ - 选项配置    │
└──────────────┘ └──────────────┘ └──────────────┘
        │                 │                 │
        │                 │                 │
        └─────────────────┼─────────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │ 订单服务      │
                  │ Order Service│
                  │              │
                  │ Port: 8004   │
                  │              │
                  │ - 订单管理    │
                  │ - 状态流转    │
                  │ - 历史记录    │
                  └──────────────┘
                          │
                          ▼
                  ┌──────────────┐
                  │ 事件总线      │
                  │Event Bus     │
                  │ (Kafka/RabbitMQ)│
                  │              │
                  │ - 订单事件    │
                  │ - 库存事件    │
                  │ - 状态事件    │
                  └──────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ 用户数据库    │ │ 店铺数据库    │ │ 商品数据库    │
│  user_db     │ │  shop_db     │ │ product_db   │
│              │ │              │ │              │
│ MySQL/PG     │ │ MySQL/PG     │ │ MySQL/PG     │
└──────────────┘ └──────────────┘ └──────────────┘
                                          │
                                          ▼
                                  ┌──────────────┐
                                  │ 订单数据库    │
                                  │  order_db    │
                                  │              │
                                  │ MySQL/PG     │
                                  └──────────────┘
```

#### 部署架构

```
┌─────────────────────────────────────────────────────────┐
│                    Kubernetes                       │
│                                                      │
│  ┌─────────────────────────────────────────────┐        │
│  │              Namespace: production         │        │
│  │                                        │        │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐│        │
│  │  │ User Pod │ │ Shop Pod │ │Product Pod││        │
│  │  │ 3 replicas│ │ 3 replicas│ │ 3 replicas││        │
│  │  └──────────┘ └──────────┘ └──────────┘│        │
│  │                                        │        │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐│        │
│  │  │Order Pod │ │ Gateway  │ │  Kafka   ││        │
│  │  │ 3 replicas│ │ 2 replicas│ │ 3 replicas││        │
│  │  └──────────┘ └──────────┘ └──────────┘│        │
│  │                                        │        │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐│        │
│  │  │MySQL Pod │ │Redis Pod │ │Prometheus││        │
│  │  │ 1 replica │ │ 1 replica │ │ 1 replica ││        │
│  │  └──────────┘ └──────────┘ └──────────┘│        │
│  └─────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────┘
```

#### 关键任务

- [ ] 完成所有服务的拆分
- [ ] 实现服务间的完整通信
- [ ] 实现分布式事务
- [ ] 完善监控告警
- [ ] 实现自动化部署
- [ ] 完善文档和培训

---

## 7. 服务间通信方式

### 7.1 同步通信（REST API）

#### 适用场景
- 需要实时返回结果
- 强一致性要求
- 查询操作
- 简单的CRUD操作

#### 实现方式

##### ProductClient接口定义
```go
package acl

import (
    "context"
    "orderease/application/dto"
    "orderease/domain/shared"
)

type ProductClient interface {
    GetProduct(ctx context.Context, productID shared.ID) (*dto.ProductDTO, error)
    GetProducts(ctx context.Context, shopID uint64, page, pageSize int) (*dto.ProductListDTO, error)
    DecreaseStock(ctx context.Context, productID shared.ID, quantity int) error
    IncreaseStock(ctx context.Context, productID shared.ID, quantity int) error
    CheckStock(ctx context.Context, productID shared.ID, quantity int) (bool, error)
}
```

##### HTTP客户端实现
```go
package acl

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "orderease/application/dto"
    "orderease/domain/shared"
)

type HTTPProductClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
    retryCount int
    retryDelay time.Duration
}

func NewHTTPProductClient(baseURL string) *HTTPProductClient {
    return &HTTPProductClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        timeout:    30 * time.Second,
        retryCount: 3,
        retryDelay: 100 * time.Millisecond,
    }
}

func (c *HTTPProductClient) GetProduct(ctx context.Context, productID shared.ID) (*dto.ProductDTO, error) {
    url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, productID)

    var result dto.ProductDTO
    err := c.doRequest(ctx, "GET", url, nil, &result)
    if err != nil {
        return nil, fmt.Errorf("获取商品失败: %w", err)
    }

    return &result, nil
}

func (c *HTTPProductClient) DecreaseStock(ctx context.Context, productID shared.ID, quantity int) error {
    url := fmt.Sprintf("%s/api/v1/products/%s/stock", c.baseURL, productID)

    reqBody := map[string]interface{}{
        "quantity": quantity,
        "action":  "decrease",
    }

    return c.doRequest(ctx, "PATCH", url, reqBody, nil)
}

func (c *HTTPProductClient) CheckStock(ctx context.Context, productID shared.ID, quantity int) (bool, error) {
    url := fmt.Sprintf("%s/api/v1/products/%s/stock/check", c.baseURL, productID)

    reqBody := map[string]interface{}{
        "quantity": quantity,
    }

    var result struct {
        Available bool `json:"available"`
    }

    err := c.doRequest(ctx, "POST", url, reqBody, &result)
    if err != nil {
        return false, fmt.Errorf("检查库存失败: %w", err)
    }

    return result.Available, nil
}

func (c *HTTPProductClient) doRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
    var reqBody io.Reader

    if body != nil {
        data, err := json.Marshal(body)
        if err != nil {
            return fmt.Errorf("序列化请求体失败: %w", err)
        }
        reqBody = bytes.NewReader(data)
    }

    var lastErr error
    for i := 0; i < c.retryCount; i++ {
        if i > 0 {
            time.Sleep(c.retryDelay)
        }

        req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
        if err != nil {
            lastErr = fmt.Errorf("创建请求失败: %w", err)
            continue
        }

        if body != nil {
            req.Header.Set("Content-Type", "application/json")
        }

        resp, err := c.httpClient.Do(req)
        if err != nil {
            lastErr = fmt.Errorf("请求失败: %w", err)
            continue
        }

        defer resp.Body.Close()

        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            if result != nil {
                if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
                    lastErr = fmt.Errorf("解析响应失败: %w", err)
                    continue
                }
            }
            return nil
        }

        respBody, _ := io.ReadAll(resp.Body)
        lastErr = fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
    }

    return lastErr
}
```

##### 使用熔断器
```go
package acl

import (
    "github.com/sony/gobreaker"
)

type CircuitBreakerProductClient struct {
    client    ProductClient
    cb        *gobreaker.CircuitBreaker
}

func NewCircuitBreakerProductClient(client ProductClient) *CircuitBreakerProductClient {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "ProductClient",
        MaxRequests: 5,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 5 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            fmt.Printf("熔断器状态变更: %s, %s -> %s\n", name, from, to)
        },
    })

    return &CircuitBreakerProductClient{
        client: client,
        cb:     cb,
    }
}

func (c *CircuitBreakerProductClient) GetProduct(ctx context.Context, productID shared.ID) (*dto.ProductDTO, error) {
    var result *dto.ProductDTO
    _, err := c.cb.Execute(func() (interface{}, error) {
        dto, err := c.client.GetProduct(ctx, productID)
        if err != nil {
            return nil, err
        }
        result = dto
        return nil, nil
    })

    if err != nil {
        return nil, err
    }

    return result, nil
}
```

---

### 7.2 异步通信（事件驱动）

#### 适用场景
- 最终一致性可接受
- 解耦服务依赖
- 通知类操作
- 复杂的业务流程

#### 事件定义

##### 订单创建事件
```go
package events

import (
    "orderease/domain/shared"
    "time"
)

type OrderCreatedEvent struct {
    EventID   string            `json:"event_id"`
    EventType string            `json:"event_type"`
    Timestamp time.Time         `json:"timestamp"`
    OrderID   shared.ID         `json:"order_id"`
    UserID    shared.ID         `json:"user_id"`
    ShopID    uint64            `json:"shop_id"`
    TotalPrice shared.Price      `json:"total_price"`
    Items     []OrderItemEvent  `json:"items"`
}

type OrderItemEvent struct {
    ProductID  shared.ID `json:"product_id"`
    Quantity   int      `json:"quantity"`
    Price      shared.Price `json:"price"`
    TotalPrice shared.Price `json:"total_price"`
    Options    []OrderItemOptionEvent `json:"options"`
}

type OrderItemOptionEvent struct {
    CategoryID      shared.ID `json:"category_id"`
    OptionID        shared.ID `json:"option_id"`
    OptionName      string    `json:"option_name"`
    CategoryName    string    `json:"category_name"`
    PriceAdjustment float64   `json:"price_adjustment"`
}
```

##### 订单状态变更事件
```go
type OrderStatusChangedEvent struct {
    EventID   string    `json:"event_id"`
    EventType string    `json:"event_type"`
    Timestamp time.Time `json:"timestamp"`
    OrderID   shared.ID `json:"order_id"`
    ShopID    uint64    `json:"shop_id"`
    OldStatus int       `json:"old_status"`
    NewStatus int       `json:"new_status"`
}
```

##### 库存变更事件
```go
type StockChangedEvent struct {
    EventID   string    `json:"event_id"`
    EventType string    `json:"event_type"`
    Timestamp time.Time `json:"timestamp"`
    ProductID shared.ID `json:"product_id"`
    ShopID    uint64    `json:"shop_id"`
    OldStock  int       `json:"old_stock"`
    NewStock  int       `json:"new_stock"`
    Change    int       `json:"change"`
}
```

##### 库存不足事件
```go
type StockInsufficientEvent struct {
    EventID   string    `json:"event_id"`
    EventType string    `json:"event_type"`
    Timestamp time.Time `json:"timestamp"`
    OrderID   shared.ID `json:"order_id"`
    ProductID shared.ID `json:"product_id"`
    ShopID    uint64    `json:"shop_id"`
    Required  int       `json:"required"`
    Available int       `json:"available"`
}
```

#### 事件发布

##### EventPublisher接口
```go
package events

import (
    "context"
)

type EventPublisher interface {
    Publish(ctx context.Context, event interface{}) error
    PublishBatch(ctx context.Context, events []interface{}) error
}
```

##### Kafka实现
```go
package events

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/segmentio/kafka-go"
)

type KafkaEventPublisher struct {
    producer *kafka.Writer
    topic    string
}

func NewKafkaEventPublisher(brokers []string, topic string) (*KafkaEventPublisher, error) {
    producer := &kafka.Writer{
        Addr:     kafka.TCP(brokers...),
        Topic:    topic,
        Balancer: &kafka.LeastBytes{},
        Compression: kafka.Snappy,
        RequiredAcks: kafka.RequireAll,
    }

    return &KafkaEventPublisher{
        producer: producer,
        topic:    topic,
    }, nil
}

func (p *KafkaEventPublisher) Publish(ctx context.Context, event interface{}) error {
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("序列化事件失败: %w", err)
    }

    msg := kafka.Message{
        Key:   []byte(fmt.Sprintf("%v", event)),
        Value: data,
    }

    if err := p.producer.WriteMessages(ctx, msg); err != nil {
        return fmt.Errorf("发布事件失败: %w", err)
    }

    return nil
}

func (p *KafkaEventPublisher) PublishBatch(ctx context.Context, events []interface{}) error {
    messages := make([]kafka.Message, len(events))

    for i, event := range events {
        data, err := json.Marshal(event)
        if err != nil {
            return fmt.Errorf("序列化事件失败: %w", err)
        }

        messages[i] = kafka.Message{
            Key:   []byte(fmt.Sprintf("%v", event)),
            Value: data,
        }
    }

    if err := p.producer.WriteMessages(ctx, messages...); err != nil {
        return fmt.Errorf("批量发布事件失败: %w", err)
    }

    return nil
}

func (p *KafkaEventPublisher) Close() error {
    return p.producer.Close()
}
```

#### 事件消费

##### EventConsumer接口
```go
package events

import (
    "context"
)

type EventConsumer interface {
    Subscribe(topic string, handler func(event interface{}) error) error
    SubscribeBatch(topic string, handler func(events []interface{}) error) error
    Close() error
}
```

##### Kafka实现
```go
package events

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/segmentio/kafka-go"
)

type KafkaEventConsumer struct {
    reader   *kafka.Reader
    handlers map[string]func(event interface{}) error
}

func NewKafkaEventConsumer(brokers []string, topic, groupID string) *KafkaEventConsumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  brokers,
        Topic:    topic,
        GroupID:  groupID,
        MinBytes: 10e3,
        MaxBytes: 10e6,
    })

    return &KafkaEventConsumer{
        reader:   reader,
        handlers: make(map[string]func(event interface{}) error),
    }
}

func (c *KafkaEventConsumer) Subscribe(topic string, handler func(event interface{}) error) error {
    c.handlers[topic] = handler
    return nil
}

func (c *KafkaEventConsumer) Start(ctx context.Context) error {
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            if err == context.Canceled {
                return nil
            }
            log.Printf("读取消息失败: %v", err)
            continue
        }

        var event interface{}
        if err := json.Unmarshal(msg.Value, &event); err != nil {
            log.Printf("解析消息失败: %v", err)
            continue
        }

        eventType := c.getEventType(event)
        handler, exists := c.handlers[eventType]
        if !exists {
            log.Printf("未找到事件处理器: %s", eventType)
            continue
        }

        if err := handler(event); err != nil {
            log.Printf("处理事件失败: %v", err)
        }
    }
}

func (c *KafkaEventConsumer) getEventType(event interface{}) string {
    if m, ok := event.(map[string]interface{}); ok {
        if eventType, ok := m["event_type"].(string); ok {
            return eventType
        }
    }
    return "unknown"
}

func (c *KafkaEventConsumer) Close() error {
    return c.reader.Close()
}
```

##### 订单事件处理器
```go
package handlers

import (
    "context"
    "log"

    "orderease/application/services"
    "orderease/domain/events"
)

type OrderEventHandler struct {
    orderService services.OrderService
}

func NewOrderEventHandler(orderService services.OrderService) *OrderEventHandler {
    return &OrderEventHandler{
        orderService: orderService,
    }
}

func (h *OrderEventHandler) HandleOrderCreated(event events.OrderCreatedEvent) error {
    log.Printf("处理订单创建事件: %+v", event)

    // 处理订单创建后的业务逻辑
    // 例如：发送通知、更新统计等

    return nil
}

func (h *OrderEventHandler) HandleOrderStatusChanged(event events.OrderStatusChangedEvent) error {
    log.Printf("处理订单状态变更事件: %+v", event)

    // 处理订单状态变更后的业务逻辑
    // 例如：发送通知、更新统计等

    return nil
}
```

##### 商品事件处理器
```go
package handlers

import (
    "context"
    "log"

    "orderease/application/services"
    "orderease/domain/events"
)

type ProductEventHandler struct {
    productService services.ProductService
    eventPublisher  events.EventPublisher
}

func NewProductEventHandler(
    productService services.ProductService,
    eventPublisher events.EventPublisher,
) *ProductEventHandler {
    return &ProductEventHandler{
        productService: productService,
        eventPublisher:  eventPublisher,
    }
}

func (h *ProductEventHandler) HandleOrderCreated(event events.OrderCreatedEvent) error {
    log.Printf("处理订单创建事件，扣减库存: %+v", event)

    for _, item := range event.Items {
        // 扣减库存
        if err := h.productService.DecreaseStock(context.Background(), item.ProductID, item.Quantity); err != nil {
            // 发布库存不足事件
            insufficientEvent := events.StockInsufficientEvent{
                EventID:   generateEventID(),
                EventType: "stock.insufficient",
                Timestamp: time.Now(),
                OrderID:   event.OrderID,
                ProductID: item.ProductID,
                ShopID:    event.ShopID,
                Required:  item.Quantity,
                Available: 0,
            }

            if err := h.eventPublisher.Publish(context.Background(), insufficientEvent); err != nil {
                log.Printf("发布库存不足事件失败: %v", err)
            }

            return fmt.Errorf("扣减库存失败: %w", err)
        }
    }

    // 发布库存扣减成功事件
    stockChangedEvent := events.StockChangedEvent{
        EventID:   generateEventID(),
        EventType: "stock.changed",
        Timestamp: time.Now(),
        OrderID:   event.OrderID,
        ShopID:    event.ShopID,
    }

    return h.eventPublisher.Publish(context.Background(), stockChangedEvent)
}

func generateEventID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

---

## 8. 数据一致性策略

### 8.1 Saga模式（分布式事务）

#### Saga模式概述

Saga模式是一种分布式事务解决方案，通过将一个长事务拆分为多个本地事务，并为每个本地事务定义补偿操作来实现最终一致性。

#### Saga实现

##### SagaStep定义
```go
package saga

import (
    "context"
)

type SagaStep struct {
    Name      string
    Execute   func(ctx context.Context) error
    Compensate func(ctx context.Context) error
}

type Saga struct {
    steps []SagaStep
}

func NewSaga() *Saga {
    return &Saga{
        steps: make([]SagaStep, 0),
    }
}

func (s *Saga) AddStep(step SagaStep) {
    s.steps = append(s.steps, step)
}

func (s *Saga) Execute(ctx context.Context) error {
    // 执行所有步骤
    for i, step := range s.steps {
        if err := step.Execute(ctx); err != nil {
            // 执行失败，回滚已执行的步骤
            log.Printf("Saga步骤执行失败: %s, 开始回滚", step.Name)
            for j := i - 1; j >= 0; j-- {
                if err := s.steps[j].Compensate(ctx); err != nil {
                    log.Printf("Saga步骤补偿失败: %s", s.steps[j].Name)
                }
            }
            return fmt.Errorf("Saga执行失败: %w", err)
        }
        log.Printf("Saga步骤执行成功: %s", step.Name)
    }
    return nil
}
```

##### 订单创建Saga
```go
package saga

import (
    "context"
    "fmt"
    "log"

    "orderease/application/dto"
    "orderease/application/services"
    "orderease/domain/events"
    "orderease/domain/order"
    "orderease/domain/shared"
)

type CreateOrderSaga struct {
    orderService   services.OrderService
    productClient  acl.ProductClient
    shopClient     acl.ShopClient
    userClient     acl.UserClient
    eventPublisher events.EventPublisher
}

func NewCreateOrderSaga(
    orderService services.OrderService,
    productClient acl.ProductClient,
    shopClient acl.ShopClient,
    userClient acl.UserClient,
    eventPublisher events.EventPublisher,
) *CreateOrderSaga {
    return &CreateOrderSaga{
        orderService:   orderService,
        productClient:  productClient,
        shopClient:     shopClient,
        userClient:     userClient,
        eventPublisher: eventPublisher,
    }
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    saga := NewSaga()

    // 步骤1：验证用户
    saga.AddStep(SagaStep{
        Name: "验证用户",
        Execute: func(ctx context.Context) error {
            user, err := s.userClient.GetUser(ctx, req.UserID)
            if err != nil {
                return fmt.Errorf("用户不存在: %w", err)
            }
            if user == nil {
                return fmt.Errorf("用户不存在")
            }
            return nil
        },
        Compensate: func(ctx context.Context) error {
            // 用户验证不需要补偿
            return nil
        },
    })

    // 步骤2：验证店铺
    saga.AddStep(SagaStep{
        Name: "验证店铺",
        Execute: func(ctx context.Context) error {
            shop, err := s.shopClient.GetShop(ctx, req.ShopID)
            if err != nil {
                return fmt.Errorf("店铺不存在: %w", err)
            }
            if shop == nil {
                return fmt.Errorf("店铺不存在")
            }
            return nil
        },
        Compensate: func(ctx context.Context) error {
            // 店铺验证不需要补偿
            return nil
        },
    })

    // 步骤3：验证商品并扣减库存
    var products []*dto.ProductDTO
    for _, item := range req.Items {
        item := item // 创建局部变量
        saga.AddStep(SagaStep{
            Name: fmt.Sprintf("扣减商品库存: %s", item.ProductID),
            Execute: func(ctx context.Context) error {
                // 获取商品信息
                product, err := s.productClient.GetProduct(ctx, item.ProductID)
                if err != nil {
                    return fmt.Errorf("获取商品信息失败: %w", err)
                }
                if product == nil {
                    return fmt.Errorf("商品不存在")
                }

                // 检查库存
                if product.Stock < item.Quantity {
                    return fmt.Errorf("商品库存不足")
                }

                // 扣减库存
                if err := s.productClient.DecreaseStock(ctx, item.ProductID, item.Quantity); err != nil {
                    return fmt.Errorf("扣减库存失败: %w", err)
                }

                products = append(products, product)
                return nil
            },
            Compensate: func(ctx context.Context) error {
                // 恢复库存
                if err := s.productClient.IncreaseStock(ctx, item.ProductID, item.Quantity); err != nil {
                    log.Printf("恢复库存失败: %v", err)
                }
                return nil
            },
        })
    }

    // 步骤4：创建订单
    var createdOrder *dto.OrderResponse
    saga.AddStep(SagaStep{
        Name: "创建订单",
        Execute: func(ctx context.Context) error {
            order, err := s.orderService.CreateOrder(req)
            if err != nil {
                return fmt.Errorf("创建订单失败: %w", err)
            }
            createdOrder = order
            return nil
        },
        Compensate: func(ctx context.Context) error {
            // 删除订单
            if createdOrder != nil {
                if err := s.orderService.DeleteOrder(ctx, createdOrder.ID, req.ShopID); err != nil {
                    log.Printf("删除订单失败: %v", err)
                }
            }
            return nil
        },
    })

    // 步骤5：发布订单创建事件
    saga.AddStep(SagaStep{
        Name: "发布订单创建事件",
        Execute: func(ctx context.Context) error {
            event := events.OrderCreatedEvent{
                EventID:   generateEventID(),
                EventType: "order.created",
                Timestamp: time.Now(),
                OrderID:   createdOrder.ID,
                UserID:    req.UserID,
                ShopID:    req.ShopID,
                TotalPrice: createdOrder.TotalPrice,
                Items:     convertToOrderItemEvents(req.Items, products),
            }

            if err := s.eventPublisher.Publish(ctx, event); err != nil {
                return fmt.Errorf("发布事件失败: %w", err)
            }
            return nil
        },
        Compensate: func(ctx context.Context) error {
            // 事件发布不需要补偿
            return nil
        },
    })

    // 执行Saga
    if err := saga.Execute(ctx); err != nil {
        return nil, err
    }

    return createdOrder, nil
}

func convertToOrderItemEvents(items []dto.OrderItemRequest, products []*dto.ProductDTO) []events.OrderItemEvent {
    result := make([]events.OrderItemEvent, len(items))
    for i, item := range items {
        result[i] = events.OrderItemEvent{
            ProductID:  item.ProductID,
            Quantity:   item.Quantity,
            Price:      item.Price,
            TotalPrice: item.Price.Multiply(item.Quantity),
        }
    }
    return result
}
```

---

### 8.2 最终一致性（事件驱动）

#### 最终一致性概述

最终一致性是一种弱一致性模型，允许系统在一段时间内处于不一致状态，但最终会达到一致状态。

#### 订单创建流程（最终一致性）

##### 订单服务创建订单
```go
package services

import (
    "context"
    "fmt"
    "log"

    "orderease/application/dto"
    "orderease/domain/events"
    "orderease/domain/order"
    "orderease/domain/shared"
)

func (s *OrderService) CreateOrderWithEventualConsistency(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    // 1. 创建订单（初始状态为"待确认"）
    items := make([]order.OrderItem, len(req.Items))
    for i, itemReq := range req.Items {
        items[i] = order.OrderItem{
            ProductID: itemReq.ProductID,
            Quantity:  itemReq.Quantity,
            Price:     shared.Price(itemReq.Price),
        }
    }

    ord, err := order.NewOrder(req.UserID, req.ShopID, items, req.Remark)
    if err != nil {
        return nil, err
    }

    // 设置初始状态为待确认
    ord.Status = order.OrderStatusPending

    // 保存订单
    if err := s.orderRepo.Save(ord); err != nil {
        return nil, fmt.Errorf("保存订单失败: %w", err)
    }

    // 2. 发布订单创建事件
    event := events.OrderCreatedEvent{
        EventID:   generateEventID(),
        EventType: "order.created",
        Timestamp: time.Now(),
        OrderID:   ord.ID,
        UserID:    req.UserID,
        ShopID:    req.ShopID,
        TotalPrice: ord.TotalPrice,
        Items:     convertToOrderItemEvents(req.Items),
    }

    if err := s.eventPublisher.Publish(ctx, event); err != nil {
        log.Printf("发布订单创建事件失败: %v", err)
        // 事件发布失败不影响订单创建，可以通过重试机制补偿
    }

    log.Printf("订单创建成功，等待库存扣减: %+v", ord)

    return &dto.OrderResponse{
        ID:         ord.ID,
        UserID:     ord.UserID,
        ShopID:     ord.ShopID,
        TotalPrice: ord.TotalPrice,
        Status:     ord.Status,
        Remark:     ord.Remark,
        CreatedAt:  ord.CreatedAt,
        UpdatedAt:  ord.UpdatedAt,
    }, nil
}
```

##### 商品服务监听订单创建事件
```go
package handlers

import (
    "context"
    "fmt"
    "log"

    "orderease/application/services"
    "orderease/domain/events"
)

type ProductEventHandler struct {
    productService services.ProductService
    eventPublisher  events.EventPublisher
}

func NewProductEventHandler(
    productService services.ProductService,
    eventPublisher events.EventPublisher,
) *ProductEventHandler {
    return &ProductEventHandler{
        productService: productService,
        eventPublisher:  eventPublisher,
    }
}

func (h *ProductEventHandler) HandleOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error {
    log.Printf("处理订单创建事件，扣减库存: %+v", event)

    // 扣减库存
    for _, item := range event.Items {
        // 检查库存
        available, err := h.productService.CheckStock(ctx, item.ProductID, item.Quantity)
        if err != nil {
            log.Printf("检查库存失败: %v", err)
            return err
        }

        if !available {
            // 发布库存不足事件
            insufficientEvent := events.StockInsufficientEvent{
                EventID:   generateEventID(),
                EventType: "stock.insufficient",
                Timestamp: time.Now(),
                OrderID:   event.OrderID,
                ProductID: item.ProductID,
                ShopID:    event.ShopID,
                Required:  item.Quantity,
                Available: 0,
            }

            if err := h.eventPublisher.Publish(ctx, insufficientEvent); err != nil {
                log.Printf("发布库存不足事件失败: %v", err)
            }

            return fmt.Errorf("商品库存不足: %s", item.ProductID)
        }

        // 扣减库存
        if err := h.productService.DecreaseStock(ctx, item.ProductID, item.Quantity); err != nil {
            log.Printf("扣减库存失败: %v", err)
            return err
        }
    }

    // 发布库存扣减成功事件
    stockChangedEvent := events.StockChangedEvent{
        EventID:   generateEventID(),
        EventType: "stock.changed",
        Timestamp: time.Now(),
        OrderID:   event.OrderID,
        ShopID:    event.ShopID,
    }

    if err := h.eventPublisher.Publish(ctx, stockChangedEvent); err != nil {
        log.Printf("发布库存变更事件失败: %v", err)
    }

    log.Printf("库存扣减成功，订单ID: %s", event.OrderID)
    return nil
}
```

##### 订单服务监听库存扣减成功事件
```go
package handlers

import (
    "context"
    "log"

    "orderease/application/services"
    "orderease/domain/events"
    "orderease/domain/order"
)

type OrderEventHandler struct {
    orderService services.OrderService
}

func NewOrderEventHandler(orderService services.OrderService) *OrderEventHandler {
    return &OrderEventHandler{
        orderService: orderService,
    }
}

func (h *OrderEventHandler) HandleStockChanged(ctx context.Context, event events.StockChangedEvent) error {
    log.Printf("处理库存变更事件，更新订单状态: %+v", event)

    // 获取订单
    ord, err := h.orderService.GetOrderByID(ctx, event.OrderID)
    if err != nil {
        log.Printf("获取订单失败: %v", err)
        return err
    }

    // 更新订单状态为"已确认"
    if ord.Status == order.OrderStatusPending {
        if err := h.orderService.UpdateOrderStatus(ctx, event.OrderID, ord.ShopID, order.OrderStatusAccepted); err != nil {
            log.Printf("更新订单状态失败: %v", err)
            return err
        }

        log.Printf("订单状态更新成功，订单ID: %s，新状态: %d", event.OrderID, order.OrderStatusAccepted)
    }

    return nil
}

func (h *OrderEventHandler) HandleStockInsufficient(ctx context.Context, event events.StockInsufficientEvent) error {
    log.Printf("处理库存不足事件，取消订单: %+v", event)

    // 获取订单
    ord, err := h.orderService.GetOrderByID(ctx, event.OrderID)
    if err != nil {
        log.Printf("获取订单失败: %v", err)
        return err
    }

    // 更新订单状态为"已取消"
    if ord.Status == order.OrderStatusPending {
        if err := h.orderService.UpdateOrderStatus(ctx, event.OrderID, ord.ShopID, order.OrderStatusCanceled); err != nil {
            log.Printf("更新订单状态失败: %v", err)
            return err
        }

        log.Printf("订单状态更新成功，订单ID: %s，新状态: %d", event.OrderID, order.OrderStatusCanceled)
    }

    return nil
}
```

#### 事件重试机制

```go
package events

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

type RetryableEventConsumer struct {
    consumer  EventConsumer
    maxRetry  int
    retryDelay time.Duration
}

func NewRetryableEventConsumer(consumer EventConsumer, maxRetry int, retryDelay time.Duration) *RetryableEventConsumer {
    return &RetryableEventConsumer{
        consumer:  consumer,
        maxRetry:  maxRetry,
        retryDelay: retryDelay,
    }
}

func (c *RetryableEventConsumer) SubscribeWithRetry(topic string, handler func(event interface{}) error) error {
    return c.consumer.Subscribe(topic, func(event interface{}) error {
        var lastErr error
        for i := 0; i < c.maxRetry; i++ {
            if i > 0 {
                log.Printf("重试处理事件，第%d次: %+v", i, event)
                time.Sleep(c.retryDelay)
            }

            if err := handler(event); err != nil {
                lastErr = err
                log.Printf("处理事件失败: %v", err)
                continue
            }

            return nil
        }

        return lastErr
    })
}
```

---

## 9. 实施计划总结

### 9.1 第一阶段（1-2个月）：架构优化

**目标**：强化现有DDD四层架构，建立清晰的上下文边界，为未来服务拆分奠定基础

**任务清单**（调整后）：
- [ ] **修复服务依赖类型错误**（最高优先级）
  - OrderService中的userRepo参数类型应为user.UserRepository
  - ProductService中的orderItemRepo参数类型应为order.OrderItemRepository
- [ ] **引入Wire依赖注入框架**（基础设施准备）
  - 建立统一的依赖注入机制
  - 解耦组件间的直接依赖
- [ ] **统一常量到共享内核**
  - 将订单状态、商品状态等常量迁移到domain/shared/constants.go
  - 形成真正的共享内核，确保上下文间一致性
- [ ] **增强领域模型，迁移业务逻辑到实体**
  - 将业务逻辑从Application Service迁移到Domain Entity
  - 实现领域服务处理跨聚合业务逻辑
  - 强化聚合根的业务规则约束
- [ ] **建立统一错误处理机制**
  - 定义领域错误类型和API错误响应
  - 实现跨上下文调用的容错机制
- [ ] **明确定义上下文接口契约**
  - 通过接口文档或代码契约定义跨上下文调用规范
  - 为第二阶段实现防腐层做准备
- [ ] **制定具体测试策略**（单元测试覆盖率≥80%）
  - 核心领域逻辑重点测试
  - 跨上下文交互集成测试
  - 使用接口mock隔离上下文依赖

**当前状态分析**（修正后）：
- [x] **DDD四层架构基础良好**：现有结构已体现上下文边界，无需大规模重组
- [x] **部分共享内核已提取**：ID、Price类型已在domain/shared包
- [ ] **常量未统一到共享内核**：订单状态、商品状态等常量仍分散在models目录
- [ ] **服务依赖类型错误存在**：OrderService、ProductService中存在参数类型错误
- [ ] **领域模型贫血**：业务逻辑集中在Application Service，Domain Entity缺乏行为
- [ ] **跨上下文直接依赖**：缺乏防腐层隔离，上下文边界保护不足
- [ ] **统一错误处理机制缺失**：错误处理分散，缺乏标准化

**里程碑**（调整后）：
- **架构边界清晰**：上下文边界明确，跨上下文依赖受控
- **核心问题修复**：服务依赖类型错误修复，常量统一到共享内核
- **领域模型增强**：业务逻辑适当迁移到Domain Entity
- **基础设施完善**：Wire依赖注入框架引入，统一错误处理机制建立
- **质量保障体系**：单元测试覆盖率≥80%，集成测试覆盖跨上下文交互

### 9.1.1 第一阶段重构调整建议

#### 9.1.1.1 方案调整要点

1. **放弃大规模目录重组**：现有DDD四层架构已体现上下文边界，无需创建user-context/等目录结构
2. **调整防腐层实现时机**：将完整防腐层实现移至第二阶段，第一阶段聚焦定义接口契约
3. **优化技术准备顺序**：
   - 优先引入Wire依赖注入框架
   - 其次修复服务依赖类型错误
   - 然后建立统一错误处理机制
4. **重新评估数据库策略**：暂不实施表前缀策略，避免增加复杂度
5. **领域事件作为可选任务**：根据团队能力和时间决定是否第一阶段实施

#### 9.1.1.2 关键实施步骤

1. **基础设施层准备**（第1个月）
   - 引入Wire依赖注入框架
   - 建立统一错误处理机制
   - 实现请求追踪（Trace ID）

2. **架构优化**（第1-2个月）
   - 修复服务依赖类型错误
   - 统一常量到共享内核
   - 增强领域模型，迁移业务逻辑到实体
   - 清理循环依赖，强化上下文边界

3. **质量保证**（贯穿全程）
   - 制定测试策略，确保核心逻辑覆盖率≥80%
   - 明确定义上下文接口契约

#### 9.1.1.3 重构优先级（调整后）

1. **🔴 最高优先级**：
   - 修复服务依赖类型错误
   - 引入Wire依赖注入框架

2. **🟡 中等优先级**：
   - 统一常量到共享内核
   - 增强领域模型，迁移业务逻辑到实体
   - 建立统一错误处理机制

3. **🟢 可选任务**：
   - 引入领域事件机制（根据团队能力决定）
   - 明确定义上下文接口契约（为第二阶段准备）

#### 9.1.1.4 质量保证措施（细化）

1. **单元测试覆盖**（核心指标）
   - 领域模型逻辑：≥90%覆盖率
   - 应用服务层：≥80%覆盖率
   - 重点测试：业务规则、数据验证、状态转换

2. **集成测试策略**
   - 跨上下文调用验证
   - 防腐层接口契约测试
   - 错误处理和容错机制测试

3. **性能评估标准**
   - 跨上下文调用延迟：增加<20%
   - 系统吞吐量：保持原有水平±10%
   - 内存使用：无明显增加

---

### 9.2 第二阶段（3-4个月）：服务拆分准备

**目标**：为微服务拆分做好技术准备

**任务清单**：
- [ ] 拆分用户服务
  - [ ] 创建用户服务项目
  - [ ] 迁移用户相关代码
  - [ ] 实现用户服务API
  - [ ] 迁移用户数据库
  - [ ] 部署用户服务
- [ ] 拆分店铺服务
  - [ ] 创建店铺服务项目
  - [ ] 迁移店铺相关代码
  - [ ] 实现店铺服务API
  - [ ] 迁移店铺数据库
  - [ ] 部署店铺服务
- [ ] 搭建API网关
  - [ ] 选择API网关技术
  - [ ] 配置路由规则
  - [ ] 实现认证授权
  - [ ] 实现限流熔断
- [ ] 实现服务注册发现
  - [ ] 选择服务注册中心
  - [ ] 实现服务注册
  - [ ] 实现服务发现
  - [ ] 实现健康检查

**里程碑**：
- 用户服务独立部署
- 店铺服务独立部署
- API网关正常运行
- 服务注册发现正常工作

---

### 9.3 第三阶段（5-6个月）：核心服务拆分

**目标**：拆分商品服务和订单服务

**任务清单**：
- [ ] 拆分商品服务
  - [ ] 创建商品服务项目
  - [ ] 迁移商品相关代码
  - [ ] 实现商品服务API
  - [ ] 迁移商品数据库
  - [ ] 实现服务间通信
  - [ ] 部署商品服务
- [ ] 拆分订单服务
  - [ ] 创建订单服务项目
  - [ ] 迁移订单相关代码
  - [ ] 实现订单服务API
  - [ ] 迁移订单数据库
  - [ ] 实现服务间通信
  - [ ] 部署订单服务
- [ ] 实现事件总线
  - [ ] 选择消息中间件
  - [ ] 实现事件发布
  - [ ] 实现事件消费
  - [ ] 实现事件重试
- [ ] 实现分布式事务
  - [ ] 实现Saga模式
  - [ ] 实现补偿机制
  - [ ] 实现事务日志

**里程碑**：
- 商品服务独立部署
- 订单服务独立部署
- 事件总线正常运行
- 分布式事务正常工作

---

### 9.4 第四阶段（7-8个月）：优化完善

**目标**：优化性能，完善监控

**任务清单**：
- [ ] 性能优化
  - [ ] 数据库查询优化
  - [ ] 缓存优化
  - [ ] 接口响应优化
  - [ ] 并发优化
- [ ] 监控告警
  - [ ] 搭建监控系统
  - [ ] 配置监控指标
  - [ ] 配置告警规则
  - [ ] 实现日志收集
- [ ] 文档完善
  - [ ] API文档
  - [ ] 架构文档
  - [ ] 运维文档
  - [ ] 开发文档
- [ ] 团队培训
  - [ ] DDD培训
  - [ ] 微服务培训
  - [ ] 运维培训
  - [ ] 故障排查培训

**里程碑**：
- 系统性能达标
- 监控告警正常
- 文档完善
- 团队培训完成

---

### 9.5 时间规划总览

| 阶段 | 时间 | 主要任务 | 交付物 |
|-----|------|---------|--------|
| 第一阶段 | 第1-2个月 | 架构优化 | 清晰的上下文边界 |
| 第二阶段 | 第3-4个月 | 服务拆分准备 | 用户服务、店铺服务 |
| 第三阶段 | 第5-6个月 | 核心服务拆分 | 商品服务、订单服务 |
| 第四阶段 | 第7-8个月 | 优化完善 | 完整的微服务架构 |

---

## 附录

### A. 参考资料

1. 《领域驱动设计》- Eric Evans
2. 《实现领域驱动设计》- Vaughn Vernon
3. 《微服务架构设计模式》- Chris Richardson
4. 《Building Microservices》- Sam Newman

### B. 技术栈推荐

- **API网关**：Kong / Nginx / Traefik
- **服务注册发现**：Consul / Etcd / Nacos
- **配置中心**：Apollo / Nacos / Spring Cloud Config
- **消息中间件**：Kafka / RabbitMQ / RocketMQ
- **分布式追踪**：Jaeger / Zipkin / SkyWalking
- **监控告警**：Prometheus / Grafana / AlertManager
- **日志收集**：ELK / Loki / Fluentd
- **容器编排**：Kubernetes / Docker Swarm

### C. 最佳实践

1. **上下文边界清晰**：每个上下文应该有明确的职责边界
2. **防腐层隔离**：使用防腐层隔离外部上下文的影响
3. **事件驱动**：使用事件驱动实现服务间解耦
4. **最终一致性**：接受最终一致性，避免强一致性带来的复杂性
5. **监控可观测**：完善的监控和日志系统
6. **自动化部署**：使用CI/CD实现自动化部署
7. **文档完善**：保持文档的及时更新

### D. 重构实施指导

#### D.1 重构步骤

1. **代码审查阶段**
   - 检查并修复服务依赖类型错误
   - 识别跨上下文直接依赖
   - 评估现有业务逻辑分布

2. **防腐层实现**
   - 创建 `infrastructure/acl` 包
   - 实现各上下文的Client接口
   - 创建DTO和Mapper用于数据转换

3. **领域事件引入**
   - 创建 `domain/events` 包
   - 定义关键业务事件
   - 实现事件发布和订阅机制

4. **领域模型增强**
   - 将业务逻辑从Application Service迁移到Domain Entity
   - 实现领域服务处理跨聚合业务逻辑

#### D.2 重构验证

1. **功能验证**
   - 确保所有现有功能正常工作
   - 验证跨上下文交互的正确性

2. **性能验证**
   - 测试引入防腐层后的性能影响
   - 确保系统响应时间在可接受范围内

3. **质量验证**
   - 单元测试覆盖率应达到80%以上
   - 集成测试覆盖跨上下文交互

---

**文档版本**：v1.1
**创建日期**：2026-01-09
**最后更新**：2026-01-10
**维护者**：架构团队

## 10. 重构审查总结

### 10.1 审查概况

本次审查针对DDD战略设计方案的第一阶段单体服务重构方案进行了全面分析，发现文档中存在多处与实际情况不符的描述，需要调整优化方案以更符合当前代码库状态和实际需求。

### 10.2 发现的主要问题（更新后）

1. **目录结构重组声明不实**
   - 文档声称代码已按上下文重组（创建user-context/、shop-context/等目录）
   - 实际情况：代码库仍保持原有的DDD四层架构，未进行大规模目录重组
   - 风险：文档与实际情况不符，可能导致团队误解和错误实施决策

2. **共享内核常量未统一**
   - 文档描述：在shared/constants.go中定义订单状态、商品状态等常量
   - 实际情况：常量分散在models/目录（如models/order.go、models/product.go）
   - 问题：文档中描述的"已完成"状态与代码实际不符

3. **防腐层实现时机过早**
   - 文档要求：第一阶段实现完整的防腐层（ACL）
   - 问题分析：在单体架构中立即实现完整防腐层属于过度设计
   - 影响：增加不必要的抽象层和代码复杂度，降低开发效率

4. **技术准备顺序不合理**
   - 当前顺序：重组代码结构 → 实现防腐层 → 引入DI框架
   - 优化后顺序：引入依赖注入框架 → 修复服务依赖类型错误 → 建立统一错误处理机制

5. **数据库前缀策略缺乏必要性评估**
   - 文档建议：使用user_、shop_等表前缀区分上下文
   - 现状分析：现有表名简洁清晰，通过ShopID等字段已实现逻辑隔离
   - 风险：增加SQL编写和调试复杂度，可能破坏现有查询和关联关系

6. **服务依赖类型错误**（文档正确识别）
   - OrderService中的userRepo参数类型错误
   - ProductService中的orderItemRepo参数类型错误

7. **领域模型贫血**（文档正确识别）
   - 业务逻辑集中在Application Service中
   - Domain Entity未充分体现业务规则

### 10.3 修复优先级（调整后）

| 任务 | 原文档状态 | 调整建议 | 优先级 |
|------|-----------|----------|--------|
| 按上下文重组代码结构 | ✓ 已完成 | 放弃大规模重组，优化现有结构 | 🔴 高 |
| 修复服务依赖类型错误 | 待修复 | 立即实施，最高优先级 | 🔴 高 |
| 提取共享内核 | ✓ 已完成 | 补充常量统一到shared/包 | 🟡 中 |
| 实现防腐层（ACL） | 待实现 | 移至第二阶段，当前明确定义接口契约 | 🟡 中 |
| 引入领域事件机制 | 待实现 | 可选任务，根据团队能力决定 | 🟢 低 |
| 完善单元测试 | 待完成 | 制定具体测试策略和覆盖率目标 | 🟡 中 |
| 引入DI框架（Wire） | 未提及 | 提升为第一阶段核心任务 | 🔴 高 |

### 10.4 第一阶段核心目标重新定义

第一阶段（1-2个月）应聚焦于**强化现有架构，为未来拆分奠定基础**，而非进行颠覆式重构：

1. **基础设施准备**（第1个月）：
   - 引入Wire依赖注入框架
   - 建立统一错误处理机制
   - 实现请求追踪（Trace ID）

2. **架构优化**（第1-2个月）：
   - 修复服务依赖类型错误
   - 统一常量到共享内核
   - 增强领域模型，迁移业务逻辑到实体
   - 清理循环依赖，强化上下文边界

3. **质量保证**（贯穿全程）：
   - 制定测试策略，确保核心逻辑覆盖率≥80%
   - 明确定义上下文接口契约

**核心原则**：第一阶段的目标不是实现完美的微服务架构，而是在单体架构中**建立清晰的上下文边界**，为未来的服务拆分做好准备。
