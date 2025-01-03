## 数据管理接口

### 1. 导出数据
- **接口**: GET `/api/v1/data/export`
- **描述**: 导出所有数据为CSV格式
- **响应**: CSV文件，包含以下数据表：
  - 商品表
  - 订单表
  - 订单项表
- **CSV格式示例**:
```csv
id,name,description,price,stock,image_url,created_at,updated_at
1,商品1,商品描述,99.90,100,/uploads/products/product_1_1234567890.jpg,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z

id,user_id,total_price,status,remark,created_at,updated_at
1,1,299.70,pending,订单备注,2024-03-14T12:00:00Z,2024-03-14T12:00:00Z

id,order_id,product_id,quantity,price
1,1,1,3,99.90
```

### 2. 导入数据
- **接口**: POST `/api/v1/data/import`
- **描述**: 从CSV文件导入数据
- **请求参数**: 
  - file: CSV文件（multipart/form-data）
- **响应**:
```json
{
    "message": "数据导入成功"
}
```

### 使用说明

1. 导出数据
```bash
# 导出所有数据为CSV文件
curl -X GET "https://your-domain/api/v1/data/export" > backup.csv
```

2. 导入数据
```bash
# 从CSV文件导入数据
curl -X POST "https://your-domain/api/v1/data/import" \
  -F "file=@backup.csv"
```

### 注意事项
1. CSV文件必须包含正确的表头
2. 数据表之间使用空行分隔
3. 导入时会自动识别数据类型
4. 所有日期时间使用RFC3339格式
5. 导入过程使用事务确保数据一致性
6. 图片文件需要单独备份，CSV只包含图片路径

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
