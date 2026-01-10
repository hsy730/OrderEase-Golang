# Wire 依赖注入框架使用说明

## 概述

本项目已引入 Google Wire 依赖注入框架，用于自动管理服务和仓储的依赖关系。

## 文件结构

```
src/application/services/
├── wire.go          # Wire 配置文件（wireinject 构建标签）
├── wire_gen.go      # Wire 自动生成的代码（不要手动编辑）
├── container.go     # ServiceContainer 结构体定义
├── order_service.go
├── product_service.go
├── shop_service.go
└── user_service.go
```

## 使用方法

### 1. 初始化服务容器

```go
import "orderease/application/services"

// 使用 Wire 生成的初始化函数
container, err := services.InitializeServiceContainer(db)
if err != nil {
    log.Fatal("初始化服务容器失败:", err)
}

// 访问服务
orderService := container.OrderService
productService := container.ProductService
```

### 2. 添加新服务时的步骤

#### Step 1: 创建服务

```go
// application/services/my_service.go
package services

type MyService struct {
    myRepo domain.MyRepository
}

func NewMyService(myRepo domain.MyRepository) *MyService {
    return &MyService{myRepo: myRepo}
}
```

#### Step 2: 在 wire.go 中注册

```go
// wire.go
func InitializeServiceContainer(db *gorm.DB) (*ServiceContainer, error) {
    wire.Build(
        // ... 现有依赖

        // 添加新服务
        NewMyService,

        // ... 其他
    )
    return &ServiceContainer{}, nil
}
```

#### Step 3: 更新 ServiceContainer

```go
// container.go
type ServiceContainer struct {
    // ... 现有服务
    MyService *MyService  // 添加新服务
}
```

#### Step 4: 重新生成代码

```bash
cd src/application/services
go generate
# 或手动运行 wire
wire generate
```

### 3. 重新生成 Wire 代码

当修改了 `wire.go` 文件或添加了新服务后，需要重新生成代码：

```bash
# 方式1: 使用 go generate
cd src/application/services
go generate

# 方式2: 使用 wire 命令
cd src/application/services
wire generate
```

## 优势

1. **编译时依赖检查**：Wire 在编译时生成代码，可以提前发现依赖问题
2. **类型安全**：所有依赖都是类型安全的，编译器会检查
3. **减少样板代码**：自动生成依赖注入代码，减少手动编写
4. **易于维护**：依赖关系集中在 `wire.go` 中，易于查看和修改

## 注意事项

1. **不要手动编辑 `wire_gen.go`**：这是 Wire 自动生成的文件，每次运行 `wire generate` 都会被覆盖
2. **保持 `wire.go` 的 `//go:build wireinject` 标签**：这是 Wire 识别配置文件的标识
3. **添加新依赖后记得重新生成**：否则编译会失败

## 故障排查

### 编译错误：undefined: services.InitializeServiceContainer

**原因**：`wire_gen.go` 没有正确生成或被忽略

**解决**：
1. 检查 `wire.go` 文件是否有 `//go:build wireinject` 标签
2. 运行 `cd src/application/services && wire generate`
3. 确保 `wire_gen.go` 文件存在

### Wire 生成失败

**原因**：依赖关系配置错误

**解决**：
1. 检查 `wire.Build()` 中的所有 provider 函数是否存在
2. 确保导入路径正确
3. 查看错误信息，通常是某个依赖无法满足
