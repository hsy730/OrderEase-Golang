# OrderEase DDD 小步重构计划

## 重构原则
1. **小步前进** - 每次改动最小化，可独立提交
2. **逻辑不变** - 重构不改变业务行为
3. **测试验证** - 每步完成后执行测试用例n
4. **可回滚** - 每步都是独立提交，出问题可快速回退

## 一、当前状态评估

### 已完成 ✅
- `domain/user/` 聚合（实体、值对象、仓储接口、领域服务）
- `domain/order/` 聚合根（实体 + 业务方法 + 领域服务）
- `domain/shared/value_objects/` (Phone, Password, OrderStatus)
- `utils/order_validation.go` (已清理，保留函数标记为 DEPRECATED)
- `models/shop_helpers.go` (已清理，保留函数标记为 DEPRECATED)
- 所有 Repository 实现（但返回 models.* 而非领域实体）
- **Step 1**: 提取 Shop 业务方法到独立函数 ✅
- **Step 2**: 统一密码验证规则 ✅
- **Step 3**: 移除 User.BeforeSave 钩子，密码哈希移到领域层 ✅
- **Step 4**: 创建 OrderStatus 值对象 ✅
- **Step 5**: 提取订单验证逻辑到独立函数 ✅
- **Step 6**: 提取库存扣减逻辑到独立函数 ✅
- **Step 7**: 创建 Order 聚合根（空壳）✅
- **Step 8**: 为 Order 添加业务方法 ✅
- **Step 12**: 移除 Shop.BeforeSave 钩子 ✅
- **Step 13**: 创建 Order 领域服务 ✅
- **Step 14**: 清理 utils 包中的领域逻辑 ✅

### DDD成熟度：75% (过渡阶段)

---

## 二、小步重构路线图

### 策略：采用「提取接口 → 适配器模式 → 逐步迁移」的方式
每步都是：**代码改动 → 运行测试 → git commit**

## 三、小步重构详细步骤

### Step 1: 提取 models.Shop 的业务方法到独立函数
**目标**: 将 `models.Shop` 的业务逻辑提取出来，不改变调用方式

**改动**:
- 创建 `src/utils/shop_helpers.go`，将以下方法移到独立函数：
  - `Shop.CheckPassword()` → `CheckShopPassword(shop, password) error`
  - `Shop.IsExpired()` → `IsShopExpired(shop) bool`
  - `Shop.RemainingDays()` → `GetShopRemainingDays(shop) int`
- 修改 `models.Shop` 保留方法，内部调用新函数
- 修改 `handlers/shop.go` 调用新函数

**验证**: 运行测试
**提交**: `refactor(shop): 提取 Shop 业务方法到独立函数`

---

### Step 2: 统一密码验证规则
**目标**: 统一 `utils.ValidatePassword` 和 `value_objects.NewPassword` 的规则

**密码规则说明**:
- **前端用户**: 6-20位，必须包含字母和数字，特殊字符可选
- **管理员/店主**: 8+位，必须包含大小写字母、数字和特殊字符

**改动**:
1. 修改 `domain/shared/value_objects/password.go`:
   - `NewPassword()`: 6-20位，字母+数字，支持特殊字符（前端用户）
   - `NewStrictPassword()`: 8+位，大小写+数字+特殊字符（管理员/店主）
   - `NewSimplePassword()`: 保持不变（6位简单密码）
2. 修改 `utils/password.go`:
   - `ValidatePassword()` 内部调用 `NewStrictPassword()`
   - 保持函数签名不变，确保兼容性

**验证**: 运行测试
**提交**: `refactor(password): 统一密码验证规则到值对象`

---

### Step 3: 移除 models.User.BeforeSave 的密码加密钩子
**目标**: 将密码加密逻辑移到领域层，解耦 models 层

**改动**:
1. 修改 `domain/user/user.go`:
   - 在 `ToModel()` 方法中添加密码 bcrypt 哈希逻辑
   - 检查密码是否已哈希（`$2a$` 前缀），避免重复哈希
2. 修改 `domain/user/service.go`:
   - `Register()` 方法使用 `NewStrictPassword()` 验证管理员密码
3. 移除 `models.User.BeforeSave` 钩子
4. 确保密码修改处使用强密码验证

**验证**: 运行测试
**提交**: `refactor(user): 移除 BeforeSave 钩子，将密码加密移到领域层`

---

### Step 4: 创建 OrderStatus 值对象
**目标**: 将订单状态相关逻辑封装为值对象

**改动**:
- 创建 `src/domain/shared/value_objects/order_status.go`
  - `type OrderStatus int`
  - `func (s OrderStatus) String() string`
  - `func (s OrderStatus) CanTransitionTo(to OrderStatus, flow OrderStatusFlow) bool`
- 修改 `handlers/order.go` 使用新值对象
- 保持 models.Order.Status 为 int 类型（数据库兼容）

**验证**: 运行测试
**提交**: `feat(domain): 添加 OrderStatus 值对象`

