# 数据库查询规则

## 1. 查询条件类型匹配规则（重要）

### 1.1 规则说明

**查询条件的类型必须与数据库字段类型严格匹配，否则会导致索引失效，引发全表扫描，严重影响查询性能。**

### 1.2 正确示例

```go
// ✅ 正确做法 - 将字符串转换为 uint64 再查询
func (r *OrderRepository) GetByIDStr(orderID string) (*models.Order, error) {
    // 将字符串 orderID 转换为 uint64，以匹配数据库 bigint 类型
    orderIDUint64, err := strconv.ParseUint(orderID, 10, 64)
    if err != nil {
        return nil, errors.New("无效的订单ID")
    }
    
    var order models.Order
    err = r.DB.First(&order, orderIDUint64).Error
    // ...
}

// ✅ 正确做法 - 高级搜索中的用户ID筛选
func (r *OrderRepository) AdvanceSearch(req AdvanceSearchOrderRequest) (*AdvanceSearchResult, error) {
    query := r.DB.Model(&models.Order{}).Where("shop_id = ?", req.ShopID)
    
    // 添加用户ID筛选
    if req.UserID != "" {
        // 将字符串 user_id 转换为 uint64，以匹配数据库 bigint 类型
        userIDUint64, err := strconv.ParseUint(req.UserID, 10, 64)
        if err != nil {
            return nil, errors.New("无效的用户ID")
        }
        query = query.Where("user_id = ?", userIDUint64)
    }
    // ...
}
```

### 1.3 错误示例

```go
// ❌ 错误做法 - 字符串直接匹配 bigint 字段，导致索引失效
func (r *OrderRepository) AdvanceSearch(req AdvanceSearchOrderRequest) (*AdvanceSearchResult, error) {
    query := r.DB.Model(&models.Order{}).Where("shop_id = ?", req.ShopID)
    
    // 错误：req.UserID 是 string 类型，但 user_id 字段是 bigint 类型
    if req.UserID != "" {
        query = query.Where("user_id = ?", req.UserID)  // 类型不匹配！索引失效！
    }
    // ...
}
```

### 1.4 常见类型映射

| Go 类型 | 数据库类型 | 说明 |
|---------|-----------|------|
| `string` | `varchar/text` | 字符串类型可直接匹配 |
| `uint64` / `snowflake.ID` | `bigint unsigned` | ID 字段通常使用此类型 |
| `int` | `int` | 状态码、数量等 |
| `time.Time` | `datetime/timestamp` | 时间字段 |
| `float64` | `double/decimal` | 价格、金额等 |

### 1.5 检查清单

在编写或审查数据库查询代码时，必须检查：

- [ ] **参数类型**: 查询参数的类型是否与数据库字段类型一致？
- [ ] **字符串转数字**: 如果接收的是字符串 ID，是否使用 `strconv.ParseUint` 转换为 `uint64`？
- [ ] **错误处理**: 类型转换失败时是否有适当的错误处理？
- [ ] **索引字段**: 查询条件中的字段是否有索引？类型不匹配会导致索引失效

### 1.6 性能影响

**类型不匹配的后果：**

1. **索引失效**: MySQL 等数据库在类型不匹配时会放弃使用索引
2. **全表扫描**: 导致查询需要扫描整个表，性能急剧下降
3. **CPU 开销**: 数据库需要隐式类型转换，增加 CPU 负担
4. **锁竞争**: 长时间查询增加锁等待时间

**示例性能对比：**

假设 `orders` 表有 100 万条数据，`user_id` 字段有索引：

| 查询方式 | 执行时间 | 扫描行数 |
|---------|---------|---------|
| `WHERE user_id = 12345` (uint64) | 0.5ms | 1 行 |
| `WHERE user_id = '12345'` (string) | 500ms+ | 100 万行 |

## 2. 批量查询规则

### 2.1 使用 IN 查询替代循环查询

```go
// ✅ 正确做法 - 使用 IN 查询
func (r *ProductRepository) GetProductsByIDs(ids []snowflake.ID, shopID snowflake.ID) ([]models.Product, error) {
    var products []models.Product
    if err := r.DB.Where("id IN (?) AND shop_id = ?", ids, shopID).Find(&products).Error; err != nil {
        return nil, err
    }
    return products, nil
}

// ❌ 错误做法 - 循环查询
func (r *ProductRepository) GetProductsByIDs(ids []snowflake.ID, shopID snowflake.ID) ([]models.Product, error) {
    var products []models.Product
    for _, id := range ids {
        var product models.Product
        r.DB.Where("id = ? AND shop_id = ?", id, shopID).First(&product)  // N+1 查询问题！
        products = append(products, product)
    }
    return products, nil
}
```

## 3. 事务使用规则

### 3.1 事务范围最小化

```go
// ✅ 正确做法 - 事务范围最小化
func (r *OrderRepository) CreateOrder(order *models.Order) error {
    tx := r.DB.Begin()
    
    // 只在必要时使用事务
    if err := tx.Create(order).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 其他操作...
    
    return tx.Commit().Error
}
```

## 4. 预加载规则

### 4.1 避免 N+1 查询

```go
// ✅ 正确做法 - 使用 Preload
func (r *OrderRepository) GetOrdersWithItems(shopID snowflake.ID) ([]models.Order, error) {
    var orders []models.Order
    err := r.DB.Where("shop_id = ?", shopID).
        Preload("Items").
        Preload("Items.Options").
        Find(&orders).Error
    return orders, err
}

// ❌ 错误做法 - 会导致 N+1 查询
func (r *OrderRepository) GetOrdersWithItems(shopID snowflake.ID) ([]models.Order, error) {
    var orders []models.Order
    r.DB.Where("shop_id = ?", shopID).Find(&orders)
    
    for i := range orders {
        r.DB.Where("order_id = ?", orders[i].ID).Find(&orders[i].Items)  // N+1 问题！
    }
    return orders, nil
}
```

## 5. 参考文档

- [GORM 文档](https://gorm.io/docs/)
- [MySQL 索引优化](https://dev.mysql.com/doc/refman/8.0/en/optimization-indexes.html)
- [数据库查询性能优化最佳实践](https://example.com/db-performance)
