package database

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (db *Database) saveTableToDisk(tableName string) error {
	table, exists := db.Tables[tableName]
	if !exists {
		return fmt.Errorf("таблица '%s' не существует", tableName)
	}
	data, err := json.MarshalIndent(table, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка маршалинга таблицы '%s': %v", tableName, err)
	}
	filename := fmt.Sprintf("%s.json", tableName)
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи таблицы '%s' на диск: %v", tableName, err)
	}
	return nil
}

func (db *Database) LoadFromDisk() error {
	files, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("ошибка чтения директории: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			data, err := os.ReadFile(file.Name())
			if err != nil {
				return fmt.Errorf("ошибка чтения файла '%s': %v", file.Name(), err)
			}
			var table Table
			err = json.Unmarshal(data, &table)
			if err != nil {
				return fmt.Errorf("ошибка маршалинга файла '%s': %v", file.Name(), err)
			}

			for i, row := range table.Rows {
				for j, col := range table.Columns {
					value := row[j]
					correctedValue, err := correctType(value, col.Type)
					if err != nil {
						return fmt.Errorf("Ошибка преобразования значения '%v' в столбце '%s' в строке %d: %v", value, col.Name, i, err)
					}
					table.Rows[i][j] = correctedValue
				}
			}

			db.Tables[strings.ToLower(table.Name)] = &table
		}
	}
	return nil
}

func correctType(value interface{}, dataType DataType) (interface{}, error) {
	switch dataType {
	case INTEGER:
		switch v := value.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		case string:
			intVal, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("не удалось преобразовать '%v' в INTEGER", v)
			}
			return intVal, nil
		default:
			return nil, fmt.Errorf("неподдерживаемый тип '%T' для INTEGER", v)
		}
	case FLOAT:
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case string:
			floatVal, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("не удалось преобразовать '%v' в FLOAT", v)
			}
			return floatVal, nil
		default:
			return nil, fmt.Errorf("неподдерживаемый тип '%T' для FLOAT", v)
		}
	case STRING:
		return fmt.Sprintf("%v", value), nil
	default:
		return nil, fmt.Errorf("неизвестный тип данных '%v'", dataType)
	}
}
