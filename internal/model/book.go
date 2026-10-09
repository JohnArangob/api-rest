// Package model define las FORMAS de los datos que circulan por la API.
//
// Es la capa más "pelada" (delgada): no depende de HTTP ni de SQL. Solo dice
// qué campos tiene cada cosa y cómo se llaman. Por eso puede ser usada por
// todas las demás capas sin generar ciclos de dependencias (que A dependa de
// B y B de A).
//
// En una API, este tipo de archivos suele llamarse "modelo" porque es el
// modelo (molde, representación) de los datos de tu dominio.
package model

// Libro es la estructura (struct) que representa un libro en todo el sistema.
//
// Un `struct` en Go es un molde que agrupa campos con nombre, como una ficha
// de fichero:
//
//   - ID     → número de identificación
//   - Titulo → texto  (string = texto en Go)
//   - Autor  → texto
//
// Esta MISMA estructura se usa en las 3 capas:
//
//   - transport: la recibe desde el JSON que manda el usuario.
//   - service:   la valida (ej: que el título no esté vacío).
//   - store:     la guarda en la tabla `books` de la base de datos.
//
// LOS TAGS `json:"..."` (las comillas invertidas) le dicen a Go cómo se
// llama cada campo cuando se convierte a JSON y viceversa:
//
//	Campo en Go        Nombre en el JSON
//	--------           -----------------
//	ID                 "id"
//	Titulo             "title"
//	Autor              "author"
//
// Si NO pusieramos estos tags, el JSON saldría con "Titulo" y "Autor", y un
// curl como el tuyo, que manda {"title": "..."}, no podría mapearse al campo
// Titulo. Es decir: sin tags, el título llegaría vacío.
type Libro struct {
	ID     int    `json:"id"`
	Titulo string `json:"title"`
	Autor  string `json:"author"`
}
