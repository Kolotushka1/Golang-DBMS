package database

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ConditionType int

const (
	Simple ConditionType = iota
	Compound
)

type Condition struct {
	Type      ConditionType
	Left      *Condition
	Right     *Condition
	LogicalOp string
	Column    string
	Operator  string
	Value     interface{}
}

func ParseAndExecute(db *Database, query string) ([][]interface{}, error) {
	query = strings.TrimSpace(query)
	if strings.HasSuffix(query, ";") {
		query = query[:len(query)-1]
	}

	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil, errors.New("пустой запрос")
	}

	command := strings.ToUpper(tokens[0])

	switch command {
	case "CREATE":
		return handleCreate(db, query, tokens)
	case "INSERT":
		return handleInsert(db, query, tokens)
	case "SELECT":
		return handleSelect(db, query, tokens)
	case "UPDATE":
		return handleUpdate(db, query, tokens)
	case "DELETE":
		return handleDelete(db, query, tokens)
	case "BEGIN":
		err := db.BeginTransaction()
		if err != nil {
			return nil, err
		}
		fmt.Println("Транзакция начата.")
		return nil, nil
	case "COMMIT":
		err := db.Commit()
		if err != nil {
			return nil, err
		}
		fmt.Println("Транзакция зафиксирована.")
		return nil, nil
	case "ROLLBACK":
		err := db.Rollback()
		if err != nil {
			return nil, err
		}
		fmt.Println("Транзакция откатана.")
		return nil, nil
	default:
		return nil, fmt.Errorf("неизвестная команда '%s'", command)
	}
}

func handleCreate(db *Database, query string, tokens []string) ([][]interface{}, error) {
	if len(tokens) < 3 || strings.ToUpper(tokens[1]) != "TABLE" {
		return nil, errors.New("неверный синтаксис CREATE TABLE")
	}
	tableName := tokens[2]
	columnsDefStart := strings.Index(query, "(")
	columnsDefEnd := strings.LastIndex(query, ")")
	if columnsDefStart == -1 || columnsDefEnd == -1 || columnsDefEnd < columnsDefStart {
		return nil, errors.New("неверный синтаксис определения столбцов")
	}
	columnsDef := query[columnsDefStart+1 : columnsDefEnd]
	columnsParts := splitCSV(columnsDef)
	var columns []Column
	for _, part := range columnsParts {
		col := strings.Fields(strings.TrimSpace(part))
		if len(col) < 2 {
			return nil, errors.New("неверный синтаксис определения столбца")
		}
		colName := col[0]
		var colType DataType
		switch strings.ToUpper(col[1]) {
		case "STRING":
			colType = STRING
		case "INTEGER":
			colType = INTEGER
		case "FLOAT":
			colType = FLOAT
		default:
			return nil, fmt.Errorf("неизвестный тип данных '%s'", col[1])
		}

		autoIncrement := false
		if len(col) > 2 && strings.ToUpper(col[2]) == "AUTO_INCREMENT" {
			if colType != INTEGER {
				return nil, errors.New("AUTO_INCREMENT поддерживается только для INTEGER типов")
			}
			autoIncrement = true
		}

		columns = append(columns, Column{Name: colName, Type: colType, AutoIncrement: autoIncrement})
	}
	err := db.CreateTable(tableName, columns)
	if err != nil {
		return nil, err
	}
	fmt.Println("Таблица создана успешно.")
	return nil, nil
}

func handleInsert(db *Database, query string, tokens []string) ([][]interface{}, error) {
	if len(tokens) < 4 || strings.ToUpper(tokens[1]) != "INTO" {
		return nil, errors.New("неверный синтаксис INSERT INTO")
	}
	tableName := tokens[2]
	upperQuery := strings.ToUpper(query)
	valuesIndex := strings.Index(upperQuery, "VALUES")
	if valuesIndex == -1 {
		return nil, errors.New("не найдено ключевое слово VALUES")
	}
	valuesPart := query[valuesIndex+6:]
	valuesPart = strings.TrimSpace(valuesPart)
	if !strings.HasPrefix(valuesPart, "(") || !strings.HasSuffix(valuesPart, ")") {
		return nil, errors.New("неверный синтаксис VALUES")
	}
	valuesPart = valuesPart[1 : len(valuesPart)-1]
	values := splitCSV(valuesPart)
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
		values[i] = strings.Trim(values[i], "'")
	}
	err := db.Insert(tableName, values)
	if err != nil {
		return nil, err
	}
	fmt.Println("Данные вставлены успешно.")
	return nil, nil
}

