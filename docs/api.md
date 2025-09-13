# OrderEase API 文档

## 产品相关接口

### 创建产品
- **方法**: POST
- **路径**: /product/create
- **描述**: 创建新的产品，支持设置产品基本信息和参数类别选项
- **请求参数**:
  - 产品数据以JSON格式传递，包含以下字段：
    ```json
    {
      "shop_id": 1234567890,            // 店铺ID，必填
      "name": "产品名称",                // 产品名称，必填
      "description": "产品描述",          // 产品描述
      "price": 99.9,                    // 产品基础价格，必填
      "stock": 100,                     // 产品库存，必填
      "image_url": "http://example.com/image.jpg", // 产品图片URL
      "option_categories": [            // 产品参数类别列表，可选
        {
          "name": "大小",               // 参数类别名称，必填
          "is_required": true,          // 是否必填，默认false
          "is_multiple": false,         // 是否允许多选，默认false
          "display_order": 1,           // 显示顺序，默认0
          "options": [                  // 参数选项列表，必填
            {
              "name": "小杯",           // 选项名称，必填
              "price_adjustment": 0,    // 价格调整值，可以是正或负
              "display_order": 1        // 显示顺序，默认0
            },
            {
              "name": "中杯",
              "price_adjustment": 5,
              "display_order": 2
            }
          ]
        }
      ]
    }
    ```
- **响应**:
  - 成功: 返回创建的产品信息
    ```json
    {
      "code": 200,
      "data": {
        "id": "1234567890123456789",
        "shop_id": 1234567890,
        "name": "产品名称",
        "description": "产品描述",
        "price": 99.9,
        "stock": 100,
        "image_url": "http://example.com/image.jpg",
        "status": "pending",
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-01T12:00:00Z",
        "option_categories": [
          // 包含创建的参数类别信息
        ]
      }
    }
    ```
  - 失败:
    ```json
    {
      "code": 400,
      "message": "无效的商品数据: [具体错误信息]"
    }
    ```

### 更新产品
- **方法**: PUT
- **路径**: /product/update?id={product_id}&shop_id={shop_id}
- **描述**: 更新产品信息，支持修改产品基本信息和参数类别选项
- **请求参数**:
  - 查询参数:
    - id (string): 产品ID，必填
    - shop_id (string): 店铺ID，必填
  - 产品数据以JSON格式传递，包含以下字段：
    ```json
    {
      "name": "更新后的产品名称",         // 产品名称
      "description": "更新后的产品描述",   // 产品描述
      "price": 109.9,                   // 产品基础价格
      "stock": 150,                     // 产品库存
      "image_url": "http://example.com/new_image.jpg", // 产品图片URL
      "option_categories": [            // 产品参数类别列表，如果提供则会替换所有现有参数类别
        {
          "name": "大小",               // 参数类别名称，必填
          "is_required": true,          // 是否必填
          "is_multiple": false,         // 是否允许多选
          "display_order": 1,           // 显示顺序
          "options": [                  // 参数选项列表，必填
            {
              "name": "小杯",           // 选项名称，必填
              "price_adjustment": 0,    // 价格调整值
              "display_order": 1        // 显示顺序
            },
            {
              "name": "大杯",           // 新增选项
              "price_adjustment": 10,
              "display_order": 2
            }
          ]
        }
      ]
    }
    ```
- **响应**:
  - 成功: 返回更新后的产品信息
    ```json
    {
      "code": 200,
      "data": {
        "id": "1234567890123456789",
        "shop_id": 1234567890,
        "name": "更新后的产品名称",
        "description": "更新后的产品描述",
        "price": 109.9,
        "stock": 150,
        "image_url": "http://example.com/new_image.jpg",
        "status": "pending",
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z",
        "option_categories": [
          // 包含更新后的参数类别信息
        ]
      }
    }
    ```
  - 失败:
    ```json
    {
      "code": 400,
      "message": "无效的更新数据: [具体错误信息]"
    }
    ```

### 获取产品图片
- **方法**: GET
- **路径**: /product/image
- **描述**: 获取指定产品的图片
- **请求参数**:
  - product_id (string): 产品ID
