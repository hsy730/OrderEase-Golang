## 数据管理接口

### 1. 导出数据
- **接口**: GET `/api/v1/data/export`
- **描述**: 导出所有数据为ZIP格式，包含多个CSV文件
- **响应**: ZIP文件，包含以下CSV文件：
  - users.csv: 用户数据
  - products.csv: 商品数据
  - tags.csv: 标签数据
  - product_tags.csv: 商品标签关联数据
  - orders.csv: 订单数据
  - order_items.csv: 订单项数据

#### CSV文件格式示例：

1. users.csv:
```csv
id,name,phone,address,type,created_at,updated_at
1,张三,13800138000,北京市朝阳区xxx街道,delivery,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z
```

2. products.csv:
```csv
id,name,description,price,stock,image_url,created_at,updated_at
1,商品1,商品描述,99.90,100,/uploads/products/product_1_1234567890.jpg,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z
```

3. tags.csv:
```csv
id,name,description,created_at,updated_at
1,新品,新上架商品,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z
```

4. product_tags.csv:
```csv
product_id,tag_id,created_at,updated_at
1,1,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z
```

5. orders.csv:
```csv
id,user_id,total_price,status,remark,created_at,updated_at
1,1,299.70,pending,订单备注,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z
```

6. order_items.csv:
```csv
id,order_id,product_id,quantity,price
1,1,1,3,99.90
```

### 2. 导入数据
- **接口**: POST `/api/v1/data/import`
- **描述**: 从ZIP文件导入数据（会先清空现有数据）
- **请求参数**: 
  - file: ZIP文件（multipart/form-data）
- **响应**:
```json
{
    "message": "数据导入成功"
}
```

### 使用说明

1. 导出数据
```bash
# 导出所有数据为ZIP文件
curl -X GET "https://your-domain/api/v1/data/export" > backup_20240314_150000.zip
```

2. 导入数据
```bash
# 从ZIP文件导入数据
curl -X POST "https://your-domain/api/v1/data/import" \
  -F "file=@backup_20240314_150000.zip"
```

### 注意事项
1. 导出的ZIP文件包含六个CSV文件，分别对应不同的数据表
2. 每个CSV文件必须包含正确的表头
3. 导入时会先清空所有表中的现有数据
4. 导入按照以下顺序处理文件（保证外键约束）：
   - users.csv
   - products.csv
   - tags.csv
   - product_tags.csv
   - orders.csv
   - order_items.csv
5. 所有日期时间使用RFC3339格式
6. 导入过程使用事务确保数据一致性
7. 图片文件需要单独备份，CSV只包含图片路径
8. 如果导入过程中出错，会回滚所有更改
9. ZIP文件中的CSV文件名必须严格匹配上述文件名

## 错误响应
所有接口在发生错误时会返回相应的HTTP状态码和错误信息：

- 400 Bad Request: 请求参数错误或文件格式不正确
- 404 Not Found: 资源未找到
- 500 Internal Server Error: 服务器内部错误

错误响应格式：
```json
{
    "error": "错误信息描述"
}
```