func handleSelect(db *Database, query string, tokens []string) ([][]interface{}, error) {
	fromIndex := -1
	for i, tok := range tokens {
		if strings.ToUpper(tok) == "FROM" {
			fromIndex = i
			break
		}
	}
	if fromIndex == -1 {
		return nil, errors.New("неверный синтаксис SELECT: отсутствует FROM")
	}
	if fromIndex < 2 {
		return nil, errors.New("неверный синтаксис SELECT: отсутствуют столбцы для выборки")
	}

	columnsList := strings.Join(tokens[1:fromIndex], " ")
	columnsList = strings.TrimSpace(columnsList)
	columnsList = strings.Trim(columnsList, ",")
	selectColumns := splitCSV(columnsList)
	for i := range selectColumns {
		selectColumns[i] = strings.TrimSpace(selectColumns[i])
		selectColumns[i] = strings.Trim(selectColumns[i], "'")
	}

	if fromIndex+1 >= len(tokens) {
		return nil, errors.New("неверный синтаксис SELECT: отсутствует имя таблицы после FROM")
	}
	tableName := tokens[fromIndex+1]

	joinType, joinTable, joinCondition, err := parseJoin(tokens)
	if err != nil {
		return nil, err
	}

	whereIndex := -1
	for i, tok := range tokens {
		if strings.ToUpper(tok) == "WHERE" {
			whereIndex = i
			break
		}
	}

	var condition *Condition
	if whereIndex != -1 {
		if whereIndex+1 >= len(tokens) {
			return nil, errors.New("неверный синтаксис WHERE: отсутствует условие")
		}
		condition, err = parseWhere(tokens, whereIndex+1)
		if err != nil {
			return nil, err
		}
	}

	var joinedData [][]interface{}
	var columnNames []string

	if joinType != "" {
		onParts := strings.Split(joinCondition, "=")
		if len(onParts) != 2 {
			return nil, errors.New("неверный синтаксис условия JOIN")
		}
		joinColumn1 := strings.TrimSpace(onParts[0])
		joinColumn2 := strings.TrimSpace(onParts[1])

		if strings.Contains(joinColumn1, ".") {
			parts := strings.Split(joinColumn1, ".")
			if len(parts) != 2 {
				return nil, errors.New("неверный синтаксис столбца для JOIN")
			}
			joinColumn1 = parts[1]
		}
		if strings.Contains(joinColumn2, ".") {
			parts := strings.Split(joinColumn2, ".")
			if len(parts) != 2 {
				return nil, errors.New("неверный синтаксис столбца для JOIN")
			}
			joinColumn2 = parts[1]
		}

		switch joinType {
		case "LEFT":
			joinedData, err = db.LeftJoin(tableName, joinTable, joinColumn1, joinColumn2)
		case "RIGHT":
			joinedData, err = db.RightJoin(tableName, joinTable, joinColumn1, joinColumn2)
		case "INNER":
			joinedData, err = db.Join(tableName, joinTable, joinColumn1, joinColumn2)
		default:
			joinedData, err = db.Join(tableName, joinTable, joinColumn1, joinColumn2)
		}
		if err != nil {
			return nil, err
		}

		table1, exists1 := db.Tables[strings.ToLower(tableName)]
		table2, exists2 := db.Tables[strings.ToLower(joinTable)]
		if !exists1 || !exists2 {
			return nil, errors.New("одна или обе таблицы не существуют")
		}
		for _, col := range table1.Columns {
			columnNames = append(columnNames, fmt.Sprintf("%s.%s", table1.Name, col.Name))
		}
		for _, col := range table2.Columns {
			columnNames = append(columnNames, fmt.Sprintf("%s.%s", table2.Name, col.Name))
		}

		if condition != nil {
			filteredData := [][]interface{}{}
			for _, row := range joinedData {
				match, err := evaluateJoinedCondition(row, columnNames, condition)
				if err != nil {
					return nil, err
				}
				if match {
					filteredData = append(filteredData, row)
				}
			}
			joinedData = filteredData
		}

	} else {
		rows, err := db.Select(tableName, nil)
		if err != nil {
			return nil, err
		}
		joinedData = rows

		table, exists := db.Tables[strings.ToLower(tableName)]
		if !exists {
			return nil, fmt.Errorf("таблица '%s' не существует", tableName)
		}
		for _, col := range table.Columns {
			columnNames = append(columnNames, col.Name)
		}

		if condition != nil {
			filteredData := [][]interface{}{}
			for _, row := range joinedData {
				match, err := evaluateCondition(row, columnNames, condition)
				if err != nil {
					return nil, err
				}
				if match {
					filteredData = append(filteredData, row)
				}
			}
			joinedData = filteredData
		}
	}

	if len(selectColumns) > 0 && selectColumns[0] != "*" {
		var selectedIndexes []int
		allColumns := columnNames

		for _, col := range selectColumns {
			index := -1
			for i, ac := range allColumns {
				if strings.ToLower(ac) == strings.ToLower(col) || (strings.Contains(ac, ".") && strings.ToLower(ac[strings.LastIndex(ac, ".")+1:]) == strings.ToLower(col)) {
					index = i
					break
				}
			}
			if index == -1 {
				return nil, fmt.Errorf("столбец '%s' не найден в результате", col)
			}
			selectedIndexes = append(selectedIndexes, index)
		}

		var finalResult [][]interface{}
		for _, row := range joinedData {
			var newRow []interface{}
			for _, idx := range selectedIndexes {
				if idx < len(row) {
					newRow = append(newRow, row[idx])
				} else {
					newRow = append(newRow, nil)
				}
			}
			finalResult = append(finalResult, newRow)
		}
		return finalResult, nil
	}

	return joinedData, nil
}

