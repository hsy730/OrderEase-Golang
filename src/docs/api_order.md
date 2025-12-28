# 订单相关 API 文档

## 创建订单
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

### 翻转订单状态
- **方法**: PUT  
- **路径**: /order/toggle-status
- **描述**: 更新订单状态
- **请求参数**:
  请求数据以JSON格式传递，包含以下字段：
  - id (string): 订单ID，必填
  - shop_id (uint64): 店铺ID，必填
  - next_status (int): 要转换到的状态，必填
- **请求示例**:
  ```json
  {
    "id": "123456789",
    "shop_id": 1,
    "next_status": 2
  }
  ```
- **响应**:
  ```json
  {
    "code": 200,
    "message": "订单状态更新成功",
    "old_status": 1,
    "new_status": 2,
    "order": {
      "id": 123456789,
      "user_id": 987654321,
      "shop_id": 1,
      "total_price": 199.8,
      "status": 2,
      "remark": "",
      "created_at": "2025-12-27T10:03:50.731+0800",
      "updated_at": "2025-12-27T10:03:50.731+0800"
    }
  }
  ```

### 获取订单状态流转配置
- **方法**: GET  
- **路径**: /admin/order/status-flow
- **描述**: 获取店铺的订单状态流转配置，定义了订单允许的状态和状态转换规则
- **请求参数**:
  - shop_id (uint64): 店铺ID，必填
- **响应**:
  ```json
  {
    "shop_id": 1,
    "order_status_flow": {
      "statuses": [
        {
          "value": 0,
          "label": "待处理",
          "type": "pending",
          "isFinal": false,
          "actions": [
            {
              "name": "处理",
              "nextStatus": 1,
              "nextStatusLabel": "处理中"
            },
            {
              "name": "取消",
              "nextStatus": 5,
              "nextStatusLabel": "已取消"
            }
          ]
        },
        {
          "value": 1,
          "label": "处理中",
          "type": "processing",
          "isFinal": false,
          "actions": [
            {
              "name": "完成",
              "nextStatus": 2,
              "nextStatusLabel": "已完成"
            }
          ]
        },
        {
          "value": 2,
          "label": "已完成",
          "type": "completed",
          "isFinal": true,
          "actions": []
        }
      ]
    }
  }
  ```