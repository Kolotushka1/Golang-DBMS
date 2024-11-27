# 📚 Простая СУБД на Go: Подробная Документация 🚀

Добро пожаловать в подробную документацию вашей собственной простой системы управления базами данных (СУБД) на языке программирования Go! Эта документация поможет вам разобраться в структуре проекта, функциональности отдельных файлов, а также предоставит примеры различных операций, включая создание таблиц, выполнение запросов с условиями `WHERE` и `JOIN`, а также работу с транзакциями. 📈

## 🔍 Обзор Файлов

### 1. `main.go` 🏁
**Назначение:**  
Главный файл, который запускает вашу СУБД. Он отвечает за считывание SQL-запросов от пользователя и передачу их на обработку.

**Ключевые Функции:**
* Чтение ввода пользователя.
* Вызов функции `ExecuteSQL` для выполнения запросов.
* Вывод результатов или ошибок.

### 2. `database/database.go` 🗃️
**Назначение:**  
Основной файл, содержащий структуру `Database`, которая управляет таблицами, транзакциями и обеспечивает взаимодействие с другими компонентами.

**Ключевые Функции:**
* Создание и управление таблицами.
* Вставка, выборка, обновление и удаление данных с поддержкой условий `WHERE`.
* Управление транзакциями (`BEGIN`, `COMMIT`, `ROLLBACK`).
* Загрузка данных с диска и сохранение их обратно.

### 3. `database/join.go` 🔗
**Назначение:**  
Файл, отвечающий за реализацию операций `JOIN` (INNER JOIN, LEFT JOIN, RIGHT JOIN) между таблицами.

**Ключевые Функции:**
* Выполнение различных типов соединений между таблицами.
* Сравнение значений столбцов для соединения.
* Обработка результатов соединений.

### 4. `database/sql_parser.go` 📝
**Назначение:**  
Парсер SQL-запросов. Этот файл разбирает вводимые пользователем SQL-запросы, включая условия `WHERE`, и вызывает соответствующие функции для их выполнения.

**Ключевые Функции:**
* Разбор и интерпретация SQL-запросов.
* Выделение команд (`CREATE`, `INSERT`, `SELECT`, `UPDATE`, `DELETE` и т.д.).
* Обработка условий `JOIN` и `WHERE`.
* Выборка конкретных столбцов.
* Вызов функций из `database.go` для выполнения операций.

### 5. `database/storage.go` 💾
**Назначение:**  
Файл, отвечающий за хранение данных на диске. Таблицы сохраняются в формате JSON и загружаются при запуске СУБД.

**Ключевые Функции:**
* Сохранение таблиц в файлы JSON.
* Загрузка таблиц из файлов JSON при инициализации.
* Обеспечение постоянства данных между сессиями.

### 6. `database/table.go` 📋
**Назначение:**  
Файл, определяющий структуру таблицы и предоставляющий вспомогательные функции для работы со столбцами.

**Ключевые Функции:**
* Получение индекса столбца по его имени.
* Управление структурой таблицы.

### 7. `database/transaction.go` 🔄
**Назначение:**  
Файл, отвечающий за реализацию транзакций. Позволяет выполнять группы операций атомарно.

**Ключевые Функции:**
* Начало транзакции (`BEGIN`).
* Фиксация транзакции (`COMMIT`).
* Откат транзакции (`ROLLBACK`).
* Отслеживание операций внутри транзакции для возможности отката.

## 🛠 Примеры Операций

### 🛠 Создание Таблиц
```sql
CREATE TABLE users (id INTEGER, name STRING, email STRING);
CREATE TABLE orders ( order_id INTEGER, user_id INTEGER, amount FLOAT );
```
### 🛠 Вставка Данных
```sql
INSERT INTO users VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users VALUES (2, 'Bob', 'bob@example.com');
INSERT INTO users VALUES (3, 'Charlie', 'charlie@example.com');

INSERT INTO orders VALUES (101, 1, 250.75);
INSERT INTO orders VALUES (102, 1, 89.50);
INSERT INTO orders VALUES (103, 2, 150.00);
```

### 🛠 Выполнение Запросов с JOIN
#### 1. INNER JOIN
```sql
SELECT users.name, orders.amount FROM users JOIN orders ON users.id = orders.user_id;
```
#### Вывод:
```sql
Выполнение INNER JOIN между 'users' и 'orders' по столбцам 'id' и 'user_id'
[Alice 250.75]
[Alice 89.5]
[Bob 150]
```
#### 2. LEFT JOIN
```sql
SELECT users.name, orders.amount FROM users LEFT JOIN orders ON users.id = orders.user_id;
```
#### Вывод:
```sql
Выполнение LEFT JOIN между 'users' и 'orders' по столбцам 'id' и 'user_id'
[Alice 250.75]
[Alice 89.5]
[Bob 150]
[Charlie <nil>]
```
#### 3. RIGHT JOIN
```sql
SELECT users.name, orders.amount FROM users RIGHT JOIN orders ON users.id = orders.user_id;
```
#### Вывод:
```sql
Выполнение RIGHT JOIN между 'users' и 'orders' по столбцам 'id' и 'user_id'
[Alice 250.75]
[Alice 89.5]
[Bob 150]
[<nil> 300]
```
### 🛠 Работа с Транзакциями
#### 1. Начало Транзакции
```sql
BEGIN;
```
#### 2. Вставка Данных в Транзакции
```sql
INSERT INTO users VALUES (4, 'Diana', 'diana@example.com');
INSERT INTO orders VALUES (105, 4, 400.00);
```
#### 3. Фиксация Транзакции
```sql
COMMIT;
```
#### 4. Откат Транзакции
```sql
BEGIN;
INSERT INTO users VALUES (5, 'Eve', 'eve@example.com');
DELETE FROM orders;
ROLLBACK;
```
```sql
Транзакция начата.
Данные вставлены успешно.
Данные удалены успешно.
Транзакция откатана.
```
### 🛠 Дополнительные Примеры
#### 1. Обновление Данных
```sql
UPDATE users SET email = 'alice@newdomain.com';
```
#### 2. Удаление Данных
```sql
DELETE FROM orders;
```
#### 3. LEFT JOIN с Условием WHERE
```sql
SELECT users.name, orders.amount FROM users LEFT JOIN orders ON users.id = orders.user_id WHERE orders.amount > 100;
```
#### 4. Обновление Данных с Условием WHERE
```sql
UPDATE users SET email = 'alice@newdomain.com' WHERE name = 'Alice';
```