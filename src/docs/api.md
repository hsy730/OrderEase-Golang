# OrderEase API 文档

## 系统概述

OrderEase 是一个基于 DDD（领域驱动设计）架构的模块化单体应用，提供订单管理、商品管理、店铺运营等功能。

### 技术架构

- **架构模式**: DDD 四层架构（Domain → Application → Infrastructure → Interfaces）
- **Web 框架**: Gin 1.9.1
- **ORM**: GORM 1.25.7
- **数据库**: MySQL 8.0+
- **认证方式**: JWT Bearer Token
- **ID 生成**: Snowflake 分布式ID

### 基础信息

- **基础路径**: `/api/order-ease/v1`
- **请求格式**: JSON
- **响应格式**: JSON
- **字符编码**: UTF-8

### 认证方式

所有需要认证的接口必须在请求头中携带 JWT Token：

```
Authorization: Bearer <your-token>
```

## 接口分类

### 1. 公开接口（无需认证）

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 统一登录 | POST | `/login` | 管理员/店主统一登录 |
| Token 刷新 | POST | `/admin/refresh-token` | 刷新管理员 Token |
| Token 刷新 | POST | `/shop/refresh-token` | 刷新店主 Token |
| 临时令牌登录 | POST | `/shop/temp-login` | 使用临时令牌登录 |
| 用户注册 | POST | `/user/register` | 前台用户注册 |
| 用户登录 | POST | `/user/login` | 前台用户登录 |
| 检查用户名 | GET | `/user/check-username` | 检查用户名是否存在 |

### 2. 管理员接口（/admin/*）

