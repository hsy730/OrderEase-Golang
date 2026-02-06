# 店铺管理接口文档

## 接口列表

| 接口 | 方法 | 路径 | 权限 | 描述 |
|------|------|------|------|------|
| 数据看板 | GET | `/dashboard/stats` | 管理员/店主 | 获取店铺统计数据 |
| 创建店铺 | POST | `/shop/create` | 管理员 | 创建新店铺 |
| 更新店铺 | PUT | `/shop/update` | 管理员/店主 | 更新店铺信息 |
| 获取店铺详情 | GET | `/shop/detail` | 管理员/店主 | 获取单个店铺详情 |
| 获取店铺列表 | GET | `/shop/list` | 管理员 | 分页获取店铺列表 |
| 删除店铺 | DELETE | `/shop/delete` | 管理员 | 删除店铺 |
| 上传图片 | POST | `/shop/upload-image` | 管理员/店主 | 上传店铺图片 |
| 获取图片 | GET | `/shop/image` | 公开 | 获取店铺图片 |
| 检查名称 | GET | `/shop/check-name` | 管理员 | 检查店铺名称是否存在 |
| 临时令牌 | GET | `/shop/temp-token` | 管理员/店主 | 获取临时登录令牌 |
| 更新订单流转 | PUT | `/shop/update-order-status-flow` | 管理员/店主 | 更新订单状态流转配置 |

**路径前缀**:
- 管理员: `/api/order-ease/v1/admin`
- 店主: `/api/order-ease/v1/shopOwner`

---

## 1. 数据看板

**接口**: GET `/dashboard/stats`

**描述**: 获取店铺的统计数据，包括订单、商品、用户、销售趋势等

**认证**: 需要

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| shop_id | uint64 | 条件 | - | 店铺ID（管理员必填，店主自动使用当前店铺）|
| period | string | 否 | week | 销售趋势周期: week(7天) / month(30天) |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "orderStats": {
      "todayOrders": 15,
      "todayAmount": 2580.50,
      "yesterdayOrders": 12,
      "yesterdayAmount": 1890.00,
      "orderGrowth": 25.0
    },
    "productStats": {
      "total": 100,
      "pending": 20,
      "online": 70,
      "offline": 10
    },
    "userStats": {
      "total": 500,
      "todayNew": 5
    },
    "orderEfficiency": {
      "avgProcessTime": 15.5
    },
    "salesTrend": {
      "dates": ["2025-01-30", "2025-01-31", "2025-02-01", "2025-02-02", "2025-02-03", "2025-02-04", "2025-02-05"],
      "amounts": [1200.0, 1500.0, 1800.0, 2100.0, 1900.0, 2200.0, 2580.5]
    },
    "hotProducts": [
      {
        "id": 1234567890123456789,
        "name": "热销商品1",
        "sales": 50,
        "amount": 5000.0
      }
    ],
    "recentOrders": [
      {
        "id": 9876543210987654321,
        "userName": "张三",
        "totalPrice": 199.8,
        "status": 1,
        "createdAt": "2025-02-06T10:00:00Z"
      }
    ]
  }
}
```

---

## 2. 创建店铺

**接口**: POST `/shop/create`

**描述**: 创建新店铺（仅管理员）

**认证**: 需要（管理员）

**请求参数**:

```json
{
  "name": "店铺A",
  "owner_username": "shopowner1",
  "owner_password": "Pass@123456",
  "contact_phone": "13800138000",
  "contact_email": "shop@example.com",
  "description": "店铺描述信息",
  "address": "北京市朝阳区xxx街道",
  "valid_until": "2025-12-31T23:59:59Z",
  "settings": {},
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
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 店铺名称 |
| owner_username | string | 是 | 店主登录用户名，需唯一 |
| owner_password | string | 是 | 店主登录密码，需满足密码强度要求 |
| contact_phone | string | 否 | 联系电话 |
| contact_email | string | 否 | 联系邮箱 |
| description | string | 否 | 店铺描述 |
| address | string | 否 | 店铺地址 |
| valid_until | string | 否 | 有效期截止时间，ISO8601格式 |
| settings | object | 否 | 店铺设置JSON |
| order_status_flow | object | 否 | 订单状态流转配置 |

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1234567890,
    "name": "店铺A",
    "description": "店铺描述信息",
    "owner_username": "shopowner1",
    "contact_phone": "13800138000",
    "address": "北京市朝阳区xxx街道",
    "contact_email": "shop@example.com",
    "valid_until": "2025-12-31T23:59:59Z",
    "settings": {},
    "order_status_flow": { ... }
  }
}
```

**错误响应**:

```json
{
  "code": 409,
  "error": "店主用户名已存在"
}
```

---

## 3. 更新店铺

**接口**: PUT `/shop/update`

**描述**: 更新店铺信息

**认证**: 需要

**请求参数**:

```json
{
  "id": 1234567890,
  "owner_username": "shopowner1",
  "owner_password": "NewPass@789",
  "name": "店铺A（新名称）",
  "contact_phone": "13900139000",
  "contact_email": "newemail@example.com",
  "description": "更新后的描述",
  "address": "新地址",
  "valid_until": "2026-12-31T23:59:59Z",
  "settings": {},
  "order_status_flow": { ... }
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 店铺ID |
| owner_username | string | 是 | 店主登录用户名 |
| owner_password | string | 否 | 新密码（不填则不修改）|
| name | string | 否 | 店铺名称 |
| contact_phone | string | 否 | 联系电话 |
| contact_email | string | 否 | 联系邮箱 |
| description | string | 否 | 店铺描述 |
| address | string | 否 | 店铺地址 |
| valid_until | string | 否 | 有效期（仅管理员可修改）|
| settings | object | 否 | 店铺设置 |
| order_status_flow | object | 否 | 订单状态流转配置 |

**注意事项**:
- 店主只能修改自己店铺的信息
- 只有管理员可以修改 `valid_until` 字段

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "id": 1234567890,
    "name": "店铺A（新名称）",
    "description": "更新后的描述",
    "owner_username": "shopowner1",
    "contact_phone": "13900139000",
    "address": "新地址",
    "contact_email": "newemail@example.com",
    "valid_until": "2026-12-31T23:59:59Z",
    "settings": {},
    "order_status_flow": { ... }
  }
}
```

---

## 4. 获取店铺详情

**接口**: GET `/shop/detail`

**描述**: 获取单个店铺的详细信息

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
    "id": 1234567890,
    "name": "店铺A",
    "owner_username": "shopowner1",
    "contact_phone": "13800138000",
    "contact_email": "shop@example.com",
    "address": "北京市朝阳区xxx街道",
    "description": "店铺描述信息",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-02-01T00:00:00Z",
    "valid_until": "2025-12-31T23:59:59Z",
    "settings": {},
    "tags": [],
    "image_url": "shop_1234567890_xxx.jpg",
    "order_status_flow": { ... }
  }
}
```

