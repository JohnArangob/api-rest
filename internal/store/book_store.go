// Package store — LA CAPA DE PERSISTENCIA (la que habla con la base de datos).
//
// Su ÚNICO trabajo: ejecutar SQL. Leer, escribir y borrar filas de la tabla
// `books`.
//
// Lo que NO hace (a propósito):
//   - No valida reglas de negocio (eso es del service).
//   - No sabe nada de HTTP, JSON ni códigos de estado (eso es del transport).
//
// Este paquete expone una INTERFAZ (`Store`) en lugar de un struct concreto.
// Eso permite que las otras capas dependan de "qué puedo hacer" (los métodos)
// y no de "cómo está hecho" (el SQL). Si mañana cambiás a PostgreSQL, solo
// reescribís este archivo y el resto del proyecto ni se entera.
package store

import (
	"api-rest/internal/model"
	"database/sql"
)

// Store es una INTERFAZ. En Go, una interfaz es una lista de "promesas":
// cualquier tipo que tenga exactamente estos 5 métodos, con estas mismas
// firmas, cumple con la interfaz y puede usarse como `store.Store`.
//
// En este caso, store.Store se implementa en este mismo archivo con el struct
// `store`. (Si otro paquete quisiera implementarla con otra base de datos,
// también podría, y el service funcionaría igual.)
//
// Traducción de cada promesa:
//
//	GetAll   → dame todos los libros.
//	GetByID  → dame el libro con ese id.
//	Create   → guardá un libro nuevo y devolvemelo con su id.
//	Update   → actualizá el libro con ese id.
//	Delete   → borrá el libro con ese id.
type Store interface {
	GetAll() ([]*model.Libro, error)
	GetByID(id int) (*model.Libro, error)
	Create(libro *model.Libro) (*model.Libro, error)
	Update(id int, libro *model.Libro) (*model.Libro, error)
	Delete(id int) error
}

// store es la implementación CONCRETA de la interfaz Store.
//
// Es un struct "no exportado" (empieza con minúscula: `store`), lo que
// significa que desde FUERA de este paquete no se puede nombrar. Para crear
// una instancia hay que usar la función New() de más abajo. Así el paquete
// controla cómo se arma su propio struct.
type store struct {
	db *sql.DB
}

// New es el constructor: recibe la conexión a la base de datos (que armó
// main.go) y devuelve algo que cumple la interfaz Store.
//
// El tipo de retorno es la INTERFAZ (Store), no el struct (store). Es una
// convención de Go: devolver siempre el tipo más genérico posible.
func New(db *sql.DB) Store {
	return &store{db: db}
}

// GetAll trae TODOS los libros de la tabla.
//
// El SQL se escribe con ` (backticks, comillas invertidas) para que los
// saltos de línea y comillas del texto no rompan el string. En Go los strings
// "normales" viven entre comillas dobles y usan \ para escapes; con backticks
// se toma el contenido literal, que es lo que queremos para SQL multilinea.
//
//	`SELECT id, title, author FROM books`
//	= "dame esas tres columnas de todos los libros".
//
// s.db.Query(q) devuelve DOS valores:
//   - rows: un CURSOR (puntero) a las filas que vienen de la base. Todavía
//     NO es la lista final: hay que ir leyendo fila por fila con rows.Next().
//   - err:  si la consulta falló (tabla inexistente, sintaxis, etc.).
func (s *store) GetAll() ([]*model.Libro, error) {
	q := `SELECT id, title, author FROM books`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}

	// rows.Close() libera los recursos que ocupa la conexión abierta por la
	// consulta. Usamos `defer` porque podríamos salir de la función antes de
	// terminar de recorrer (por ejemplo, si falla un Scan). Así se cierra
	// SIEMPRE, pase lo que pase.
	defer rows.Close()

	// Este slice (lista de punteros) va acumulando los libros encontrados.
	// Arranca vacío (se declara con var y no se le pone valor).
	var libros []*model.Libro

	// rows.Next() mueve el cursor al SIGUIENTE renglón y devuelve true mientras
	// haya. La primera vez posiciona en la primera fila; cuando no queden más,
	// devuelve false y el bucle termina.
	for rows.Next() {
		// Creamos un libro vacío donde ir cargando los datos de esta fila.
		b := model.Libro{}

		// rows.Scan(&...) COPIA los valores de las columnas de la fila actual
		// dentro de las variables indicadas.
		//
		// El ORDEN importa: tiene que coincidir exactamente con el orden del
		// SELECT (id, title, author), no con el orden de los campos del struct.
		//
		// Los `&` son "direcciones de memoria". Scan necesita poder ESCRIBIR
		// dentro de esas variables, y para eso se le pasa su puntero. Sin el
		// &, estarías pasando una copia y la escritura no llegaría a `b`.
		if err := rows.Scan(&b.ID, &b.Titulo, &b.Autor); err != nil {
			return nil, err
		}

		// Guardamos el PUNTERO al struct recién leído (&b = dirección de b).
		// Se usa puntero para que cada vuelta del bucle no tenga que copiar el
		// struct completo, y para poder modificarlo luego si hiciera falta.
		libros = append(libros, &b)
	}

	// rows.Err() chequea si hubo algún error DURANTE la iteración (por ejemplo
	// que se cortó la conexión a mitad de camino). Por eso se revisa después
	// del bucle, no dentro.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Devolvemos la lista y nil (nil = "sin error").
	return libros, nil
}

