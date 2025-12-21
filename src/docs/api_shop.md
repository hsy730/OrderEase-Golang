# 店铺相关 API 文档

## 创建店铺
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

## 店铺认证接口

### 获取店铺临时令牌
- **方法**: GET
- **路径**: /shop/temp-token
- **描述**: 获取店铺的临时访问令牌，用于临时登录
- **请求参数**:
  - 查询参数:
    - shop_id (string): 店铺ID，必填
- **响应**: 
  成功时返回临时令牌信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  {
    "code": 200,
    "data": {
      "shop_id": 1234567890,
      "token": "123456",
      "expires_at": "2024-01-01T13:00:00Z"
    }
  }
  ```
  失败:
  ```json
  {
    "code": 400,
    "message": "无效的店铺ID"
  }
  ```

### 临时令牌登录
- **方法**: POST
- **路径**: /shop/temp-login
- **描述**: 使用临时令牌进行店铺登录，成功后返回JWT令牌
- **请求参数**:
  登录数据以JSON格式传递，包含以下字段：
  ```json
  {
    "shop_id": 1234567890,  // 店铺ID，必填
    "token": "123456"       // 临时令牌，6位数字，必填
  }
  ```
- **响应**: 
  成功时返回JWT令牌和用户信息，失败时返回错误信息。示例如下：
  成功:
  ```json
  {
    "code": 200,
    "data": {
      "role": "user",
      "user_info": {
        "id": 1,
        "name": "shop_user",
        "shop_id": 1234567890,
        "shop_name": "店铺A"
      },
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expiredAt": 1641000000
    }
  }
  ```
  失败:
  ```json
  {
    "code": 401,
    "message": "无效的临时令牌"
  }
  ```