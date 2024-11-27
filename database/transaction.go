package database

type Transaction struct {
	operations []Operation
}

type Operation struct {
	Type       string
	TableName  string
	Data       interface{}
	RowIndex   int
	ColumnName string
	NewValue   interface{}
}