// GetByID busca UN solo libro por su id.
//
// Cambio respecto del anterior: acá se usa QueryRow en vez de Query, porque
// esta consulta devuelve COMO MÁXIMO UNA fila. QueryRow devuelve directamente
// un objeto al que se le aplica .Scan(...) para llenar el struct.
//
// Si no existe ninguna fila con ese id, Scan devuelve un error conocido
// (sql.ErrNoRows). Ese error llega hasta el handler, que lo traduce en un
// 404 "No encontrado".
func (s *store) GetByID(id int) (*model.Libro, error) {
	// El `?` es un "placeholder": el valor real se pasa aparte como argumento.
	//
	// ¿Por qué no escribir el id directamente dentro del string? Porque eso
	// abriría la puerta a la INYECCIÓN SQL: si el valor viniera de texto del
	// usuario y se concatenara, alguien podría mandar algo como
	// "1; DROP TABLE books" y borrar la tabla. Con `?`, la base de datos
	// interpreta SIEMPRE el valor como dato, jamás como parte del comando.
	q := `SELECT id, title, author FROM books WHERE id = ?`

	b := model.Libro{}

	err := s.db.QueryRow(q, id).Scan(&b.ID, &b.Titulo, &b.Autor)
	if err != nil {
		return nil, err
	}

	// Devolvemos &b (puntero) para que el llamador reciba un *model.Libro,
	// que es lo que promete la firma de la función.
	return &b, nil
}

// Create inserta un libro nuevo en la tabla.
//
// `resp` es el resultado de la operación (tipo sql.Result). No es el libro:
// es un objeto que sirve para preguntar cosas como "cuántas filas afectó" o
// "qué id se generó".
func (s *store) Create(libro *model.Libro) (*model.Libro, error) {
	// IMPORTANTE: no mandamos el id en el INSERT. Lo genera SQLite sola gracias
	// a AUTOINCREMENT. Por eso solo aparecen title y author.
	q := `INSERT INTO books (title, author) VALUES (?, ?)`

	// db.Exec sirve para SQL que no devuelve filas (INSERT, UPDATE, DELETE).
	resp, err := s.db.Exec(q, libro.Titulo, libro.Autor)
	if err != nil {
		return nil, err
	}

	// LastInsertId devuelve el id que se acaba de crear (por eso nos sirve
	// resp). Así, en la respuesta del POST podemos devolver el libro con su
	// id recién asignado, en vez de devolverlo sin id.
	id, err := resp.LastInsertId()

	if err != nil {
		return nil, err
	}

	// Guardamos ese id dentro del struct. int(id) convierte de int64 (el tipo
	// que devuelve la BD) a int (el tipo del campo ID).
	libro.ID = int(id)

	// Devolvemos el MISMO puntero que entró, pero ahora "relleno" con el id.
	return libro, nil
}

// Update modifica title y author del libro cuyo id coincida.
//
// Detalle clave: el `?` del WHERE lleva `id` (el parámetro de la función),
// que es el que viene de la URL (/books/1). NO se usa libro.ID, porque libro
// es lo que mandó el cliente en el body, y el cliente podría mandar un id
// distinto o ninguno. El que manda sobre qué fila modificar es la URL.
func (s *store) Update(id int, libro *model.Libro) (*model.Libro, error) {
	// "UPDATE books SET title = ?, author = ? WHERE id = ?"
	// = "en la tabla books, cambia title y author donde el id sea este".
	q := `UPDATE books SET title = ?, author = ? WHERE id = ?`

	// Usamos db.Exec (no Query) porque un UPDATE no devuelve filas.
	// El primer valor lo ignoramos con `_`.
	_, err := s.db.Exec(q, libro.Titulo, libro.Autor, id)
	if err != nil {
		return nil, err
	}

	// Ajustamos el id del struct devuelto para que refleje el id real que se
	// actualizó. Esto ocurre DESPUÉS del UPDATE, así que solo afecta a la
	// respuesta que ve el cliente (el id en la BD no se toca).
	libro.ID = id

	// Devolvemos el libro con los datos nuevos.
	// OJO: si acá devolvieras `return nil, nil` (sin error), el handler
	// codificaría ese nil con json.Encode y el cliente vería literalmente
	// la palabra `null` como respuesta — que era exactamente el error que
	// tenías con tu curl. El "null" nunca salía de la URL ni del JSON que
	// mandabas: salía de acá.
	return libro, nil
}

// Delete borra el libro con el id indicado.
//
// Devuelve SOLO un error, porque no hay nada que devolver: el libro ya no
// existe después de la operación (por eso el handler responde 204 y vacío).
func (s *store) Delete(id int) error {
	// OJO: este SQL está escrito como `DELETE from books WHERE id = ?`
	// (con comilla simple y 'from' en minúsculas). SQL no distingue
	// mayúsculas/minúsculas en las palabras clave, así que funciona igual.
	q := `DELETE from books WHERE id = ?`

	// El `id` es el segundo argumento de Exec: rellena el `?` del WHERE.
	// Siempre con placeholder, nunca concatenado (ver nota de inyección SQL
	// en GetByID).
	_, err := s.db.Exec(q, id)

	if err != nil {
		return err
	}

	// `nil` como primer (y único) valor de retorno = "no hubo error".
	// El primer valor de Exec (las filas afectadas) lo ignoramos con `_`.
	//
	//TODO: Hacer esto
	//
	// Idea para cuando avances: chequear RowsAffected() para poder
	// distinguir "se borró el libro" de "no existía ese id". Hoy un DELETE
	// sobre un id inexistente también devuelve nil, y el handler respondería
	// 204 igual aunque no borrara nada.
	return nil
}
