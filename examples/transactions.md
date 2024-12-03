
### 8. `examples/transactions.md`

# Транзакции

## Пример: Использование транзакций

### SQL-команды

```sql
-- Начать транзакцию, вставить данные и откатить
BEGIN;
INSERT INTO users (name, age) VALUES ('Dave', 40);
ROLLBACK;

-- Начать транзакцию, обновить данные и зафиксировать
BEGIN;
UPDATE users SET age = age - 1 WHERE name = 'Bob';
COMMIT;
