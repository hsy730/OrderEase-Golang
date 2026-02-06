# 订单管理接口文档

## 接口列表

| 接口 | 方法 | 路径 | 权限 | 描述 |
|------|------|------|------|------|
| 创建订单 | POST | `/order/create` | 管理员/店主/前台 | 创建新订单 |
| 更新订单 | PUT | `/order/update` | 管理员/店主 | 更新订单信息 |
| 获取订单列表 | GET | `/order/list` | 管理员/店主 | 分页获取订单列表 |
| 获取未完成订单 | GET | `/order/unfinished-list` | 店主 | 获取未完成订单列表 |
| 获取订单详情 | GET | `/order/detail` | 管理员/店主/前台 | 获取单个订单详情 |
| 查询用户订单 | GET | `/order/user/list` | 前台 | 查询用户订单列表 |
| 删除订单 | DELETE | `/order/delete` | 管理员/店主/前台 | 删除订单 |
| 切换状态 | PUT | `/order/toggle-status` | 管理员/店主 | 切换订单状态 |
| 获取状态流转 | GET | `/order/status-flow` | 管理员/店主/前台 | 获取订单状态流转配置 |
| 高级搜索 | POST | `/order/advance-search` | 管理员/店主 | 高级搜索订单 |
| SSE 通知 | GET | `/order/sse` | 管理员/店主 | 订单实时通知 |

**路径前缀**:
- 管理员: `/api/order-ease/v1/admin`
- 店主: `/api/order-ease/v1/shopOwner`
- 前台: `/api/order-ease/v1`

---

## 订单状态说明

订单状态由店铺配置的 `order_status_flow` 决定，每个店铺可以配置不同的状态流转规则。

### 默认状态流转配置

| 状态值 | 状态名称 | 类型 | 是否终态 | 可操作 |
|--------|----------|------|----------|--------|
| 0 | 待处理 | pending | 否 | 处理、取消 |
| 1 | 处理中 | processing | 否 | 完成 |
| 2 | 已完成 | completed | 是 | - |
| 5 | 已取消 | cancelled | 是 | - |

### 状态流转规则

1. 只有非终态订单可以流转
2. 流转必须按照店铺配置的状态流转图进行
3. 每个状态可以配置多个可执行的操作

---

## 1. 创建订单

**接口**: POST `/order/create`

**描述**: 创建新订单，支持选择商品参数选项

**认证**: 需要

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "user_id": 9876543210987654321,
  "items": [
    {
      "product_id": 1234567890123456789,
      "quantity": 2,
      "options": [
        {
          "option_id": 1234567890123456791,
          "category_id": 1234567890123456790
        }
      ]
    },
    {
      "product_id": 1234567890123456800,
      "quantity": 1,
      "options": []
    }
  ],
  "remark": "请尽快处理"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| user_id | uint64 | 是 | 用户ID |
| items | array | 是 | 订单商品列表 |
| remark | string | 否 | 订单备注 |

**OrderItem 字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | uint64 | 是 | 商品ID |
| quantity | int | 是 | 购买数量，必须大于0 |
| options | array | 否 | 选中的参数选项列表 |

**OrderItemOption 字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| option_id | uint64 | 是 | 选项ID |
| category_id | uint64 | 是 | 参数类别ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "order_id": 1111111111111111111,
    "total_price": 89.80,
    "created_at": "2025-02-06T10:30:00Z",
    "status": 0
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "商品库存不足"
}
```

```json
{
  "code": 400,
  "error": "商品不存在或已下架"
}
```

---

## 2. 更新订单

**接口**: PUT `/order/update`

**描述**: 更新订单信息，包括修改商品和数量

**认证**: 需要（管理员/店主）

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 订单ID |

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "items": [
    {
      "product_id": 1234567890123456789,
      "quantity": 3,
      "options": [
        {
          "option_id": 1234567890123456792,
          "category_id": 1234567890123456790
        }
      ]
    }
  ],
  "remark": "更新后的备注"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| items | array | 是 | 新的订单商品列表（会替换原列表）|
| remark | string | 否 | 订单备注 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1111111111111111111,
    "shop_id": 1234567890,
    "user_id": 9876543210987654321,
    "items": [ ... ],
    "total_price": 129.70,
    "status": 0,
    "remark": "更新后的备注",
    "created_at": "2025-02-06T10:30:00Z",
    "updated_at": "2025-02-06T11:00:00Z"
  }
}
```

---

## 3. 获取订单列表

**接口**: GET `/order/list`

**描述**: 分页获取店铺的订单列表

**认证**: 需要（管理员/店主）

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 100,
    "page": 1,
    "pageSize": 10,
    "data": [
      {
        "id": 1111111111111111111,
        "user_id": 9876543210987654321,
        "shop_id": 1234567890,
        "total_price": 89.80,
        "status": 1,
        "remark": "",
        "created_at": "2025-02-06T10:30:00Z",
        "updated_at": "2025-02-06T10:30:00Z"
      }
    ]
  }
}
```

---

## 4. 获取未完成订单列表

**接口**: GET `/order/unfinished-list`

**描述**: 获取店铺所有未完成（非终态）的订单列表

**认证**: 需要（店主）

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量，最大100 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 25,
    "page": 1,
    "pageSize": 10,
    "data": [
      {
        "id": 1111111111111111111,
        "user_id": 9876543210987654321,
        "shop_id": 1234567890,
        "total_price": 89.80,
        "status": 0,
        "remark": "",
        "created_at": "2025-02-06T10:30:00Z",
        "updated_at": "2025-02-06T10:30:00Z"
      }
    ]
  }
}
```

