# 店铺临时令牌登录设计方案

## 1. 需求分析

* 为每个店铺生成6位数字临时令牌
* 令牌每1小时自动刷新
* 方便用户快速登录，减少登录门槛
* 与现有认证系统兼容

## 2. 设计思路

### 2.1 数据模型设计

创建独立的临时令牌表，关联店铺和系统用户：

```go
type TempToken struct {
    ID        uint64    `gorm:"primarykey" json:"id"`
    ShopID    uint64    `gorm:"index;not null" json:"shop_id"`
    UserID    uint64    `gorm:"index;not null" json:"user_id"` // 关联系统用户
    Token     string    `gorm:"size:6;not null" json:"token"`  // 6位数字令牌
    ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
    CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
```

### 2.2 核心功能设计

#### 2.2.1 令牌生成

* 生成6位随机数字令牌
* 为每个店铺关联一个系统用户
* 令牌有效期为1小时
* 支持自动刷新

#### 2.2.2 认证流程

1. 用户获取店铺的临时令牌
2. 使用临时令牌调用登录接口
3. 系统验证令牌有效性
4. 验证通过后返回JWT令牌
5. 后续请求使用JWT令牌进行认证

#### 2.2.3 定时刷新机制

* 使用cron库设置每小时执行一次的定时任务
* 刷新所有店铺的临时令牌
* 确保令牌的时效性和安全性

### 2.3 安全设计

* 令牌长度为6位，适合快速登录场景
* 每小时自动刷新，降低泄露风险
* 临时令牌仅用于获取JWT令牌，后续请求仍使用JWT认证
* 验证店铺ID和令牌的对应关系

## 3. 实现细节

### 3.1 令牌生成算法

```go
func GenerateTempToken() string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    token := r.Intn(900000) + 100000
    return strconv.Itoa(token)
}
```

### 3.2 核心服务接口

* `GenerateTempToken(shopID uint64)` - 生成店铺临时令牌
* `GetValidTempToken(shopID uint64)` - 获取有效令牌，过期则自动刷新
* `ValidateTempToken(shopID uint64, token string)` - 验证令牌有效性
* `RefreshAllTempTokens()` - 刷新所有店铺的令牌
* `SetupCronJob()` - 设置定时刷新任务

### 3.3 API接口设计

* `GET /shop/:shopId/temp-token` - 获取店铺当前的临时令牌
* `POST /shop/temp-login` - 使用临时令牌登录，返回JWT令牌

### 3.4 中间件集成

修改前端认证中间件，支持临时令牌认证：

* 兼容现有JWT认证
* 增加对店铺ID和临时令牌的验证逻辑

## 4. 系统集成

### 4.1 启动初始化

在系统启动时初始化临时令牌服务：

```go
tempTokenService := services.NewTempTokenService()
tempTokenService.SetupCronJob()
```

### 4.2 数据库迁移

确保临时令牌表在系统启动时自动创建。

## 5. 测试与验证

### 5.1 功能测试

* 生成令牌功能测试
* 令牌有效期测试
* 定时刷新功能测试
* 登录验证测试

### 5.2 性能测试

* 大量店铺时的令牌生成性能
* 并发请求下的令牌验证性能

## 6. 优化方向

1. **动态令牌长度**：根据安全需求可配置令牌长度
2. **个性化过期时间**：支持为不同店铺设置不同的过期时间
3. **令牌使用次数限制**：限制每个令牌的使用次数
4. **令牌刷新策略优化**：支持基于使用情况的动态刷新
5. **安全监控**：增加令牌使用日志和异常检测
6. **多因素认证**：结合临时令牌和其他认证方式

## 7. 兼容性考虑

* 与现有认证系统完全兼容
* 不影响现有API接口的正常使用
* 支持渐进式部署

## 8. 部署说明

* 无需额外的硬件资源
* 仅需在系统启动时初始化定时任务
* 数据库表自动创建

## 9. 监控与维护

* 定期检查定时任务运行状态
* 监控令牌生成和验证的错误日志
* 根据实际使用情况调整令牌策略

## 10. 总结

店铺临时令牌登录方案设计简洁高效，适合快速登录场景。通过6位数字令牌和1小时自动刷新机制，在保证安全性的同时，极大降低了用户登录门槛。该方案与现有认证系统完全兼容，支持渐进式部署，可根据实际需求进行优化和扩展。