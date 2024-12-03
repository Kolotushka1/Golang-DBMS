
### 4. `examples/select_data.md`

# Выборка данных

## Пример: Выборка данных из таблиц

### SQL-команды

```sql
-- Выбрать всех пользователей
SELECT * FROM users;

-- Выбрать имена пользователей старше 30 лет
SELECT name FROM users WHERE age > 30;

-- Получить имена пользователей и названия их заказов
SELECT users.name, orders.product_name FROM users JOIN orders ON users.id = orders.user_id;

-- Сложный запрос с условием
SELECT users.name, orders.product_name FROM users JOIN orders ON users.id = orders.user_id WHERE (users.age < 30 AND orders.product_name = 'Laptop') OR users.name = 'Alice';