---

## 5. 获取订单详情

**接口**: GET `/order/detail`

**描述**: 获取单个订单的详细信息

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 订单ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1111111111111111111,
    "shop_id": 1234567890,
    "user_id": 9876543210987654321,
    "total_price": 89.80,
    "status": 1,
    "remark": "请尽快处理",
    "created_at": "2025-02-06T10:30:00Z",
    "updated_at": "2025-02-06T10:30:00Z",
    "items": [
      {
        "id": 2222222222222222222,
        "order_id": 1111111111111111111,
        "product_id": 1234567890123456789,
        "product_name": "拿铁咖啡",
        "product_image": "product_xxx.jpg",
        "quantity": 2,
        "price": 28.00,
        "subtotal": 56.00,
        "options": [
          {
            "id": 3333333333333333333,
            "item_id": 2222222222222222222,
            "category_name": "杯型",
            "option_name": "大杯",
            "price_adjustment": 5.00
          }
        ]
      }
    ]
  }
}
```

---

## 6. 查询用户订单列表

**接口**: GET `/order/user/list`

**描述**: 查询指定用户的订单列表（前台接口）

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| user_id | string | 是 | - | 用户ID |
| shop_id | string | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量，最大100 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 10,
    "page": 1,
    "pageSize": 10,
    "data": [
      {
        "id": 1111111111111111111,
        "shop_id": 1234567890,
        "total_price": 89.80,
        "status": 2,
        "created_at": "2025-02-06T10:30:00Z"
      }
    ]
  }
}
```

---

## 7. 删除订单

**接口**: DELETE `/order/delete`

**描述**: 删除订单，未完成订单会恢复商品库存

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 订单ID |
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "message": "订单删除成功"
}
```

**错误响应**:

```json
{
  "code": 404,
  "error": "订单不存在"
}
```

---

## 8. 切换订单状态

**接口**: PUT `/order/toggle-status`

**描述**: 切换订单状态，必须按照店铺配置的状态流转规则

**认证**: 需要（管理员/店主）

**请求参数**:

```json
{
  "id": 1111111111111111111,
  "shop_id": 1234567890,
  "next_status": 1
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 订单ID |
| shop_id | uint64 | 是 | 店铺ID |
| next_status | int | 是 | 目标状态值 |

**成功响应**:

```json
{
  "code": 200,
  "message": "订单状态更新成功",
  "old_status": 0,
  "new_status": 1,
  "order": {
    "id": 1111111111111111111,
    "user_id": 9876543210987654321,
    "shop_id": 1234567890,
    "total_price": 89.80,
    "status": 1,
    "remark": "",
    "created_at": "2025-02-06T10:30:00Z",
    "updated_at": "2025-02-06T11:00:00Z"
  }
}
```

**错误响应**:

```json
{
  "code": 400,
  "error": "当前状态为终态，不允许转换"
}
```

```json
{
  "code": 400,
  "error": "无效的状态转换: 从状态 0 到状态 3 不被允许"
}
```

---

## 9. 获取订单状态流转配置

**接口**: GET `/order/status-flow`

**描述**: 获取店铺的订单状态流转配置

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "shop_id": 1234567890,
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
        },
        {
          "value": 5,
          "label": "已取消",
          "type": "cancelled",
          "isFinal": true,
          "actions": []
        }
      ]
    }
  }
}
```

---

## 10. 高级搜索订单

**接口**: POST `/order/advance-search`

**描述**: 根据多个条件高级搜索订单

**认证**: 需要（管理员/店主）

**请求参数**:

```json
{
  "shop_id": 1234567890,
  "page": 1,
  "page_size": 10,
  "user_id": 9876543210987654321,
  "status": 1,
  "start_time": "2025-02-01T00:00:00Z",
  "end_time": "2025-02-06T23:59:59Z"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | uint64 | 是 | - | 店铺ID |
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量 |
| user_id | uint64 | 否 | - | 按用户ID筛选 |
| status | int | 否 | - | 按状态筛选 |
| start_time | string | 否 | - | 开始时间，ISO8601格式 |
| end_time | string | 否 | - | 结束时间，ISO8601格式 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 50,
    "page": 1,
    "pageSize": 10,
    "data": [ ... ]
  }
}
```

---

## 11. SSE 订单通知

**接口**: GET `/order/sse`

**描述**: Server-Sent Events 实时订单通知，用于新订单提醒

**认证**: 需要（管理员/店主）

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | string | 是 | 店铺ID |

**响应**: SSE 流

```
event: message
data: {"order_id": 1111111111111111111, "type": "new_order", "created_at": "2025-02-06T10:30:00Z"}
```

---

## 错误码汇总

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 无效的订单数据 | 请求参数缺失或格式错误 |
| 400 | 缺少订单ID | 未提供订单ID |
| 400 | 无效的店铺ID | 店铺ID格式错误 |
| 400 | 商品库存不足 | 下单时库存不足 |
| 400 | 商品不存在或已下架 | 商品无法购买 |
| 400 | 当前状态为终态，不允许转换 | 尝试修改终态订单 |
| 400 | 无效的状态转换 | 状态流转不符合规则 |
| 404 | 订单不存在 | 订单ID不存在 |
| 500 | 创建订单失败 | 服务器内部错误 |