---

## 5. 获取店铺列表

**接口**: GET `/shop/list`

**描述**: 分页获取店铺列表（仅管理员）

**认证**: 需要（管理员）

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 10 | 每页数量 |
| search | string | 否 | - | 搜索关键词（店铺名称）|

**成功响应**:

```json
{
  "code": 200,
  "data": {
    "total": 100,
    "page": 1,
    "data": [
      {
        "id": 1234567890,
        "name": "店铺A",
        "owner_username": "shopowner1",
        "contact_phone": "13800138000",
        "valid_until": "2025-12-31T23:59:59Z",
        "tags_count": 5
      }
    ]
  }
}
```

---

## 6. 删除店铺

**接口**: DELETE `/shop/delete`

**描述**: 删除店铺（仅管理员）

**认证**: 需要（管理员）

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | string | 是 | 店铺ID |

**成功响应**:

```json
{
  "code": 200,
  "message": "店铺删除成功"
}
```

**错误响应**:

```json
{
  "code": 404,
  "error": "店铺不存在"
}
```

```json
{
  "code": 409,
  "error": "存在关联商品，无法删除店铺"
}
```

```json
{
  "code": 409,
  "error": "存在关联订单，无法删除店铺"
}
```

---

## 7. 上传店铺图片

**接口**: POST `/shop/upload-image`

**描述**: 上传店铺图片

**认证**: 需要

**请求类型**: multipart/form-data

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | 是 | 店铺ID |

**请求体**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | file | 是 | 图片文件（jpg/png，最大2MB）|

**成功响应**:

```json
{
  "message": "图片上传成功",
  "url": "shop_1234567890_abc123.jpg",
  "type": "upload"
}
```

---

## 8. 获取店铺图片

**接口**: GET `/shop/image`

**描述**: 获取店铺图片

**认证**: 公开（无需认证）

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 图片文件名 |

**响应**: 图片二进制数据

---

## 9. 检查店铺名称

**接口**: GET `/shop/check-name`

**描述**: 检查店铺名称是否已存在（仅管理员）

**认证**: 需要（管理员）

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 店铺名称 |

**成功响应**:

```json
{
  "code": 200,
  "exists": false
}
```

---

## 10. 获取临时令牌

**接口**: GET `/shop/temp-token`

**描述**: 获取店铺的6位数字临时令牌，用于临时登录

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
    "token": "123456",
    "expires_at": "2025-02-06T14:30:00Z"
  }
}
```

**说明**:
- 临时令牌有效期1小时
- 同一时间只能有一个有效令牌
- 令牌为6位纯数字

---

## 11. 更新订单状态流转配置

**接口**: PUT `/shop/update-order-status-flow`

**描述**: 更新店铺的订单状态流转配置

**认证**: 需要

**请求参数**:

```json
{
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
            "name": "接单",
            "nextStatus": 1,
            "nextStatusLabel": "处理中"
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

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| order_status_flow | object | 是 | 订单状态流转配置 |

**OrderStatus 字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| value | int | 状态值，唯一标识 |
| label | string | 状态显示名称 |
| type | string | 状态类型标识 |
| isFinal | boolean | 是否为终态（终态订单不能再流转）|
| actions | array | 可执行的操作列表 |

**Action 字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 操作名称 |
| nextStatus | int | 操作后的目标状态值 |
| nextStatusLabel | string | 目标状态显示名称 |

**成功响应**:

```json
{
  "code": 200,
  "message": "店铺订单流转状态配置更新成功",
  "data": {
    "shop_id": 1234567890,
    "order_status_flow": { ... }
  }
}
```

---

## 错误码汇总

| 状态码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | 无效的店铺ID | 店铺ID格式错误 |
| 400 | 无效的有效期格式 | valid_until 格式错误 |
| 400 | 无效的请求数据 | 请求参数缺失或格式错误 |
| 401 | 未获取到用户信息 | 认证失败 |
| 403 | 无权操作 | 非管理员尝试修改有效期 |
| 404 | 店铺不存在 | 店铺ID不存在 |
| 409 | 店主用户名已存在 | 创建时用户名重复 |
| 409 | 存在关联商品，无法删除店铺 | 店铺下有商品 |
| 409 | 存在关联订单，无法删除店铺 | 店铺下有订单 |
| 500 | 服务器内部错误 | 服务器处理异常 |
