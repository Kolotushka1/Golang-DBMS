package database

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type DataType int

const (
	STRING DataType = iota
	INTEGER
	FLOAT
)

func (dt DataType) String() string {
	switch dt {
	case STRING:
		return "STRING"
	case INTEGER:
		return "INTEGER"
	case FLOAT:
		return "FLOAT"
	default:
		return "UNKNOWN"
	}
}

type Column struct {
	Name          string
	Type          DataType
	AutoIncrement bool
}

type Table struct {
	Name            string
	Columns         []Column
	Rows            [][]interface{}
	autoIncrementID map[string]int
}

type Database struct {
	Tables      map[string]*Table
	mu          sync.RWMutex
	transaction *Transaction
}

func NewDatabase() *Database {
	db := &Database{
		Tables: make(map[string]*Table),
	}
	err := db.LoadFromDisk()
	if err != nil {
		fmt.Println("Ошибка загрузки данных с диска:", err)
	}
	return db
}

func (db *Database) ExecuteSQL(query string) ([][]interface{}, error) {
	return ParseAndExecute(db, query)
}

func (db *Database) CreateTable(tableName string, columns []Column) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tableName = strings.ToLower(tableName)
	if _, exists := db.Tables[tableName]; exists {
		return fmt.Errorf("таблица '%s' уже существует", tableName)
	}

	autoIncMap := make(map[string]int)
	for _, col := range columns {
		if col.AutoIncrement {
			autoIncMap[col.Name] = 0
		}
	}

	table := &Table{
		Name:            tableName,
		Columns:         columns,
		Rows:            [][]interface{}{},
		autoIncrementID: autoIncMap,
	}

	db.Tables[tableName] = table

	err := db.saveTableToDisk(tableName)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Insert(tableName string, values []string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tableName = strings.ToLower(tableName)
	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("таблица '%s' не существует", tableName)
	}

	if len(values) < len(table.Columns) {
		newValues := make([]interface{}, len(table.Columns))
		valueIndex := 0
		for i, col := range table.Columns {
			if col.AutoIncrement {
				table.autoIncrementID[col.Name]++
				newValues[i] = table.autoIncrementID[col.Name]
			} else {
				if valueIndex >= len(values) {
					newValues[i] = nil
				} else {
					val, err := parseValue(values[valueIndex], col.Type)
					if err != nil {
						return err
					}
					newValues[i] = val
					valueIndex++
				}
			}
		}
		table.Rows = append(table.Rows, newValues)
	} else {
		newValues := make([]interface{}, len(table.Columns))
		for i, col := range table.Columns {
			if col.AutoIncrement {
				table.autoIncrementID[col.Name]++
				newValues[i] = table.autoIncrementID[col.Name]
			} else {
				if i >= len(values) {
					newValues[i] = nil
				} else {
					val, err := parseValue(values[i], col.Type)
					if err != nil {
						return err
					}
					newValues[i] = val
				}
			}
		}
		table.Rows = append(table.Rows, newValues)
	}

	// Сохранение на диск
	err := db.saveTableToDisk(tableName)
	if err != nil {
		return err
	}

	return nil
}

func parseValue(value string, dataType DataType) (interface{}, error) {
	switch dataType {
	case INTEGER:
		return strconv.Atoi(value)
	case FLOAT:
		return strconv.ParseFloat(value, 64)
	case STRING:
		return value, nil
	default:
		return nil, fmt.Errorf("неизвестный тип данных %v", dataType)
	}
}

func (db *Database) Select(tableName string, condition *Condition) ([][]interface{}, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	table, exists := db.Tables[strings.ToLower(tableName)]
	if !exists {
		return nil, fmt.Errorf("таблица '%s' не существует", tableName)
	}

	var columnNames []string
	for _, col := range table.Columns {
		columnNames = append(columnNames, col.Name)
	}

	var result [][]interface{}
	for _, row := range table.Rows {
		if condition != nil {
			match, err := evaluateCondition(row, columnNames, condition)
			if err != nil {
				return nil, err
			}
			if match {
				result = append(result, row)
			}
		} else {
			result = append(result, row)
		}
	}
	return result, nil
}

