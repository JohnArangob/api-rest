package model

type Libro struct {
	ID     int    `json:"id"`
	Titulo string `json:"title"`
	Autor  string `json:"author"`
}
