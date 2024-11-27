package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"SQL/database"
)

func main() {
	db := database.NewDatabase()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Простая СУБД на Go. Введите SQL-запросы или 'EXIT' для выхода.")
	for {
		fmt.Print("SQL> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.TrimSpace(strings.ToUpper(input)) == "EXIT" {
			fmt.Println("Выход из СУБД.")
			break
		}
		result, err := db.ExecuteSQL(input)
		if err != nil {
			fmt.Println("Ошибка:", err)
			continue
		}
		if result != nil {
			for _, row := range result {
				fmt.Println(row)
			}
		}
	}
}
