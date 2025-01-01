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
- **请求体**:
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
- **参数**:
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
- **参数**:
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
- **参数**:
  - id: 商品ID
- **请求体**:
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
- **参数**:
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
- **参数**:
  - id: 商品ID
- **请求体**: multipart/form-data
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
- **参数**:
  - path: 图片路径（不包含/uploads/前缀）
- **响应**: 图片文件

## 订单接口

### 1. 创建订单
- **接口**: POST `/order/create`
- **描述**: 创建新订单
- **请求体**:
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
- **参数**:
  - id: 订单ID
- **请求体**:
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
- **参数**:
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
- **参数**:
  - id: 订单ID
- **响应**: 返回订单详细信息，包含订单项

### 5. 删除订单
- **接口**: DELETE `/order/delete`
- **描述**: 删除指定订单
- **参数**:
  - id: 订单ID
- **响应**:
```json
{
    "message": "订单删除成功"
}
```

## 错误响应
所有接口在发生错误时会返回相应的HTTP状态码和错误信息：

- 400 Bad Request: 请求参数错误
- 404 Not Found: 资源未找到
- 500 Internal Server Error: 服务器内部错误

错误响应格式：
```json
{
    "error": "错误信息描述"
}
```

