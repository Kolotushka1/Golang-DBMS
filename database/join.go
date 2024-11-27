package database

import (
	"errors"
	"fmt"
	"strings"
)

func isEqual(a, b interface{}) bool {
	switch aTyped := a.(type) {
	case int:
		switch bTyped := b.(type) {
		case int:
			return aTyped == bTyped
		case float64:
			return float64(aTyped) == bTyped
		}
	case float64:
		switch bTyped := b.(type) {
		case int:
			return aTyped == float64(bTyped)
		case float64:
			return aTyped == bTyped
		}
	case string:
		if bTyped, ok := b.(string); ok {
			return aTyped == bTyped
		}
	}
	return false
}

func (db *Database) Join(table1Name, table2Name, joinColumn1, joinColumn2 string) ([][]interface{}, error) {
	fmt.Printf("Выполнение INNER JOIN между '%s' и '%s' по столбцам '%s' и '%s'\n", table1Name, table2Name, joinColumn1, joinColumn2)
	db.mu.RLock()
	defer db.mu.RUnlock()
	table1Name = strings.ToLower(table1Name)
	table2Name = strings.ToLower(table2Name)

	table1, exists1 := db.Tables[table1Name]
	table2, exists2 := db.Tables[table2Name]

	if !exists1 || !exists2 {
		return nil, errors.New("одна или обе таблицы не существуют")
	}

	joinIndex1 := getColumnIndex(table1, joinColumn1)
	if joinIndex1 == -1 {
		return nil, fmt.Errorf("столбец '%s' не найден в таблице '%s'", joinColumn1, table1Name)
	}

	joinIndex2 := getColumnIndex(table2, joinColumn2)
	if joinIndex2 == -1 {
		return nil, fmt.Errorf("столбец '%s' не найден в таблице '%s'", joinColumn2, table2Name)
	}

	var result [][]interface{}
	for _, row1 := range table1.Rows {
		for _, row2 := range table2.Rows {
			if isEqual(row1[joinIndex1], row2[joinIndex2]) {
				combinedRow := append(row1, row2...)
				result = append(result, combinedRow)
			}
		}
	}
	return result, nil
}

func (db *Database) LeftJoin(table1Name, table2Name, joinColumn1, joinColumn2 string) ([][]interface{}, error) {
	fmt.Printf("Выполнение LEFT JOIN между '%s' и '%s' по столбцам '%s' и '%s'\n", table1Name, table2Name, joinColumn1, joinColumn2)
	db.mu.RLock()
	defer db.mu.RUnlock()
	table1Name = strings.ToLower(table1Name)
	table2Name = strings.ToLower(table2Name)

	table1, exists1 := db.Tables[table1Name]
	table2, exists2 := db.Tables[table2Name]

	if !exists1 || !exists2 {
		return nil, errors.New("одна или обе таблицы не существуют")
	}

	joinIndex1 := getColumnIndex(table1, joinColumn1)
	if joinIndex1 == -1 {
		return nil, fmt.Errorf("столбец '%s' не найден в таблице '%s'", joinColumn1, table1Name)
	}

	joinIndex2 := getColumnIndex(table2, joinColumn2)
	if joinIndex2 == -1 {
		return nil, fmt.Errorf("столбец '%s' не найден в таблице '%s'", joinColumn2, table2Name)
	}

	var result [][]interface{}
	for _, row1 := range table1.Rows {
		matched := false
		for _, row2 := range table2.Rows {
			if isEqual(row1[joinIndex1], row2[joinIndex2]) {
				combinedRow := append(row1, row2...)
				result = append(result, combinedRow)
				matched = true
			}
		}
		if !matched {
			nulls := make([]interface{}, len(table2.Columns))
			for i := range nulls {
				nulls[i] = nil
			}
			combinedRow := append(row1, nulls...)
			result = append(result, combinedRow)
		}
	}
	return result, nil
}

func (db *Database) RightJoin(table1Name, table2Name, joinColumn1, joinColumn2 string) ([][]interface{}, error) {
	fmt.Printf("Выполнение RIGHT JOIN между '%s' и '%s' по столбцам '%s' и '%s'\n", table1Name, table2Name, joinColumn1, joinColumn2)
	return db.LeftJoin(table2Name, table1Name, joinColumn2, joinColumn1)
}
