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

-- 移除 orders 表的自增属性
ALTER TABLE `orders` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 order_items 表的自增属性
ALTER TABLE `order_items` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 order_item_options 表的自增属性
ALTER TABLE `order_item_options` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 users 表的自增属性
ALTER TABLE `users` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 temp_tokens 表的自增属性
ALTER TABLE `temp_tokens` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 product_option_categories 表的自增属性
ALTER TABLE `product_option_categories` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

-- 移除 product_options 表的自增属性
ALTER TABLE `product_options` MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL;

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 验证迁移结果
-- ============================================
-- 请运行以下命令验证 AUTO_INCREMENT 已移除:

SHOW CREATE TABLE shops;
SHOW CREATE TABLE products;
SHOW CREATE TABLE orders;
SHOW CREATE TABLE order_items;
SHOW CREATE TABLE order_item_options;
SHOW CREATE TABLE users;
SHOW CREATE TABLE temp_tokens;
SHOW CREATE TABLE product_option_categories;
SHOW CREATE TABLE product_options;

-- 预期结果: CREATE TABLE 语句中不应该包含 AUTO_INCREMENT
-- ============================================