func handleUpdate(db *Database, query string, tokens []string) ([][]interface{}, error) {
	if len(tokens) < 4 || strings.ToUpper(tokens[2]) != "SET" {
		return nil, errors.New("неверный синтаксис UPDATE")
	}
	tableName := tokens[1]

	whereIndex := -1
	for i, tok := range tokens {
		if strings.ToUpper(tok) == "WHERE" {
			whereIndex = i
			break
		}
	}

	setPart := ""
	if whereIndex == -1 {
		setPart = strings.Join(tokens[3:], " ")
	} else {
		setPart = strings.Join(tokens[3:whereIndex], " ")
	}

	setParts := splitCSV(setPart)
	if len(setParts) != 1 {
		return nil, errors.New("поддерживается только одно обновление столбца за раз")
	}

	setTokens := strings.SplitN(setParts[0], "=", 2)
	if len(setTokens) != 2 {
		return nil, errors.New("неверный синтаксис SET")
	}
	columnName := strings.TrimSpace(setTokens[0])
	newValue := strings.TrimSpace(setTokens[1])
	newValue = strings.Trim(newValue, "'")

	var condition *Condition
	if whereIndex != -1 {
		condition, _ = parseWhere(tokens, whereIndex+1)
	}

	err := db.Update(tableName, columnName, newValue, condition)
	if err != nil {
		return nil, err
	}
	fmt.Println("Данные обновлены успешно.")
	return nil, nil
}

func handleDelete(db *Database, query string, tokens []string) ([][]interface{}, error) {
	if len(tokens) < 3 || strings.ToUpper(tokens[1]) != "FROM" {
		return nil, errors.New("неверный синтаксис DELETE")
	}
	tableName := tokens[2]

	whereIndex := -1
	for i, tok := range tokens {
		if strings.ToUpper(tok) == "WHERE" {
			whereIndex = i
			break
		}
	}

	var condition *Condition
	if whereIndex != -1 {
		condition, _ = parseWhere(tokens, whereIndex+1)
	}

	err := db.Delete(tableName, condition)
	if err != nil {
		return nil, err
	}
	fmt.Println("Данные удалены успешно.")
	return nil, nil
}

