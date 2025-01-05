# API 接口文档

## 基础说明
- 基础路径: `/api/v1/admin`
- 认证方式: Bearer Token
- 请求头: 需要认证的接口必须包含 `Authorization: Bearer <your-token>`

## 接口列表

### 管理员接口
#### 1. 管理员登录
- **接口**: POST `/login`
- **描述**: 管理员账户登录
- **认证**: 不需要
- **请求参数**:
```json
{
    "username": "admin",
    "password": "your_password"
}
```
- **响应**:
```json
{
    "message": "登录成功",
    "admin": {
        "id": 1,
        "username": "admin"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### 2. 修改管理员密码
- **接口**: POST `/admin/change-password`
- **描述**: 修改管理员密码
- **认证**: 需要
- **请求参数**:
```json
{
    "old_password": "当前密码",
    "new_password": "新密码"
}
```
- **响应**:
```json
{
    "message": "密码修改成功"
}
```

### 商品接口
#### 1. 创建商品
- **接口**: POST `/product/create`
- **描述**: 创建新商品
- **认证**: 需要
- **请求参数**:
```json
{
    "name": "商品名称",
    "description": "商品描述",
    "price": 99.99,
    "stock": 100
}
```

#### 2. 获取商品列表
- **接口**: GET `/product/list`
- **描述**: 获取商品列表
- **认证**: 需要
- **查询参数**:
  - page: 页码（可选）
  - limit: 每页数量（可选）

#### 3. 获取商品详情
- **接口**: GET `/product/detail`
- **描述**: 获取单个商品详情
- **认证**: 需要
- **查询参数**:
  - id: 商品ID

#### 4. 更新商品
- **接口**: PUT `/product/update`
- **描述**: 更新商品信息
- **认证**: 需要
- **请求参数**:
```json
{
    "id": 1,
    "name": "更新后的商品名称",
    "description": "更新后的描述",
    "price": 88.88,
    "stock": 50
}
```

#### 5. 删除商品
- **接口**: DELETE `/product/delete`
- **描述**: 删除商品
- **认证**: 需要
- **查询参数**:
  - id: 商品ID

#### 6. 上传商品图片
- **接口**: POST `/product/upload-image`
- **描述**: 上传商品图片
- **认证**: 需要
- **请求体**: multipart/form-data
  - image: 图片文件
  - product_id: 商品ID

#### 7. 获取商品图片
- **接口**: GET `/product/image`
- **描述**: 获取商品图片
- **认证**: 需要
- **查询参数**:
  - id: 商品ID

### 用户接口
#### 1. 创建用户
- **接口**: POST `/user/create`
- **描述**: 创建新用户
- **认证**: 需要

#### 2. 获取用户列表
- **接口**: GET `/user/list`
- **描述**: 获取用户列表
- **认证**: 需要
- **查询参数**:
  - page: 页码（可选）
  - limit: 每页数量（可选）

#### 3. 获取简单用户列表
- **接口**: GET `/user/simple-list`
- **描述**: 获取简化的用户列表
- **认证**: 需要

#### 4. 获取用户详情
- **接口**: GET `/user/detail`
- **描述**: 获取单个用户详情
- **认证**: 需要
- **查询参数**:
  - id: 用户ID

#### 5. 更新用户
- **接口**: PUT `/user/update`
- **描述**: 更新用户信息
- **认证**: 需要

#### 6. 删除用户
- **接口**: DELETE `/user/delete`
- **描述**: 删除用户
- **认证**: 需要
- **查询参数**:
  - id: 用户ID

### 订单接口
#### 1. 创建订单
- **接口**: POST `/order/create`
- **描述**: 创建新订单
- **认证**: 需要

#### 2. 更新订单
- **接口**: PUT `/order/update`
- **描述**: 更新订单信息
- **认证**: 需要

#### 3. 获取订单列表
- **接口**: GET `/order/list`
- **描述**: 获取订单列表
- **认证**: 需要
- **查询参数**:
  - page: 页码（可选）
  - limit: 每页数量（可选）

#### 4. 获取订单详情
- **接口**: GET `/order/detail`
- **描述**: 获取单个订单详情
- **认证**: 需要
- **查询参数**:
  - id: 订单ID

#### 5. 删除订单
- **接口**: DELETE `/order/delete`
- **描述**: 删除订单
- **认证**: 需要
- **查询参数**:
  - id: 订单ID

#### 6. 切换订单状态
- **接口**: PUT `/order/toggle-status`
- **描述**: 更改订单状态
- **认证**: 需要
- **请求参数**:
```json
{
    "id": 1,
    "status": "processing"
}
```

### 数据管理接口
#### 1. 导出数据
- **接口**: GET `/data/export`
- **描述**: 导出系统数据
- **认证**: 需要

#### 2. 导入数据
- **接口**: POST `/data/import`
- **描述**: 导入系统数据
- **认证**: 需要
- **请求体**: multipart/form-data
  - file: 数据文件

## 错误响应
所有接口的错误响应格式统一为：
```json
{
    "error": "错误信息描述"
}
```

常见错误码：
- 400 Bad Request: 请求参数错误
- 401 Unauthorized: 未认证或认证失败
- 403 Forbidden: 权限不足
- 404 Not Found: 资源不存在
- 500 Internal Server Error: 服务器内部错误

## 注意事项
1. 所有需要认证的接口必须在请求头中携带有效的 token
2. Token 格式：`Authorization: Bearer <your-token>`
3. Token 有效期为 2 小时
4. 文件上传接口需要使用 multipart/form-data 格式
5. 分页接口默认每页 20 条数据
