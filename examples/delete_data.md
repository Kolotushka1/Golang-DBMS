
### 6. `examples/delete_data.md`

# Удаление данных

## Пример: Удаление данных из таблиц

### SQL-команды

```sql
-- Удалить заказы с названием продукта 'Smartphone'
DELETE FROM orders WHERE product_name = 'Smartphone';

-- Удалить пользователей младше 25 лет
DELETE FROM users WHERE age < 25;