func parseJoin(tokens []string) (joinType string, joinTable string, joinCondition string, err error) {
	for i := 0; i < len(tokens); i++ {
		upperTok := strings.ToUpper(tokens[i])
		if upperTok == "LEFT" || upperTok == "RIGHT" {
			if i+1 < len(tokens) && strings.ToUpper(tokens[i+1]) == "JOIN" {
				joinType = upperTok
				if i+2 >= len(tokens) {
					return "", "", "", errors.New("неверный синтаксис JOIN: отсутствует имя таблицы")
				}
				joinTable = tokens[i+2]
				onIndex := -1
				for j := i + 3; j < len(tokens); j++ {
					if strings.ToUpper(tokens[j]) == "ON" {
						onIndex = j
						break
					}
				}
				if onIndex == -1 {
					return "", "", "", errors.New("не найдено условие ON для JOIN")
				}
				endIndex := len(tokens)
				for j := onIndex + 1; j < len(tokens); j++ {
					if strings.ToUpper(tokens[j]) == "WHERE" {
						endIndex = j
						break
					}
				}
				joinCondition = strings.Join(tokens[onIndex+1:endIndex], " ")
				return joinType, joinTable, joinCondition, nil
			}
		} else if upperTok == "JOIN" {
			joinType = "INNER"
			if i+1 >= len(tokens) {
				return "", "", "", errors.New("неверный синтаксис JOIN: отсутствует имя таблицы")
			}
			joinTable = tokens[i+1]
			onIndex := -1
			for j := i + 2; j < len(tokens); j++ {
				if strings.ToUpper(tokens[j]) == "ON" {
					onIndex = j
					break
				}
			}
			if onIndex == -1 {
				return "", "", "", errors.New("не найдено условие ON для JOIN")
			}
			endIndex := len(tokens)
			for j := onIndex + 1; j < len(tokens); j++ {
				if strings.ToUpper(tokens[j]) == "WHERE" {
					endIndex = j
					break
				}
			}
			joinCondition = strings.Join(tokens[onIndex+1:endIndex], " ")
			return joinType, joinTable, joinCondition, nil
		}
	}
	return "", "", "", nil
}

func parseWhere(tokens []string, startIndex int) (*Condition, error) {
	cond, nextIndex, err := parseCondition(tokens, startIndex, len(tokens))
	if err != nil {
		return nil, err
	}
	if nextIndex != len(tokens) && tokens[nextIndex] != ";" {
		return nil, fmt.Errorf("неверный синтаксис WHERE: неожиданный токен '%s'", tokens[nextIndex])
	}
	return cond, nil
}

func parseCondition(tokens []string, start, end int) (*Condition, int, error) {
	var left *Condition
	current := start

	for current < end {
		token := tokens[current]

		if token == "(" {
			subCond, nextIndex, err := parseCondition(tokens, current+1, end)
			if err != nil {
				return nil, current, err
			}
			left = subCond
			current = nextIndex
		} else if token == ")" {
			current++
			return left, current, nil
		} else if strings.ToUpper(token) == "AND" || strings.ToUpper(token) == "OR" {
			if left == nil {
				return nil, current, fmt.Errorf("неверный синтаксис WHERE: логический оператор '%s' без левого условия", token)
			}
			logicalOp := strings.ToUpper(token)
			current++
			rightCond, nextIndex, err := parseCondition(tokens, current, end)
			if err != nil {
				return nil, current, err
			}
			left = &Condition{
				Type:      Compound,
				Left:      left,
				Right:     rightCond,
				LogicalOp: logicalOp,
			}
			current = nextIndex
		} else {
			if current+2 >= end {
				return nil, current, errors.New("неверный синтаксис WHERE: недостаточно токенов для условия")
			}
			column := token
			operator := tokens[current+1]
			valueToken := tokens[current+2]
			current += 3

			var value interface{}
			if strings.HasPrefix(valueToken, "'") && strings.HasSuffix(valueToken, "'") {
				value = strings.Trim(valueToken, "'")
			} else {
				if intVal, err := strconv.Atoi(valueToken); err == nil {
					value = intVal
				} else if floatVal, err := strconv.ParseFloat(valueToken, 64); err == nil {
					value = floatVal
				} else {
					return nil, current, fmt.Errorf("неизвестный тип значения '%s' в WHERE", valueToken)
				}
			}

			cond := &Condition{
				Type:     Simple,
				Column:   column,
				Operator: operator,
				Value:    value,
			}

			if left == nil {
				left = cond
			} else {
				return nil, current, errors.New("неверный синтаксис WHERE: отсутствует логический оператор между условиями")
			}
		}

		if current >= end {
			break
		}

		token = tokens[current]
		upperToken := strings.ToUpper(token)
		if upperToken == "AND" || upperToken == "OR" {
			continue
		} else if token == ")" {
			current++
			return left, current, nil
		} else {
			break
		}
	}

	return left, current, nil
}

