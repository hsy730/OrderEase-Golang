# OrderEase 数据库迁移工具

## 功能说明

这个工具用于生成数据库迁移SQL脚本，将表的自增ID主键迁移为雪花ID（Snowflake ID）主键，即移除 `AUTO_INCREMENT` 属性。

**特点**：
- ✅ 纯SQL生成，不需要连接数据库
- ✅ 不修改现有数据，只修改表结构
- ✅ 包含详细的警告说明和验证步骤
- ✅ 开箱即用，零依赖

## 快速开始

### 1. 运行工具生成SQL

```bash
cd src/cmd/migrate
go run main.go
```

输出：
```
================================================================================
OrderEase 数据库迁移工具 - 雪花ID自增禁用SQL生成
================================================================================

✅ 迁移SQL已生成到: migrations_remove_auto_increment.sql

使用方法:
1. 查看生成的SQL文件
2. 在MySQL中执行:
   mysql -u root -p orderease < migrations_remove_auto_increment.sql

⚠️  注意: 执行前请备份数据库!
```

### 2. 查看生成的SQL

```bash
cat migrations_remove_auto_increment.sql
```

### 3. 执行迁移（可选）

```bash
mysql -u root -p orderease < migrations_remove_auto_increment.sql
```

或使用MySQL客户端：

```bash
mysql -u root -p
use orderease;
source /path/to/migrations_remove_auto_increment.sql;
```

## 生成的SQL文件结构

```sql
-- ============================================
-- OrderEase 数据库迁移脚本
-- 移除所有雪花ID表的自增属性
-- 生成时间: 2026-01-13 23:22:04
-- ============================================

-- ⚠️  重要提示:
-- 1. 执行前请备份数据库!
-- 2. 此操作会移除主键的自增属性
-- 3. 现有数据不会被修改或删除
-- 4. 执行后需要使用雪花ID生成器创建新记录

SET FOREIGN_KEY_CHECKS = 0;

-- 移除 shops 表的自增属性
ALTER TABLE `shops` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 products 表的自增属性
ALTER TABLE `products` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- ... 其他表

SET FOREIGN_KEY_CHECKS = 1;

-- 验证迁移结果
SHOW CREATE TABLE shops;
-- ...
```

## 涉及的表

以下9个表会被迁移：

| 表名 | 说明 |
|-----|------|
| `shops` | 店铺表 |
| `products` | 商品表 |
| `orders` | 订单表 |
| `order_items` | 订单项表 |
| `order_item_options` | 订单项选项表 |
| `users` | 用户表 |
| `temp_tokens` | 临时令牌表 |
| `product_option_categories` | 商品参数类别表 |
| `product_options` | 商品参数选项表 |

## 验证迁移结果

执行迁移后，使用以下命令验证：

```sql
SHOW CREATE TABLE shops;
```

**迁移前**：
```sql
CREATE TABLE `shops` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  ...
) ENGINE=InnoDB;
```

**迁移后**：
```sql
CREATE TABLE `shops` (
  `id` bigint unsigned NOT NULL,
  ...
) ENGINE=InnoDB;
```

注意：**不应该包含** `AUTO_INCREMENT`。

## ⚠️ 重要提示

### 现有数据

- ✅ 迁移**不会删除或修改现有数据**
- ✅ 现有的自增ID（1, 2, 3...）会被保留
- ⚠️ 新插入的记录将使用雪花ID（通常是很大的数字）
- ✅ 两种ID类型的数据可以共存

### 应用代码

迁移后，创建新记录**必须使用雪花ID生成器**：

```go
import "orderease/utils"

// ✅ 正确
shop := &shop.Shop{
    ID:   utils.GenerateSnowflakeID(),
    Name: "新店铺",
}

// ❌ 错误
shop := &shop.Shop{
    Name: "新店铺",  // 缺少ID，可能导致问题
}
```

### 生产环境

1. **备份数据库**：
   ```bash
   mysqldump -u root -p orderease > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **在测试环境先测试**

3. **选择低峰时段执行**

4. **查看生成的SQL文件内容，确认无误后再执行**

## 测试

运行单元测试：

```bash
# 进入工具目录
cd src/cmd/migrate

# 运行测试
go test -v -cover

# 运行性能测试
go test -bench=. -benchmem
```

预期输出：
```
=== RUN   TestGenerateMigrations
--- PASS: TestGenerateMigrations (0.00s)
=== RUN   TestWriteMigrationsToFile
--- PASS: TestWriteMigrationsToFile (0.02s)
=== RUN   TestRepeat
--- PASS: TestRepeat (0.00s)
PASS
coverage: 66.0% of statements
ok      orderease/cmd/migrate
```

## 故障排除

### 问题1：执行SQL时报错

```
Error 1146: Table 'orderease.shops' doesn't exist
```

**原因**：表不存在

**解决**：检查数据库中是否已创建表。如果表不存在，GORM会在应用首次启动时自动创建。

### 问题2：权限不足

```
Error 1227: Access denied; you need (at least one of the) the SUPER privilege(s)
```

**解决**：使用有足够权限的数据库用户（需要 ALTER 权限）。

### 问题3：外键约束错误

```
Error 1451: Cannot delete or update a parent row: a foreign key constraint fails
```

**解决**：SQL脚本中已包含 `SET FOREIGN_KEY_CHECKS = 0;` 来临时禁用外键检查。如果仍有问题，请手动检查外键关系。

## 常见问题

### Q: 会删除数据吗？

A: **不会**。这个SQL脚本只修改表结构（移除AUTO_INCREMENT属性），不涉及数据操作。

### Q: 执行后可以回滚吗？

A: 可以，但**不推荐**。回滚SQL：
```sql
ALTER TABLE `shops` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT;
```

### Q: 新记录的ID会是多少？

A: 使用雪花ID生成器，通常是类似 `123456789012345678` 这样的大数字。

### Q: 现有ID=1,2,3的数据会受影响吗？

A: **不会**。现有数据完全不受影响，保留原ID。

### Q: 可以只迁移部分表吗？

A: 可以。编辑生成的 `migrations_remove_auto_increment.sql` 文件，注释掉不需要的表。

## 技术支持

- 项目文档：`D:\selfcoding\gowork\OrderEase-Golang\CLAUDE.md`
- DDD设计：`DDD_战略设计方案.md`
- 单元测试：`src/cmd/migrate/main_test.go`

## License

MIT