| 功能模块 | 文档 | 描述 |
|----------|------|------|
| 数据看板 | [Dashboard](#dashboard) | 获取统计数据、销售趋势、热销商品等 |
| 店铺管理 | [Shop](#shop) | 店铺CRUD、图片上传、临时令牌、订单流转配置 |
| 商品管理 | [Product](#product) | 商品CRUD、状态管理、图片上传、参数类别 |
| 订单管理 | [Order](#order) | 订单CRUD、状态流转、高级搜索 |
| 用户管理 | [User](#user) | 用户CRUD、简单列表 |
| 标签管理 | [Tag](#tag) | 标签CRUD、批量操作、商品绑定 |
| 数据管理 | [Data](#data) | 数据导入导出 |

### 3. 店主接口（/shopOwner/*）

店主接口与管理员接口功能相同，但只能操作自己店铺的数据：

| 功能模块 | 路径前缀 | 描述 |
|----------|----------|------|
| 数据看板 | `/shopOwner/dashboard` | 店铺统计数据 |
| 商品管理 | `/shopOwner/product` | 商品管理（仅当前店铺）|
| 订单管理 | `/shopOwner/order` | 订单管理（含未完成订单列表）|
| 标签管理 | `/shopOwner/tag` | 标签管理 |
| 店铺信息 | `/shopOwner/shop` | 查看/修改当前店铺信息 |
| 用户管理 | `/shopOwner/user` | 用户管理 |

### 4. 前台接口

| 功能模块 | 路径 | 描述 |
|----------|------|------|
| 商品浏览 | `/product/list`, `/product/detail` | 浏览商品列表和详情 |
| 订单管理 | `/order/create`, `/order/user/list` | 创建订单、查看用户订单 |
| 标签浏览 | `/tag/list`, `/tag/bound-products` | 浏览标签和商品 |
| 店铺信息 | `/shop/detail`, `/shop/{shopId}/tags` | 查看店铺信息 |

## 接口详情

### <a name="dashboard"></a>数据看板接口

**路径**: `/admin/dashboard` 或 `/shopOwner/dashboard`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 统计数据 | GET | `/stats` | 获取看板统计数据 |

**查询参数**:
- `shop_id` (可选): 店铺ID（管理员必填，店主自动使用当前店铺）
- `period` (可选): 销售趋势周期，`week` 或 `month`，默认 `week`

**响应包含**:
- 订单统计（今日/昨日订单数、金额）
- 商品统计（总数、待上架、已上架、已下架）
- 用户统计（总用户数、今日新增）
- 订单效率（平均处理时间）
- 销售趋势（最近7天/30天）
- 热销商品TOP5
- 最近订单TOP5

### <a name="shop"></a>店铺管理接口

**路径**: `/admin/shop` 或 `/shopOwner/shop`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 创建店铺 | POST | `/create` | 创建新店铺（仅管理员）|
| 更新店铺 | PUT | `/update` | 更新店铺信息 |
| 获取店铺详情 | GET | `/detail` | 获取单个店铺详情 |
| 获取店铺列表 | GET | `/list` | 分页获取店铺列表（仅管理员）|
| 删除店铺 | DELETE | `/delete` | 删除店铺（仅管理员）|
| 上传图片 | POST | `/upload-image` | 上传店铺图片 |
| 获取图片 | GET | `/image` | 获取店铺图片 |
| 检查名称 | GET | `/check-name` | 检查店铺名称是否存在 |
| 临时令牌 | GET | `/temp-token` | 获取6位临时登录令牌 |
| 更新订单流转 | PUT | `/update-order-status-flow` | 更新订单状态流转配置 |

### <a name="product"></a>商品管理接口

**路径**: `/admin/product` 或 `/shopOwner/product`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 创建商品 | POST | `/create` | 创建新商品（支持参数类别）|
| 更新商品 | PUT | `/update` | 更新商品信息 |
| 获取商品列表 | GET | `/list` | 分页获取商品列表 |
| 获取商品详情 | GET | `/detail` | 获取单个商品详情 |
| 删除商品 | DELETE | `/delete` | 删除商品 |
| 切换状态 | PUT | `/toggle-status` | 切换商品状态 |
| 上传图片 | POST | `/upload-image` | 上传商品图片 |
| 获取图片 | GET | `/image` | 获取商品图片 |

**商品状态**:
- `pending`: 待上架（可修改所有信息）
- `online`: 已上架（不可修改名称和价格）
- `offline`: 已下架（商品不可见）

**参数类别**: 商品支持配置参数选项（如大小、颜色），每个参数可以有多个选项并设置价格调整。

### <a name="order"></a>订单管理接口

**路径**: `/admin/order` 或 `/shopOwner/order`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 创建订单 | POST | `/create` | 创建新订单 |
| 更新订单 | PUT | `/update` | 更新订单信息 |
| 获取订单列表 | GET | `/list` | 分页获取订单列表 |
| 获取订单详情 | GET | `/detail` | 获取单个订单详情 |
| 删除订单 | DELETE | `/delete` | 删除订单 |
| 切换状态 | PUT | `/toggle-status` | 切换订单状态 |
| 获取状态流转 | GET | `/status-flow` | 获取订单状态流转配置 |
| 未完成订单 | GET | `/unfinished-list` | 获取未完成订单列表（仅店主）|
| 高级搜索 | POST | `/advance-search` | 高级搜索订单 |
| SSE 通知 | GET | `/sse` | 订单实时通知 |

**订单状态**: 由店铺配置的状态流转定义，通常包括：待处理、处理中、已完成、已取消等。

### <a name="user"></a>用户管理接口

**路径**: `/admin/user` 或 `/shopOwner/user`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 创建用户 | POST | `/create` | 创建新用户 |
| 更新用户 | PUT | `/update` | 更新用户信息 |
| 获取用户列表 | GET | `/list` | 分页获取用户列表 |
| 获取简单列表 | GET | `/simple-list` | 获取简化用户列表（仅ID和名称）|
| 获取用户详情 | GET | `/detail` | 获取单个用户详情 |
| 删除用户 | DELETE | `/delete` | 删除用户 |

**用户类型**:
- `delivery`: 邮寄配送
- `pickup`: 门店自提

**用户角色**:
- `private`: 私有用户（店主创建）
- `public`: 公开用户（前台注册）

### <a name="tag"></a>标签管理接口

**路径**: `/admin/tag` 或 `/shopOwner/tag`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 创建标签 | POST | `/create` | 创建新标签 |
| 更新标签 | PUT | `/update` | 更新标签信息 |
| 获取标签列表 | GET | `/list` | 获取标签列表 |
| 获取标签详情 | GET | `/detail` | 获取单个标签详情 |
| 删除标签 | DELETE | `/delete` | 删除标签 |
| 批量打标签 | POST | `/batch-tag` | 为多个商品打标签 |
| 批量解绑 | DELETE | `/batch-untag` | 批量解绑商品标签 |
| 批量设置标签 | POST | `/batch-tag-product` | 为单个商品设置多个标签 |
| 获取已绑定标签 | GET | `/bound-tags` | 获取商品已绑定的标签 |
| 获取未绑定标签 | GET | `/unbound-tags` | 获取商品未绑定的标签 |
| 获取标签商品 | GET | `/bound-products` | 获取标签绑定的商品列表 |
| 获取未绑定商品 | GET | `/unbound-products` | 获取标签未绑定的商品列表 |
| 获取未绑定标签列表 | GET | `/unbound-list` | 获取没有绑定商品的标签 |
| 获取上架商品 | GET | `/online-products` | 获取标签关联的已上架商品 |

### <a name="data"></a>数据管理接口

**路径**: `/admin/data`

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 导出数据 | GET | `/export` | 导出系统数据 |
| 导入数据 | POST | `/import` | 导入系统数据 |

## 通用响应格式

### 成功响应

```json
{
  "code": 200,
  "data": { ... },
  "message": "操作成功"
}
```

### 错误响应

```json
{
  "code": 400,
  "error": "错误信息描述"
}
```

## 错误码说明

| HTTP 状态码 | 说明 | 常见场景 |
|------------|------|----------|
| 200 | 成功 | 请求处理成功 |
| 400 | 请求参数错误 | 参数缺失或格式错误 |
| 401 | 未认证 | Token 缺失或无效 |
| 403 | 权限不足 | 无权限访问该资源 |
| 404 | 资源不存在 | 请求的资源不存在 |
| 409 | 资源冲突 | 数据已存在或冲突 |
| 429 | 请求过于频繁 | 触发限流 |
| 500 | 服务器内部错误 | 服务器处理异常 |

## 分页规范

列表接口默认支持以下分页参数：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码，从1开始 |
| pageSize | int | 否 | 10 | 每页数量，最大100 |

**分页响应格式**:

```json
{
  "code": 200,
  "data": {
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "data": [ ... ]
  }
}
```

## 限流规则

| 接口类型 | 限制规则 |
|----------|----------|
| 登录接口 | 每10秒最多1个请求，最大突发3个 |
| 其他接口 | 每秒最多2个请求，最大突发5个 |

超出限制将返回 `429 Too Many Requests`。

## 详细文档

- [认证接口文档](./auth.md) - 登录、Token管理、密码修改
- [店铺接口文档](./api_shop.md) - 店铺管理相关接口
- [商品接口文档](./api_product.md) - 商品管理相关接口
- [订单接口文档](./api_order.md) - 订单管理相关接口
- [用户接口文档](./api_user.md) - 用户管理相关接口
- [标签接口文档](./api_tag.md) - 标签管理相关接口

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.0 | 2025-02-06 | 完善接口文档，添加 DDD 架构说明 |
| 1.0 | 2024-01-01 | 初始版本 |
