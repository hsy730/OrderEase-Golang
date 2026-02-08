# 商品参数类别 API 文档

## 概述

商品参数类别（Category）用于定义商品的规格参数，如大小、甜度、颜色等。每个商品可以拥有多个参数类别，每个类别下可以有多个选项。

### 数据模型

#### ProductOptionCategory（商品参数类别）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 类别唯一标识 |
| product_id | int64 | 所属商品ID |
| name | string | 类别名称，如"大小"、"甜度" |
| is_required | bool | 是否必填 |
| is_multiple | bool | 是否允许多选 |
| display_order | int | 显示顺序 |
| options | []ProductOption | 选项列表 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

#### ProductOption（商品参数选项）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 选项唯一标识 |
| category_id | int64 | 所属类别ID |
| name | string | 选项名称，如"小杯"、"无糖" |
| price_adjustment | float64 | 价格调整值（可为正负） |
| display_order | int | 显示顺序 |
| is_default | bool | 是否为默认选项 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

---

## 接口列表

### 1. 创建商品参数类别

**注意**: 参数类别通常随商品一起创建或更新，不单独提供创建接口。请使用 [商品创建接口](./api_product.md) 或 [商品更新接口](./api_product.md)。

---

### 2. 更新商品参数类别

**注意**: 参数类别通常随商品一起更新。更新商品时传入 `option_categories` 字段会替换该商品的所有参数类别。

**接口**: `PUT /product/update`

**详细说明**: 参见 [商品更新接口](./api_product.md)

**请求示例**:
```json
{
  "name": "更新后的产品名称",
  "option_categories": [
    {
      "name": "大小",
      "is_required": true,
      "is_multiple": false,
      "display_order": 1,
      "options": [
        {
          "name": "小杯",
          "price_adjustment": 0,
          "display_order": 1
        },
        {
          "name": "中杯",
          "price_adjustment": 5,
          "display_order": 2
        },
        {
          "name": "大杯",
          "price_adjustment": 10,
          "display_order": 3
        }
      ]
    },
    {
      "name": "甜度",
      "is_required": true,
      "is_multiple": false,
      "display_order": 2,
      "options": [
        {
          "name": "无糖",
          "price_adjustment": 0,
          "display_order": 1
        },
        {
          "name": "半糖",
          "price_adjustment": 0,
          "display_order": 2
        },
        {
          "name": "全糖",
          "price_adjustment": 0,
          "display_order": 3
        }
      ]
    }
  ]
}
```

---

### 3. 获取商品参数类别

**注意**: 参数类别信息随商品详情一起返回。

**接口**: `GET /product/detail`

**详细说明**: 参见 [商品详情接口](./api_product.md)

