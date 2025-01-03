# OrderEase API 文档

基础路径: `/api/v1`
域名: gkbdewdnxhwz.sealoshzh.site
端口: 8080

## 访问说明
所有API都应通过以下格式访问：
https://gkbdewdnxhwz.sealoshzh.site/api/v1/[接口路径]

例如，创建商品的完整URL为：
https://gkbdewdnxhwz.sealoshzh.site/api/v1/product/create

## 商品接口

### 1. 创建商品
- **接口**: POST `/product/create`
- **描述**: 新增一条商品信息
- **请求参数 (Body)**:
```json
{
    "name": "商品名称",
    "description": "商品描述",
    "price": 99.9,
    "stock": 100,
    "image_url": "/uploads/products/1_product.jpg"
}
```
- **响应**:
```json
{
    "id": 1,
    "name": "商品名称",
    "description": "商品描述",
    "price": 99.9,
    "stock": 100,
    "image_url": "/uploads/products/1_product.jpg",
    "created_at": "2024-03-14T12:00:00Z",
    "updated_at": "2024-03-14T12:00:00Z"
}
```

### 2. 获取商品列表
- **接口**: GET `/product/list`
- **描述**: 分页查询商品列表
- **请求参数 (Query)**:
  - page: 页码（默认1）
  - pageSize: 每页数量（默认10）
- **响应**:
```json
[
    {
        "id": 1,
        "name": "商品1",
        "description": "描述1",
        "price": 99.9,
        "stock": 100,
        "created_at": "2024-03-14T12:00:00Z",
        "updated_at": "2024-03-14T12:00:00Z"
    }
]
```

### 3. 获取商品详情
- **接口**: GET `/product/detail`
- **描述**: 获取单个商品的详细信息
- **请求参数 (Query)**:
  - id: 商品ID
- **响应**:
```json
{
    "id": 1,
    "name": "商品名称",
    "description": "商品描述",
    "price": 99.9,
    "stock": 100,
    "created_at": "2024-03-14T12:00:00Z",
    "updated_at": "2024-03-14T12:00:00Z"
}
```

### 4. 更新商品
- **接口**: PUT `/product/update`
- **描述**: 更新商品信息
- **请求参数 (Query)**:
  - id: 商品ID
- **请求参数 (Body)**:
```json
{
    "name": "新商品名称",
    "description": "新商品描述",
    "price": 199.9,
    "stock": 200
}
```
- **响应**: 返回更新后的商品信息

### 5. 删除商品
- **接口**: DELETE `/product/delete`
- **描述**: 删除指定商品
- **请求参数 (Query)**:
  - id: 商品ID
- **响应**:
```json
{
    "message": "商品删除成功"
}
```

### 6. 上传商品图片
- **接口**: POST `/product/upload-image`
- **描述**: 上传或更新商品图片
- **请求参数 (Query)**:
  - id: 商品ID
- **请求参数 (Form-data)**:
  - image: 图片文件（支持 jpg、png、gif 格式）
- **响应**:
```json
{
    "message": "图片上传成功",
    "url": "/uploads/products/1_product_1234567890.jpg",
    "type": "create"
}
```
- **说明**:
  - 如果商品没有图片，则为新增操作
  - 如果商品已有图片，则为更新操作，会自动删除旧图片
  - 删除商品时会自动删除关联的图片文件

### 7. 获取商品图片
- **接口**: GET `/product/image`
- **描述**: 获取商品图片
- **请求参数 (Query)**:
  - path: 图片路径（不包含/uploads/前缀）
- **响应**: 图片文件

## 订单接口

### 1. 创建订单
- **接口**: POST `/order/create`
- **描述**: 创建新订单
- **请求参数 (Body)**:
```json
{
    "user_id": 1,
    "total_price": "299.7",
    "status": "pending",
    "remark": "这是订单备注信息",
    "items": [
        {
            "product_id": 1,
            "quantity": 3,
            "price": "99.9"
        }
    ]
}
```
- **响应**:
```json
{
    "id": 1,
    "user_id": 1,
    "total_price": 299.7,
    "status": "pending",
    "remark": "这是订单备注信息",
    "items": [...],
    "created_at": "2024-03-14T12:00:00Z",
    "updated_at": "2024-03-14T12:00:00Z"
}
```

### 2. 更新订单
- **接口**: PUT `/order/update`
- **描述**: 更新订单信息
- **请求参数 (Query)**:
  - id: 订单ID
- **请求参数 (Body)**:
```json
{
    "status": "completed",
    "total_price": "299.7",
    "remark": "更新的订单备注"
}
```
- **响应**: 返回更新后的订单信息

### 3. 获取订单列表
- **接口**: GET `/order/list`
- **描述**: 分页查询订单列表
- **请求参数 (Query)**:
  - page: 页码（默认1）
  - pageSize: 每页数量（默认10）
