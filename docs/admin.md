# 管理员接口文档

## 管理员账户说明
系统只有一个管理员账户，初始账户信息：
- 用户名：admin
- 初始密码：Admin@123456

**请在首次登录后立即修改密码！**

## 密码要求
管理员密码必须满足以下所有条件：
1. 长度至少8位
2. 必须包含数字
3. 必须包含大写字母
4. 必须包含小写字母
5. 必须包含特殊字符（如：@#$%^&*等）

## 基础说明
- 基础路径: `/api/v1/admin`
- 认证方式: Bearer Token
- 请求头: 需要认证的接口必须包含 `Authorization: Bearer <your-token>`

## 接口列表

### 管理员基础接口
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
- **接口**: POST `/change-password`
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

### 商品管理接口
#### 1. 创建商品
- **接口**: POST `/product/create`
- **描述**: 创建新商品（创建后默认为待上架状态）
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
- **响应**:
```json
{
    "message": "商品创建成功",
    "product": {
        "id": 1,
        "name": "商品名称",
        "description": "商品描述",
        "price": 99.99,
        "stock": 100,
        "status": "pending",
        "created_at": "2024-01-20T10:00:00Z",
        "updated_at": "2024-01-20T10:00:00Z"
    }
}
```

#### 2. 更新商品状态
- **接口**: PUT `/product/toggle-status`
- **描述**: 更改商品状态（待上架 -> 已上架 -> 已下架）
- **认证**: 需要
- **请求参数**:
```json
{
    "id": 1,
    "status": "online"  // pending: 待上架, online: 已上架, offline: 已下架
}
```
- **响应**:
```json
{
    "message": "商品状态更新成功",
    "product": {
        "id": 1,
        "status": "online",
        "updated_at": "2024-01-20T10:00:00Z"
    }
}
```

#### 3. 更新商品
- **接口**: PUT `/product/update`
- **描述**: 更新商品信息（已上架商品不允许修改名称和价格）
- **认证**: 需要
- **请求参数**:
```json
{
    "id": 1,
    "name": "更新后的商品名称",     // 已上架商品不可修改
    "description": "更新后的描述",
    "price": 88.88,               // 已上架商品不可修改
    "stock": 50
}
```
- **响应**:
```json
{
    "message": "商品更新成功",
    "product": {
        "id": 1,
        "name": "更新后的商品名称",
        "description": "更新后的描述",
        "price": 88.88,
        "stock": 50,
        "status": "pending",
        "updated_at": "2024-01-20T10:00:00Z"
    }
}
```

**商品状态说明**:
- pending: 待上架（初始状态，可以修改所有信息）
- online: 已上架（不可修改名称和价格，可以修改库存和描述）
- offline: 已下架（商品不可见，不可购买）

**状态流转规则**:
1. 新建商品默认为"待上架"状态
2. "待上架"状态可以转为"已上架"
3. "已上架"状态可以转为"已下架"
4. "已下架"状态不可再次上架，需要创建新商品

#### 4. 获取商品列表
- **接口**: GET `/product/list`
- **描述**: 获取商品列表
- **认证**: 需要
- **查询参数**:
  - page: 页码（可选）
  - limit: 每页数量（可选）

#### 5. 获取商品详情
- **接口**: GET `/product/detail`
- **描述**: 获取单个商品详情
- **认证**: 需要
- **查询参数**:
  - id: 商品ID

#### 6. 删除商品
- **接口**: DELETE `/product/delete`
- **描述**: 删除商品
- **认证**: 需要
- **查询参数**:
  - id: 商品ID

#### 7. 上传商品图片
- **接口**: POST `/product/upload-image`
- **描述**: 上传商品图片
- **认证**: 需要
- **请求体**: multipart/form-data
  - image: 图片文件
  - product_id: 商品ID

#### 8. 获取商品图片
- **接口**: GET `/product/image`
- **描述**: 获取商品图片
- **认证**: 不需要
- **查询参数**:
  - id: 商品ID

### 用户管理接口
#### 1. 创建用户
- **接口**: POST `/admin/user/create`
- **描述**: 创建新用户
- **认证**: 需要
- **请求参数**:
```json
{
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "delivery"
}
```
- **curl示例**:
```bash
curl -X POST 'http://localhost:8080/api/v1/admin/user/create' \
-H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIs...' \
-H 'Content-Type: application/json' \
-d '{
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "delivery"
}'
```
- **响应**:
```json
{
    "message": "用户创建成功",
    "user": {
        "id": 1,
        "name": "张三",
        "phone": "13800138000",
        "address": "北京市朝阳区xxx街道",
        "type": "delivery",
        "created_at": "2024-01-20T10:00:00Z",
        "updated_at": "2024-01-20T10:00:00Z"
    }
}
```

**参数说明**:
- name: 用户姓名
- phone: 手机号码
- address: 收货地址
- type: 用户类型
  - delivery: 邮寄配送
  - pickup: 门店自提

#### 2. 获取用户列表
- **接口**: GET `/user/list`
- **描述**: 获取用户列表
- **认证**: 需要
- **查询参数**:
  - page: 页码（可选）
  - limit: 每页数量（可选）
- **响应示例**:
```json
{
    "total": 100,
    "users": [
        {
            "id": 1,
            "name": "张三",
            "phone": "13800138000",
            "address": "北京市朝阳区xxx街道",
            "type": "delivery",
            "created_at": "2024-01-20T10:00:00Z",
            "updated_at": "2024-01-20T10:00:00Z"
        }
        // ... 更多用户数据
    ]
}
```

#### 3. 获取用户详情
- **接口**: GET `/user/detail`
- **描述**: 获取单个用户详情
- **认证**: 需要
- **查询参数**:
  - id: 用户ID
- **响应示例**:
```json
{
    "user": {
        "id": 1,
        "name": "张三",
        "phone": "13800138000",
        "address": "北京市朝阳区xxx街道",
        "type": "delivery",
        "created_at": "2024-01-20T10:00:00Z",
        "updated_at": "2024-01-20T10:00:00Z"
    }
}
```

#### 4. 更新用户
- **接口**: PUT `/user/update`
- **描述**: 更新用户信息
- **认证**: 需要
- **请求参数**:
```json
{
    "id": 1,
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "pickup"
}
```

#### 5. 删除用户
- **接口**: DELETE `/user/delete`
- **描述**: 删除用户
- **认证**: 需要
- **查询参数**:
  - id: 用户ID

### 订单管理接口
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
4. 2 小时内没有任何请求，token 将自动失效
5. Token 失效后需要重新登录获取新的 token
6. 文件上传接口需要使用 multipart/form-data 格式
7. 分页接口默认每页 20 条数据
8. 请妥善保管 token，不要泄露给他人
