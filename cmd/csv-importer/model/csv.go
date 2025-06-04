package model

type TodoCSV struct {
	TodoName string `csv:"todo_name"`
	Note     string `csv:"note"`
}
