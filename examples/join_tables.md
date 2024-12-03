
### 7. `examples/join_tables.md`

# Соединение таблиц (JOIN)

## Пример: Использование JOIN для объединения таблиц `users` и `orders`

### SQL-команды

```sql
-- LEFT JOIN для получения всех пользователей и их заказов (если есть)
SELECT users.name, orders.product_name FROM users LEFT JOIN orders ON users.id = orders.user_id;

-- Получить пользователей без заказов
SELECT users.name FROM users LEFT JOIN orders ON users.id = orders.user_id WHERE orders.id IS NULL;