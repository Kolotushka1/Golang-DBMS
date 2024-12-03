# Создание таблиц

## Пример: Создание таблиц `users` и `orders`

### SQL-команды

```sql
CREATE TABLE users (id INTEGER AUTO_INCREMENT, name STRING, age INTEGER);

CREATE TABLE orders ( id INTEGER AUTO_INCREMENT, user_id INTEGER, product_name STRING, order_date STRING );