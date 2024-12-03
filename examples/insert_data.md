
### 3. `examples/insert_data.md`

```markdown
# Вставка данных

## Пример: Вставка данных в таблицы `users` и `orders`

### SQL-команды

```sql
-- Вставка пользователей
INSERT INTO users (name, age) VALUES ('Alice', 25);
INSERT INTO users (name, age) VALUES ('Bob', 32);
INSERT INTO users (name, age) VALUES ('Charlie', 28);

-- Вставка заказов
INSERT INTO orders (user_id, product_name, order_date) VALUES (1, 'Laptop', '2023-01-15');
INSERT INTO orders (user_id, product_name, order_date) VALUES (2, 'Smartphone', '2023-02-20');
INSERT INTO orders (user_id, product_name, order_date) VALUES (1, 'Tablet', '2023-03-10');