---

### Step 5: 提取订单验证逻辑到独立函数
**目标**: 将 `CreateOrder` 中的验证逻辑提取，不改变调用方式

**改动**:
- 创建 `src/utils/order_validation.go`
  - `func ValidateOrderItems(items []models.OrderItem) error`
  - `func ValidateProductStock(tx *gorm.DB, items []models.OrderItem) error`
  - `func CalculateOrderTotal(items []models.OrderItem) float64`
- 修改 `handlers/order.go:CreateOrder` 调用新函数

**验证**: 运行测试
**提交**: `refactor(order): 提取订单验证逻辑到独立函数`

---

### Step 6: 提取库存扣减逻辑到独立函数
**目标**: 将库存扣减逻辑提取，便于后续迁移到领域层

**改动**:
- 在 `src/utils/order_validation.go` 添加：
  - `func DeductProductStock(tx *gorm.DB, items []models.OrderItem) error`
  - `func RestoreProductStock(tx *gorm.DB, order models.Order) error`
- 修改 `handlers/order.go:CreateOrder` 和 `DeleteOrder` 调用新函数

**验证**: 运行测试
**提交**: `refactor(order): 提取库存扣减逻辑到独立函数`

---

### Step 7: 创建 Order 聚合根（空壳）
**目标**: 创建领域层结构，暂不迁移逻辑

**改动**:
- 创建 `src/domain/order/order.go` - 定义 `Order` 结构体
- 创建 `src/domain/order/order_item.go` - 定义 `OrderItem` 结构体
- 创建 `src/domain/order/repository.go` - 定义仓储接口
- 创建 `src/domain/order/mapper.go` - ToModel/FromModel 转换
- 暂时不添加业务方法，不修改 handler

**验证**: 代码编译通过
**提交**: `feat(domain): 创建 Order 聚合根结构`

---

### Step 8: 为 Order 添加业务方法（内部验证）
**目标**: 在 Order 实体中添加业务方法，暂不调用

**改动**:
- 在 `domain/order/order.go` 添加：
  - `func (o *Order) CalculateTotal() error`
  - `func (o *Order) ValidateItems() error`
- 添加单元测试验证方法正确性
- Handler 仍使用旧逻辑

**验证**: 运行单元测试
**提交**: `feat(domain): 为 Order 添加业务方法`

---

### Step 9: 逐步迁移 CreateOrder 调用新方法
**目标**: 在 Handler 中逐步使用 Order 实体的方法

**改动**:
- 修改 `handlers/order.go:CreateOrder`:
  - 将 `CalculateOrderTotal()` 调用替换为 Order 实体方法
- 保持其他逻辑不变

**验证**: 运行测试
**提交**: `refactor(order): CreateOrder 使用 Order 实体方法计算总价`

---

### Step 10: 创建 Shop 聚合根
**目标**: 复用 Step 7-9 的模式重构 Shop

**改动**:
- 创建 `src/domain/shop/shop.go` - 包含业务方法
- 创建 `src/domain/shop/repository.go`
- 创建 `src/domain/shop/mapper.go`
- 修改 `handlers/shop.go` 逐步调用新方法

**验证**: 运行测试
**提交**: `feat(domain): 创建 Shop 聚合根`

---

### Step 11: 创建 Product 聚合根
**目标**: 复用相同模式重构 Product

**改动**:
- 创建 `src/domain/product/product.go`
- 创建 `src/domain/product/repository.go`
- 创建 `src/domain/product/mapper.go`
- 修改 `handlers/product.go` 调用新方法

**验证**: 运行测试
**提交**: `feat(domain): 创建 Product 聚合根`

---

### Step 12: 清理 models 层业务逻辑 ✅
**目标**: models 只保留 GORM 映射，移除 GORM 钩子

**改动**:
- `domain/shop/shop.go`: 添加密码哈希到 ToModel() 方法
- `handlers/shop.go`: CreateShop 和 UpdateShop 中添加 bcrypt 密码哈希
- `models/shop.go`: 移除 BeforeSave 钩子（密码哈希现在在 handler 中处理）
- `models/shop.go`: 移除 HashPassword 方法（仅在 BeforeSave 中使用）
- 保留 models.Shop 的 wrapper 方法（CheckPassword, IsExpired, RemainingDays）
- models/shop_helpers.go 中的 helper 函数保持不变

**验证**: 运行测试 ✅ (72 passed in 163s)
**提交**: `refactor(shop): Step 12 移除 Shop.BeforeSave 钩子，密码哈希移到 handler/领域层` ✅

---

### Step 13: 创建 Order 领域服务 ✅
**目标**: 将跨实体的订单编排逻辑移到领域服务