- **响应**:
  - 成功: 返回图片二进制数据
  - 失败: 
    ```json
    {
      "code": 404,
      "message": "Product not found"
    }
    ```

### 获取产品列表  
- **方法**: GET
- **路径**: /product/list
- **描述**: 获取产品列表
- **请求参数**:
  - page (int): 页码，默认1
  - page_size (int): 每页数量，默认10
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "products": [
        {
          "id": "123",
          "name": "产品A",
          "price": 99.9,
          "image_url": "http://example.com/image.jpg"
        }
      ],
      "total": 100
    }
  }
  ```

### 获取产品详情
- **方法**: GET  
- **路径**: /product/detail
- **描述**: 获取单个产品详情
- **请求参数**:
  - product_id (string): 产品ID
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "id": "123",
      "name": "产品A",
      "description": "产品描述",
      "price": 99.9,
      "stock": 100,
      "tags": ["tag1", "tag2"]
    }
  }
  ```

## 订单相关接口

### 创建订单
- **方法**: POST
- **路径**: /order/create
- **描述**: 创建新订单
- **请求参数**:
  订单数据以JSON格式传递，具体字段参考 `models.Order` 结构体。
- **响应**:
  成功时返回创建的订单信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  {
    "code": 200,
    "message": "Order created successfully"
  }
  ```

### 获取订单列表
- **方法**: GET
- **路径**: /order/list  
- **描述**: 获取用户订单列表
- **请求参数**:
  - page (int): 页码，默认1
  - page_size (int): 每页数量，默认10
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "orders": [
        {
          "id": "ORDER123",
          "product_name": "产品A",
          "total_price": 199.8,
          "status": "completed"
        }
      ],
      "total": 10
    }
  }
  ```

### 获取订单详情
- **方法**: GET
- **路径**: /order/detail
- **描述**: 获取单个订单详情
- **请求参数**:
  - order_id (string): 订单ID
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "id": "ORDER123",
      "products": [
        {
          "id": "123",
          "name": "产品A",
          "quantity": 2,
          "price": 99.9
        }
      ],
      "total_price": 199.8,
      "status": "completed"
    }
  }
  ```

### 查询用户订单列表
- **方法**: GET
- **路径**: /order/user/list
- **描述**: 查询用户订单列表
- **请求参数**:
  - user_id (string): 用户ID
  - page (int): 页码，默认1
  - page_size (int): 每页数量，默认10
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "orders": [
        {
          "id": "ORDER123",
          "product_name": "产品A",
          "total_price": 199.8,
          "status": "completed"
        }
      ],
      "total": 10
    }
  }
  ```

### 删除订单
- **方法**: DELETE  
- **路径**: /order/delete
- **描述**: 删除指定订单
- **请求参数**:
  - order_id (string): 订单ID
- **响应**:
  ```json
  {
    "code": 200,
    "message": "Order deleted successfully"
  }
  ```

## 用户管理接口

### 删除用户
`DELETE /user/delete`

**请求参数:**
```json
{
  "user_id": "雪flake ID"
}
```

**权限要求:**
- 需要店铺管理员权限

**成功响应:**
```json
{
  "code": 200,
  "msg": "用户已删除",
  "data": {
    "deleted_at": "2024-03-20T15:04:05Z"
  }
}
```

### 创建用户
- **方法**: POST
- **路径**: /user/create
- **请求参数**:
  - username (string): 用户名
  - password (string): 密码
  - role (string): 用户角色
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "id": "USER123",
      "username": "testuser",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 获取用户列表
