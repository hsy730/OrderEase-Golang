# OrderEase API 文档

## 产品相关接口

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
  - name (string): 店铺名称
  - description (string): 店铺描述
  - address (string): 店铺地址
  - contact (string): 联系方式
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
      "address": "店铺地址",
      "contact": "联系方式"
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
  - id (string): 店铺ID
  - name (string): 店铺名称
  - description (string): 店铺描述
  - address (string): 店铺地址
  - contact (string): 联系方式
- **响应**: 
  成功时返回更新后的店铺信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  { 
    "code": 200,
    "data": {
      "id": "SHOP123",
      "name": "更新后的店铺名称",
      "description": "更新后的店铺描述",
      "address": "更新后的店铺地址",
      "contact": "更新后的联系方式"
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
      "address": "店铺地址",
      "contact": "联系方式"
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