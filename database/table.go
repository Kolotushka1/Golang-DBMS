package database

import "strings"

func getColumnIndex(table *Table, columnName string) int {
	for i, col := range table.Columns {
		if strings.ToLower(col.Name) == strings.ToLower(columnName) {
			return i
		}
	}
	return -1
}