- **响应**:
```json
[
    {
        "id": 1,
        "user_id": 1,
        "total_price": 299.7,
        "status": "pending",
        "items": [
            {
                "id": 1,
                "order_id": 1,
                "product_id": 1,
                "quantity": 3,
                "price": 99.9
            }
        ],
        "created_at": "2024-03-14T12:00:00Z",
        "updated_at": "2024-03-14T12:00:00Z"
    }
]
```

### 4. 获取订单详情
- **接口**: GET `/order/detail`
- **描述**: 获取单个订单的详细信息
- **请求参数 (Query)**:
  - id: 订单ID
- **响应**: 返回订单详细信息，包含订单项

### 5. 删除订单
- **接口**: DELETE `/order/delete`
- **描述**: 删除指定订单
- **请求参数 (Query)**:
  - id: 订单ID
- **响应**:
```json
{
    "message": "订单删除成功"
}
```

### 6. 翻转订单状态
- **接口**: PUT `/order/toggle-status`
- **描述**: 将订单状态转换为下一个状态
- **请求参数 (Query)**:
  - id: 订单ID
- **请求示例**:
```bash
curl -X PUT "https://gkbdewdnxhwz.sealoshzh.site/api/v1/order/toggle-status?id=1"
```
- **状态转换规则**:
  - pending -> accepted（待处理 -> 已接单）
  - accepted -> shipped（已接单 -> 已发货）
  - shipped -> completed（已发货 -> 已完成）
  - rejected/completed/canceled 状态保持不变
- **响应**:
```json
{
    "message": "订单状态更新成功",
    "old_status": "pending",
    "new_status": "accepted",
    "order": {
        "id": 1,
        "user_id": 1,
        "total_price": 299.7,
        "status": "accepted",
        "remark": "订单备注",
        "items": [...],
        "created_at": "2024-03-14T12:00:00Z",
        "updated_at": "2024-03-14T12:00:00Z"
    }
}
```

## 用户接口

### 1. 创建用户
- **接口**: POST `/api/v1/user/create`
- **描述**: 创建新用户
- **请求参数 (Body)**:
```json
{
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "delivery"  // delivery: 邮寄, pickup: 自提
}
```
- **响应**:
```json
{
    "id": 1,
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "delivery",
    "created_at": "2024-03-14T12:00:00Z",
    "updated_at": "2024-03-14T12:00:00Z"
}
```

### 2. 获取用户列表
- **接口**: GET `/api/v1/user/list`
- **描述**: 分页查询用户列表
- **请求参数 (Query)**:
  - page: 页码（默认1）
  - pageSize: 每页数量（默认10）
- **请求示例**:
```bash
# 基本查询（使用默认分页）
curl -X GET "http://devbox.ns-ojjsi3o6.svc.cluster.local:8080/api/v1/user/list"

# 带分页参数的查询
curl -X GET "http://devbox.ns-ojjsi3o6.svc.cluster.local:8080/api/v1/user/list?page=1&pageSize=20"
```
- **响应**:
```json
{
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "data": [
        {
            "id": 1,
            "name": "张三",
            "phone": "13800138000",
            "address": "北京市朝阳区xxx街道",
            "type": "delivery",
            "created_at": "2024-03-14T12:00:00Z",
            "updated_at": "2024-03-14T12:00:00Z"
        }
    ]
}
```

### 3. 获取用户详情
- **接口**: GET `/api/v1/user/detail`
- **描述**: 获取单个用户的详细信息
- **请求参数 (Query)**:
  - id: 用户ID
- **响应**:
```json
{
    "id": 1,
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "delivery",
    "created_at": "2024-03-14T12:00:00Z",
    "updated_at": "2024-03-14T12:00:00Z"
}
```

### 4. 更新用户信息
- **接口**: PUT `/api/v1/user/update`
- **描述**: 更新用户信息
- **请求参数 (Query)**:
  - id: 用户ID
- **请求参数 (Body)**:
```json
{
    "name": "张三",
    "phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "type": "pickup"
}
```
- **响应**: 返回更新后的用户信息

### 5. 删除用户
- **接口**: DELETE `/api/v1/user/delete`
- **描述**: 删除指定用户
- **请求参数 (Query)**:
  - id: 用户ID
- **响应**:
```json
{
    "message": "用户删除成功"
}
```

### 用户类型说明
- delivery: 邮寄配送
- pickup: 自提

### 注意事项
1. 手机号必须是11位数字，以1开头
2. 用户类型只能是 delivery 或 pickup
3. 删除用户前请确保该用户没有关联的订单
4. 创建和更新用户时，name 和 phone 为必填字段
5. 如果用户类型为 delivery，address 为必填字段

### 6. 获取简单用户列表
- **接口**: GET `/api/v1/user/simple-list`
- **描述**: 获取用户ID和名称列表
- **响应**:
```json
[
    {
        "id": 1,
        "name": "张三"
    },
    {
        "id": 2,
        "name": "李四"
    }
]
```

### 使用示例
```bash
curl -X GET "http://devbox.ns-ojjsi3o6.svc.cluster.local:8080/api/v1/user/simple-list"
```
