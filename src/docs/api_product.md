# 商品相关 API 文档

## 创建产品
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