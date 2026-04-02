# DDD 架构规则

## 1. 分层架构原则

### 1.1 领域层 (Domain Layer)
- **职责**: 包含核心业务逻辑和业务规则
- **禁止**: 依赖其他层（基础设施层、应用层）
- **组成**:
  - 实体 (Entity): 包含业务逻辑和状态的领域对象
  - 值对象 (Value Object): 不可变对象，用于描述特征
  - 领域服务 (Domain Service): 处理跨实体的业务逻辑
  - 仓储接口 (Repository Interface): 定义数据访问契约

### 1.2 应用层 (Application Layer)
- **职责**: 协调领域对象完成用例，不包含业务逻辑
- **禁止**: 直接操作数据库（必须通过领域服务）
- **组成**:
  - 应用服务: 编排领域服务完成业务流程
  - DTO: 数据传输对象
  - 事件处理器: 处理领域事件

### 1.3 基础设施层 (Infrastructure Layer)
- **职责**: 提供技术实现（数据库、缓存、消息队列等）
- **组成**:
  - 仓储实现: 实现领域层定义的仓储接口
  - 外部服务客户端: 调用第三方服务
  - 配置管理: 应用配置

### 1.4 接口层 (Interface Layer)
- **职责**: 处理外部请求（HTTP、消息队列等）
- **禁止**: 包含业务逻辑
- **组成**:
  - 控制器/处理器: 接收请求，调用应用服务
  - 路由配置: 定义 API 端点

## 2. 编码规范

### 2.1 Handler 层规范
```go
// ✅ 正确做法
func (h *Handler) UpdateUser(c *gin.Context) {
    // 1. 获取请求参数
    userID := c.Param("id")
    
    // 2. 调用领域服务处理业务逻辑
    if err := h.userDomain.UpdateAvatar(userdomain.UserID(userID), avatarURL); err != nil {
        return errorResponse(c, err)
    }
    
    // 3. 返回响应
    return successResponse(c, result)
}

// ❌ 错误做法 - 直接操作 Repository
func (h *Handler) UpdateUser(c *gin.Context) {
    user, _ := h.userRepo.GetUserByID(userID)
    user.Avatar = avatarURL
    h.userRepo.Update(user)  // 违反 DDD 原则
}
```

### 2.2 领域服务规范
```go
// ✅ 正确做法
func (s *Service) UpdateAvatar(id UserID, avatarURL string) error {
    // 1. 获取领域实体
    user, err := s.repo.GetByID(id)
    if err != nil {
        return err
    }
    
    // 2. 使用实体方法更新状态
    user.SetAvatar(avatarURL)
    
    // 3. 持久化
    return s.repo.Update(user)
}
```

### 2.3 实体规范
```go
// ✅ 正确做法 - 充血模型
type User struct {
    id     UserID
    name   string
    avatar string
}

func (u *User) SetAvatar(avatar string) {
    u.avatar = avatar
}

func (u *User) Avatar() string {
    return u.avatar
}
```

## 3. 依赖关系

```
接口层 → 应用层 → 领域层 ← 基础设施层
```

- **依赖方向**: 只能向内依赖，不能反向依赖
- **依赖注入**: 通过构造函数注入依赖
- **接口隔离**: 领域层定义接口，基础设施层实现

## 4. 新增功能的 DDD 合规检查清单

在实现新功能时，必须检查以下项目：

- [ ] **领域实体**: 是否在领域层定义了实体和值对象？
- [ ] **领域服务**: 业务逻辑是否封装在领域服务中？
- [ ] **仓储接口**: 是否在领域层定义了仓储接口？
- [ ] **Handler 层**: 是否只处理 HTTP 请求，不直接操作 Repository？
- [ ] **依赖方向**: 是否遵循从内到外的依赖方向？
- [ ] **单元测试**: 是否为领域服务编写了单元测试？

## 5. 常见违规场景

### 5.1 直接操作 Repository
```go
// ❌ 违规 - Handler 直接操作 Repository
func (h *Handler) UploadAvatar(c *gin.Context) {
    user, _ := h.userRepo.GetUserByID(userID)
    user.Avatar = avatarURL
    h.userRepo.Update(user)
}
```

### 5.2 贫血模型
```go
// ❌ 违规 - 贫血模型，只有 getter/setter，没有业务逻辑
type User struct {
    Avatar string
}
```

### 5.3 跨层依赖
```go
// ❌ 违规 - 领域层依赖基础设施层
package domain

import "orderease/infrastructure/database"  // 违规！
```

## 6. 修复示例

### 问题: 头像上传直接操作 Repository

**修复步骤**:
1. 在领域实体 `User` 中添加 `avatar` 字段和 `SetAvatar` 方法
2. 在领域服务 `UserService` 中添加 `UpdateAvatar` 方法
3. 修改 Handler，调用领域服务而非直接操作 Repository
4. 添加单元测试验证领域服务行为

**参考实现**:
- 领域实体: `src/contexts/ordercontext/domain/user/user.go`
- 领域服务: `src/contexts/ordercontext/domain/user/service.go`
- Handler: `src/contexts/ordercontext/application/handlers/user.go`
- 单元测试: `src/contexts/ordercontext/domain/user/service_test.go`

## 7. 参考文档

- [DDD 架构改进博客](file:///d:/local_code_repo/OrderEase/OrderEase-Golang/DDD_IMPROVEMENT_BLOG.md)
- [DDD 重构计划](file:///d:/local_code_repo/OrderEase/OrderEase-Golang/DDD_REFACTOR_PLAN.md)
- [DDD 架构说明](file:///d:/local_code_repo/OrderEase/OrderEase-Golang/DDD_ARCHITECTURE.md)