- **方法**: GET
- **路径**: /user/list
- **请求参数**:
  - page (int): 页码
  - page_size (int): 每页数量
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "users": [
        {
          "id": "USER123",
          "username": "testuser",
          "role": "admin"
        }
      ],
      "total": 1
    }
  }
  ```

### 其他接口...

## 标签相关接口

### 获取商品标签列表
- **方法**: GET
- **路径**: /tag/list
- **描述**: 获取所有标签
- **响应**:
  ```json
  {
    "code": 200,
    "data": [
      {
        "id": 1,
        "name": "标签1"
      },
      {
        "id": 2,
        "name": "标签2"
      }
    ]
  }
  ```

### 获取标签详情  
- **方法**: GET
- **路径**: /tag/detail
- **描述**: 获取单个标签详情
- **请求参数**:
  - tag_id (string): 标签ID
- **响应**:
  ```json
  {
    "code": 200,
    "data": {
      "id": "TAG1",
      "name": "标签1",
      "description": "标签描述",
      "product_count": 10
    }
  }
  ```



## 店铺相关接口

### 创建店铺
- **方法**: POST
- **路径**: /shop/create
- **描述**: 创建新的店铺
- **请求参数**:
  店铺数据以JSON格式传递，具体字段参考 `models.Shop` 结构体，包含以下字段：
  - owner_username (string): 店主登录用户名
  - owner_password (string): 店主登录密码
  - name (string): 店铺名称
  - contact_phone (string): 联系电话
  - contact_email (string): 联系邮箱
  - description (string): 店铺描述
  - valid_until (string): 有效期截止时间（ISO8601格式）
- **响应**: 
  成功时返回创建的店铺信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  { 
    "code": 200,
    "data": {
      "id": "SHOP123",
      "name": "店铺A",
      "description": "店铺描述",
      "owner_username": "testuser",
      "owner_password": "123",
      "contact_phone": "13800138000",
      "address": "address",
      "contact_email": "shop@example.com",
      "valid_until": "2025-12-31T23:59:59Z",
      "settings": {}
    }
  }
  ```
  失败:
  ```json
  { 
    "code": 404,
    "message": "Shop creation failed"
  }
  ```

### 更新店铺
- **方法**: PUT
- **路径**: /shop/update
- **描述**: 更新店铺信息
- **请求参数**:
  店铺数据以JSON格式传递，具体字段参考 `models.Shop` 结构体，包含以下字段：
  - owner_username (string): 店主登录用户名
  - owner_password (string): 店主登录密码（可选修改）
  - name (string): 店铺名称
  - contact_phone (string): 联系电话
  - contact_email (string): 联系邮箱
  - description (string): 店铺描述
  - valid_until (string): 新的有效期截止时间（ISO8601格式）
- **响应**: 
  成功时返回更新后的店铺信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  { 
    "code": 200,
    "data": {
      "id": "SHOP123",
      "name": "店铺A",
      "description": "店铺描述",
      "owner_username": "testuser",
      "owner_password": "123",
      "contact_phone": "13800138000",
      "address": "address",
      "contact_email": "shop@example.com",
      "valid_until": "2025-12-31T23:59:59Z",
      "settings": {}
    }
  }
  ```
  失败:
  ```json
  { 
    "code": 404,
    "message": "Shop update failed"
  }
  ```

### 获取店铺信息
- **方法**: GET
- **路径**: /shop/detail
- **描述**: 获取单个店铺的详细信息
- **请求参数**:
  - shop_id (string): 店铺ID
- **响应**: 
  成功时返回店铺详细信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  { 
    "code": 200,
    "data": {
      "id": "SHOP123",
      "name": "店铺A",
      "description": "店铺描述",
      "contact_phone": "13800138000",
      "contact_email": "shop@example.com",
      "valid_until": "2025-12-31T23:59:59Z"
      "tags": []
    }
  }
  ```
  失败:
  ```json
  { 
    "code": 404,
    "message": "Shop not found"
  }
  ```

### 获取店铺列表
- **方法**: GET
- **路径**: /shop/list
- **描述**: 获取店铺列表
- **请求参数**:
  - page (int): 页码，默认1
  - page_size (int): 每页数量，默认10
- **响应**: 
  成功时返回店铺列表信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  { 
    "code": 200,
    "data": {
      "shops": [
        { 
          "id": "SHOP123",
          "name": "店铺A",
          "description": "店铺描述",
          "address": "店铺地址",
          "contact": "联系方式"
        }
      ],
      "total": 10
    }
  }
  ```
  失败:
  ```json
  { 
    "code": 404,
    "message": "Shops not found"
  }
  ```
// 商品图片示例
"image_url": "product_1966495020227760128_1757684275.jpg"

// 订单项示例
"image_url": "product_123456789_987654321.jpg"