**响应示例**:
```json
{
  "code": 200,
  "data": {
    "id": "1234567890123456789",
    "shop_id": 1234567890,
    "name": "珍珠奶茶",
    "description": "香浓奶茶配Q弹珍珠",
    "price": 15.0,
    "stock": 100,
    "image_url": "http://example.com/image.jpg",
    "status": "online",
    "option_categories": [
      {
        "id": "9876543210987654321",
        "product_id": "1234567890123456789",
        "name": "大小",
        "is_required": true,
        "is_multiple": false,
        "display_order": 1,
        "options": [
          {
            "id": "1111111111111111111",
            "category_id": "9876543210987654321",
            "name": "中杯",
            "price_adjustment": 0,
            "display_order": 1,
            "is_default": true
          },
          {
            "id": "2222222222222222222",
            "category_id": "9876543210987654321",
            "name": "大杯",
            "price_adjustment": 5,
            "display_order": 2,
            "is_default": false
          }
        ]
      },
      {
        "id": "8765432109876543210",
        "product_id": "1234567890123456789",
        "name": "甜度",
        "is_required": true,
        "is_multiple": false,
        "display_order": 2,
        "options": [
          {
            "id": "3333333333333333333",
            "category_id": "8765432109876543210",
            "name": "无糖",
            "price_adjustment": 0,
            "display_order": 1,
            "is_default": false
          },
          {
            "id": "4444444444444444444",
            "category_id": "8765432109876543210",
            "name": "半糖",
            "price_adjustment": 0,
            "display_order": 2,
            "is_default": true
          },
          {
            "id": "5555555555555555555",
            "category_id": "8765432109876543210",
            "name": "全糖",
            "price_adjustment": 0,
            "display_order": 3,
            "is_default": false
          }
        ]
      }
    ],
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### 4. 删除商品参数类别

**注意**: 参数类别随商品更新时替换或随商品删除而删除，不提供单独删除接口。

- 更新商品时传入空的 `option_categories` 数组可删除所有参数类别
- 删除商品时会级联删除其所有参数类别和选项

---

## 字段说明

### ProductOptionCategory 字段详解

| 字段 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| name | 是 | - | 类别名称，如"大小"、"甜度"、"温度" |
| is_required | 否 | false | 是否必须选择，true表示用户必须选择该类别下的一个选项 |
| is_multiple | 否 | false | 是否允许多选，true表示用户可以选择多个选项 |
| display_order | 否 | 0 | 显示顺序，数值越小越靠前 |
| options | 是 | - | 选项列表，至少包含一个选项 |

### ProductOption 字段详解

| 字段 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| name | 是 | - | 选项名称，如"小杯"、"无糖"、"去冰" |
| price_adjustment | 否 | 0 | 价格调整值，正数表示加价，负数表示减价 |
| display_order | 否 | 0 | 显示顺序，数值越小越靠前 |
| is_default | 否 | false | 是否为默认选中选项 |

---

## 使用场景示例

### 场景1：奶茶商品参数配置

```json
{
  "option_categories": [
    {
      "name": "大小",
      "is_required": true,
      "is_multiple": false,
      "display_order": 1,
      "options": [
        { "name": "中杯", "price_adjustment": 0, "display_order": 1, "is_default": true },
        { "name": "大杯", "price_adjustment": 5, "display_order": 2 }
      ]
    },
    {
      "name": "甜度",
      "is_required": true,
      "is_multiple": false,
      "display_order": 2,
      "options": [
        { "name": "无糖", "price_adjustment": 0, "display_order": 1 },
        { "name": "微糖", "price_adjustment": 0, "display_order": 2 },
        { "name": "半糖", "price_adjustment": 0, "display_order": 3, "is_default": true },
        { "name": "全糖", "price_adjustment": 0, "display_order": 4 }
      ]
    },
    {
      "name": "温度",
      "is_required": true,
      "is_multiple": false,
      "display_order": 3,
      "options": [
        { "name": "常温", "price_adjustment": 0, "display_order": 1 },
        { "name": "去冰", "price_adjustment": 0, "display_order": 2, "is_default": true },
        { "name": "少冰", "price_adjustment": 0, "display_order": 3 },
        { "name": "正常冰", "price_adjustment": 0, "display_order": 4 }
      ]
    },
    {
      "name": "加料",
      "is_required": false,
      "is_multiple": true,
      "display_order": 4,
      "options": [
        { "name": "珍珠", "price_adjustment": 2, "display_order": 1 },
        { "name": "椰果", "price_adjustment": 2, "display_order": 2 },
        { "name": "布丁", "price_adjustment": 3, "display_order": 3 }
      ]
    }
  ]
}
```

### 场景2：服装商品参数配置

```json
{
  "option_categories": [
    {
      "name": "颜色",
      "is_required": true,
      "is_multiple": false,
      "display_order": 1,
      "options": [
        { "name": "黑色", "price_adjustment": 0, "display_order": 1 },
        { "name": "白色", "price_adjustment": 0, "display_order": 2 },
        { "name": "红色", "price_adjustment": 0, "display_order": 3 }
      ]
    },
    {
      "name": "尺码",
      "is_required": true,
      "is_multiple": false,
      "display_order": 2,
      "options": [
        { "name": "S", "price_adjustment": 0, "display_order": 1 },
        { "name": "M", "price_adjustment": 0, "display_order": 2 },
        { "name": "L", "price_adjustment": 0, "display_order": 3 },
        { "name": "XL", "price_adjustment": 0, "display_order": 4 }
      ]
    }
  ]
}
```

---

## 注意事项

1. **参数类别与商品绑定**: 参数类别是商品的附属属性，不能独立存在
2. **级联操作**: 
   - 创建商品时可同时创建参数类别
   - 更新商品时会替换所有参数类别（全量更新）
   - 删除商品会级联删除其所有参数类别和选项
3. **选项限制**: 每个类别至少需要一个选项
4. **价格计算**: 商品最终价格 = 商品基础价格 + 所选选项的 price_adjustment 总和
5. **默认值**: 每个类别最多只能有一个默认选项（is_default = true）

---

## 相关接口

- [商品创建接口](./api_product.md#创建产品)
- [商品更新接口](./api_product.md#更新产品)
- [商品详情接口](./api_product.md#获取产品详情)
- [商品列表接口](./api_product.md#获取产品列表)