func tokenize(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune

	operatorChars := "=!<>"

	for i := 0; i < len(query); i++ {
		r := rune(query[i])
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		case r == '\'' || r == '"':
			current.WriteRune(r)
			if inQuotes && r == quoteChar {
				inQuotes = false
			} else if !inQuotes {
				inQuotes = true
				quoteChar = r
			}
		case strings.ContainsRune(operatorChars, r):
			if inQuotes {
				current.WriteRune(r)
			} else {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				if i+1 < len(query) {
					nextR := rune(query[i+1])
					op := string(r) + string(nextR)
					if op == "<=" || op == ">=" || op == "!=" || op == "<>" {
						tokens = append(tokens, op)
						i++
						continue
					}
				}
				tokens = append(tokens, string(r))
			}
		case r == '(', r == ')', r == ',', r == ';':
			if inQuotes {
				current.WriteRune(r)
			} else {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				tokens = append(tokens, string(r))
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func splitCSV(input string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, r := range input {
		switch r {
		case ',':
			if !inQuotes {
				result = append(result, current.String())
				current.Reset()
				continue
			}
		case '\'':
			inQuotes = !inQuotes
		}
		current.WriteRune(r)
	}
	result = append(result, current.String())
	return result
}

func evaluateJoinedCondition(row []interface{}, columnNames []string, condition *Condition) (bool, error) {
	return evaluateCondition(row, columnNames, condition)
}

func evaluateCondition(row []interface{}, columnNames []string, condition *Condition) (bool, error) {
	if condition.Type == Simple {
		var colIndex int = -1
		for i, col := range columnNames {
			if strings.ToLower(col) == strings.ToLower(condition.Column) || (strings.Contains(col, ".") && strings.ToLower(col[strings.LastIndex(col, ".")+1:]) == strings.ToLower(condition.Column)) {
				colIndex = i
				break
			}
		}

		if colIndex == -1 {
			return false, fmt.Errorf("столбец '%s' не найден в результате", condition.Column)
		}

		value := row[colIndex]

		return evaluateSimpleCondition(value, condition.Operator, condition.Value)
	} else if condition.Type == Compound {
		leftResult, err := evaluateCondition(row, columnNames, condition.Left)
		if err != nil {
			return false, err
		}

		rightResult, err := evaluateCondition(row, columnNames, condition.Right)
		if err != nil {
			return false, err
		}

		switch condition.LogicalOp {
		case "AND":
			return leftResult && rightResult, nil
		case "OR":
			return leftResult || rightResult, nil
		default:
			return false, fmt.Errorf("неизвестный логический оператор '%s'", condition.LogicalOp)
		}
	}

	return false, errors.New("неизвестный тип условия")
}

func evaluateSimpleCondition(value interface{}, operator string, target interface{}) (bool, error) {
	switch v := value.(type) {
	case int:
		targetVal, ok := target.(int)
		if !ok {
			return false, fmt.Errorf("несоответствие типов: сравнение INTEGER с %T", target)
		}
		switch operator {
		case "=":
			return v == targetVal, nil
		case "!=":
			return v != targetVal, nil
		case "<":
			return v < targetVal, nil
		case ">":
			return v > targetVal, nil
		case "<=":
			return v <= targetVal, nil
		case ">=":
			return v >= targetVal, nil
		default:
			return false, fmt.Errorf("неподдерживаемый оператор '%s'", operator)
		}
	case float64:
		var targetVal float64
		switch tv := target.(type) {
		case float64:
			targetVal = tv
		case int:
			targetVal = float64(tv)
		default:
			return false, fmt.Errorf("несоответствие типов: сравнение FLOAT с %T", target)
		}
		switch operator {
		case "=":
			return v == targetVal, nil
		case "!=":
			return v != targetVal, nil
		case "<":
			return v < targetVal, nil
		case ">":
			return v > targetVal, nil
		case "<=":
			return v <= targetVal, nil
		case ">=":
			return v >= targetVal, nil
		default:
			return false, fmt.Errorf("неподдерживаемый оператор '%s'", operator)
		}
	case string:
		targetVal, ok := target.(string)
		if !ok {
			return false, fmt.Errorf("несоответствие типов: сравнение STRING с %T", target)
		}
		switch operator {
		case "=":
			return v == targetVal, nil
		case "!=":
			return v != targetVal, nil
		default:
			return false, fmt.Errorf("неподдерживаемый оператор '%s' для строк", operator)
		}
	default:
		return false, fmt.Errorf("неподдерживаемый тип данных '%T' в WHERE", value)
	}
}