func (db *Database) Update(tableName string, columnName string, newValue string, condition *Condition) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	tableName = strings.ToLower(tableName)
	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("таблица '%s' не существует", tableName)
	}
	var colIndex int = -1
	var colType DataType
	for i, col := range table.Columns {
		if strings.ToLower(col.Name) == strings.ToLower(columnName) {
			colIndex = i
			colType = col.Type
			break
		}
	}
	if colIndex == -1 {
		return fmt.Errorf("столбец '%s' не найден в таблице '%s'", columnName, tableName)
	}

	var columnNames []string
	for _, col := range table.Columns {
		columnNames = append(columnNames, col.Name)
	}

	for rowIdx, row := range table.Rows {
		if condition != nil {
			match, err := evaluateCondition(row, columnNames, condition)
			if err != nil {
				return err
			}
			if !match {
				continue
			}
		}

		var val interface{}
		switch colType {
		case STRING:
			val = newValue
		case INTEGER:
			intval, err := strconv.Atoi(newValue)
			if err != nil {
				return fmt.Errorf("ошибка преобразования '%s' в INTEGER: %v", newValue, err)
			}
			val = intval
		case FLOAT:
			floatval, err := strconv.ParseFloat(newValue, 64)
			if err != nil {
				return fmt.Errorf("ошибка преобразования '%s' в FLOAT: %v", newValue, err)
			}
			val = floatval
		}
		oldValue := row[colIndex]
		table.Rows[rowIdx][colIndex] = val
		if db.transaction != nil {
			op := Operation{
				Type:       "UPDATE",
				TableName:  tableName,
				RowIndex:   rowIdx,
				ColumnName: columnName,
				NewValue:   val,
				Data:       oldValue,
			}
			db.transaction.operations = append(db.transaction.operations, op)
		}
	}
	return db.saveTableToDisk(tableName)
}

func (db *Database) Delete(tableName string, condition *Condition) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	tableName = strings.ToLower(tableName)
	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("таблица '%s' не существует", tableName)
	}

	// Создаем список имен столбцов
	var columnNames []string
	for _, col := range table.Columns {
		columnNames = append(columnNames, col.Name)
	}

	var newRows [][]interface{}
	for _, row := range table.Rows {
		deleteRow := false
		if condition != nil {
			match, err := evaluateCondition(row, columnNames, condition)
			if err != nil {
				return err
			}
			if match {
				deleteRow = true
			}
		}
		if deleteRow {
			if db.transaction != nil {
				op := Operation{
					Type:      "DELETE",
					TableName: tableName,
					Data:      row,
				}
				db.transaction.operations = append(db.transaction.operations, op)
			}
			continue
		}
		newRows = append(newRows, row)
	}
	table.Rows = newRows
	return db.saveTableToDisk(tableName)
}

func (db *Database) BeginTransaction() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.transaction != nil {
		return fmt.Errorf("транзакция уже начата")
	}
	db.transaction = &Transaction{
		operations: []Operation{},
	}
	return nil
}
func (db *Database) Commit() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.transaction == nil {
		return fmt.Errorf("нет активной транзакции")
	}
	db.transaction = nil
	return nil
}

func (db *Database) Rollback() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.transaction == nil {
		return fmt.Errorf("нет активной транзакции")
	}
	for i := len(db.transaction.operations) - 1; i >= 0; i-- {
		op := db.transaction.operations[i]
		switch op.Type {
		case "INSERT":
			table, _ := db.Tables[op.TableName]
			if len(table.Rows) > 0 {
				table.Rows = table.Rows[:len(table.Rows)-1]
				db.saveTableToDisk(op.TableName)
			}
		case "DELETE":
			table, _ := db.Tables[op.TableName]
			table.Rows = append(table.Rows, op.Data.([]interface{}))
			db.saveTableToDisk(op.TableName)
		case "UPDATE":
			table, _ := db.Tables[op.TableName]
			colIndex := getColumnIndex(table, op.ColumnName)
			if colIndex != -1 && op.RowIndex < len(table.Rows) {
				table.Rows[op.RowIndex][colIndex] = op.Data
				db.saveTableToDisk(op.TableName)
			}
		}
	}
	db.transaction = nil
	return nil
}
