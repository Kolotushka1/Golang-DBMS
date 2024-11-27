package database

import (
	"encoding/json"
	"fmt"
	"os"
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
			db.Tables[strings.ToLower(table.Name)] = &table
		}
	}
	return nil
}
