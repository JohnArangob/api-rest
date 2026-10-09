package main

import (
	"api-rest/internal/service"
	"api-rest/internal/store"
	"api-rest/internal/transport"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	//Conectar SQLLite
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//Crear table si no existe

	q := `
		CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT NOT NULL
		)
	`
	if _, err := db.Exec(q); err != nil {
		log.Fatal(err.Error())
	}

	//Inyectar nuestras dependencias
	bookStore := store.New(db)
	bookService := service.New(bookStore)
	bookHandler := transport.New(bookService)

	//COnfigurar rutas
	http.HandleFunc("/books", bookHandler.HandleBooks)
	http.HandleFunc("/books/", bookHandler.HandleBookByID)

	fmt.Println("Servidor ejecutandose en http:localhots:8080")
	fmt.Println("API Endpoints:")
	fmt.Println("Get	/books			-Obtener todos los libros")
	fmt.Println("POST	/books			-Crear un nuevo libro")
	fmt.Println("GET	/books/{id}		-Obtener un libro en especifico")
	fmt.Println("PUT	/books/{id}		-Actualizar un libro")
	fmt.Println("DELETE	/books/{id}		-Eliminar un libro")

	//Empezar y escuchar el servidor

	log.Fatal(http.ListenAndServe(":8080", nil))
}