**改动**:
- 创建 `src/domain/order/service.go`
  - `Service` 结构体（接受 gorm.DB 依赖）
  - `CreateOrder()`: 创建订单（接受 DTO，返回订单和总价）
  - `processOrderItems()`: 处理订单项（验证库存、保存快照、计算价格、扣减库存）
  - `ValidateOrder()`: 验证订单基础数据
  - `CalculateTotal()`: 计算订单总价
  - `RestoreStock()`: 恢复商品库存
  - DTO 结构（CreateOrderDTO, CreateOrderItemDTO, CreateOrderItemOptionDTO）
- Handler 暂未调用服务（保持原有逻辑不变）

**验证**: 运行测试 ✅ (72 passed in 163s)
**提交**: `feat(domain): Step 13 创建 Order 领域服务` ✅

---

### Step 14: 清理 utils 包中的领域逻辑 ✅
**目标**: 清理 utils 中未使用的函数，标记保留函数为 DEPRECATED

**改动**:
- `utils/order_validation.go`:
  - 删除未使用：ValidateOrderItems, ValidateProductStock, DeductProductStock, CalculateOrderTotal
  - 保留并标记 DEPRECATED：RestoreProductStock, ValidateOrder
- `models/shop_helpers.go`:
  - 删除未使用：GetShopRemainingDays
  - 保留并标记 DEPRECATED：CheckShopPassword, HashShopPassword, IsShopExpired
- `models/shop.go`:
  - 删除未使用的 RemainingDays() 方法

注意：
- 保留的函数标记为 DEPRECATED，说明未来应该使用 domain service
- Handler 层仍在使用这些函数，暂不删除
- 业务逻辑未改变

**验证**: 运行测试 ✅ (72 passed in 163s)
**提交**: `refactor(utils): Step 14 清理 utils 中的领域逻辑` ✅

---

## 四、每步操作模板

```bash
# 1. 修改代码
# 2. 运行测试
cd ../OrderEase-Deploy/test
pytest -v

# 3. 如果测试通过，提交
cd ../../OrderEase-Golang
git add .
git commit -m "描述改动"
git log --oneline -1  # 确认提交

# 4. 如果测试失败，回退
git checkout .
```

---

## 五、关键文件清单

### 需要创建的文件（按步骤）
| Step | 文件 | 用途 |
|------|------|------|
| 1 | `src/utils/shop_helpers.go` | Shop 业务逻辑临时存放 |
| 4 | `src/domain/shared/value_objects/order_status.go` | 订单状态值对象 |
| 5 | `src/utils/order_validation.go` | 订单验证逻辑临时存放 |
| 7 | `src/domain/order/order.go` | Order 实体 |
| 7 | `src/domain/order/order_item.go` | OrderItem 值对象 |
| 7 | `src/domain/order/repository.go` | Order 仓储接口 |
| 7 | `src/domain/order/mapper.go` | Order 转换器 |
| 10 | `src/domain/shop/shop.go` | Shop 实体 |
| 10 | `src/domain/shop/repository.go` | Shop 仓储接口 |
| 10 | `src/domain/shop/mapper.go` | Shop 转换器 |
| 11 | `src/domain/product/product.go` | Product 实体 |
| 11 | `src/domain/product/repository.go` | Product 仓储接口 |
| 11 | `src/domain/product/mapper.go` | Product 转换器 |
| 13 | `src/domain/order/service.go` | Order 领域服务 |

### 需要修改的文件
| Step | 文件 | 改动类型 |
|------|------|----------|
| 1 | `src/models/shop.go` | 调用新函数 |
| 1 | `src/handlers/shop.go` | 调用新函数 |
| 2 | `src/domain/shared/value_objects/password.go` | 统一验证规则 |
| 2 | `src/utils/password.go` | 调用值对象 |
| 3 | `src/models/user.go` | 移除 BeforeSave |
| 3 | `src/handlers/auth.go` | 添加密码加密 |
| 3 | `src/domain/user/service.go` | 添加密码加密 |
| 4 | `src/handlers/order.go` | 使用值对象 |
| 5 | `src/handlers/order.go` | 调用验证函数 |
| 6 | `src/handlers/order.go` | 调用库存函数 |
| 9 | `src/handlers/order.go` | 使用 Order 方法 |
| 10 | `src/handlers/shop.go` | 使用 Shop 实体 |
| 11 | `src/handlers/product.go` | 使用 Product 实体 |
| 12 | `src/models/shop.go` | 移除业务方法 |
| 13 | `src/handlers/order.go` | 调用领域服务 |
| 14 | (删除) `src/utils/*.go` | 清理临时文件 |

---

## 六、测试验证命令

```bash
# 运行所有测试
cd ../OrderEase-Deploy/test
pytest -v

# 运行特定模块测试
pytest admin/test_business_flow.py -v
pytest shop_owner/test_business_flow.py -v

# 运行前端测试
pytest front/test_user_flow.py -v

# 生成测试报告
pytest -v --html=report.html
```

---

## 七、回滚策略

每步都是独立 commit，如出现问题：

```bash
# 查看最近提交
git log --oneline -10

# 回滚到指定提交
git reset --hard <commit-hash>

# 或回滚一步
git reset --hard HEAD~1
```
