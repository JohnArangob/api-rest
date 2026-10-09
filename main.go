// =============================================================================
// main.go  —  PUNTO DE ENTRADA DE LA API
// =============================================================================
//
// Este archivo es el "director de orquesta" del proyecto. Cuando corrés
// `go run .`, Go entra por acá.
//
// Sus responsabilidades, en orden:
//
//  1. Conectar con la base de datos (SQLite).
//  2. Crear la tabla `books` si todavía no existe.
//  3. Crear las tres capas del proyecto y conectarlas entre sí.
//  4. Definir las rutas (qué URL llama a qué función).
//  5. Encender el servidor HTTP y quedarse escuchando en el puerto 8080.
//
// Si querés entender "por qué está cada cosa acá", este es el archivo que
// tenés que leer primero: es el único que arma todo el sistema.
package main

import (
	"api-rest/internal/service"
	"api-rest/internal/store"
	"api-rest/internal/transport"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// El guion bajo `_` delante del import es un "import anónimo": no usamos
	// ningún símbolo de ese paquete en este archivo, pero lo importamos por su
	// EFECTO SECUNDARIO. Al cargarse, el driver de SQLite se registra solo
	// dentro de `database/sql`, permitiendo que sql.Open("sqlite3", ...)
	// reconozca ese nombre. Sin esta línea, sql.Open devolvería un error
	// diciendo que no conoce el driver "sqlite3".
	_ "github.com/mattn/go-sqlite3"
)

// main es la función que Go ejecuta al arrancar el programa.
// Go obliga a que se llame exactamente "main" y que esté en el paquete "main".
func main() {
	// -------------------------------------------------------------------------
	// PASO 1: conectar con SQLite
	// -------------------------------------------------------------------------
	//
	// sql.Open("sqlite3", "./books.db") NO abre la base de datos en ese
	// instante: solo arma el "puente" (el objeto `db`). La conexión real se
	// hace de forma perezosa, cuando se usa por primera vez.
	//
	//   - "sqlite3"     → nombre que registró el driver importado arriba.
	//   - "./books.db"  → archivo físico que guarda todos los datos. Si no
	//     existe, SQLite lo crea solito en esta carpeta. Por eso corre sin
	//     instalar ningún servidor de base de datos.
	//
	// En Go, las operaciones que pueden fallar devuelven DOS valores: el
	// resultado y un `error`. Ese `error` es `nil` cuando todo salió bien.
	//Conectar SQLLite
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		// log.Fatal = imprimir el error por consola y TERMINAR el programa
		// (sale con código de salida 1). Si no podemos abrir la base de datos,
		// el resto del programa no tendría sentido seguir ejecutándose.
		log.Fatal(err)
	}
	// `defer` significa: "guardá esta línea y ejecutala al FINAL de la
	// función", sin importar si salimos por un return o por un error.
	// Así garantizamos que la conexión a la BD se cierre siempre.
	defer db.Close()

	// -------------------------------------------------------------------------
	// PASO 2: crear la tabla si no existe
	// -------------------------------------------------------------------------
	//
	// Este script SQL es "idempotente": se puede ejecutar muchas veces sin
	// problema, gracias a `IF NOT EXISTS`. Si la tabla ya existe, no hace nada.
	//
	// Significado de cada parte:
	//
	//   id INTEGER PRIMARY KEY AUTOINCREMENT
	//       → número único que identifica a cada libro. PRIMARY KEY = es la
	//         clave única de la tabla; AUTOINCREMENT = SQLite la va dando
	//         sola (1, 2, 3...).
	//   title TEXT NOT NULL
	//       → texto obligatorio: no se permite que esté vacío (NULL).
	//   author TEXT NOT NULL
	//       → texto obligatorio: no se permite que esté vacío (NULL).
	//
	//Crear table si no existe

	q := `
		CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT NOT NULL
		)
	`
	if _, err := db.Exec(q); err != nil {
		// db.Exec ejecuta un SQL que NO devuelve filas (CREATE, INSERT,
		// UPDATE, DELETE). El primer valor de retorno lo ignoramos con `_`
		// porque es un "resultado" con filas afectadas, id generado, etc.
		log.Fatal(err.Error())
	}

	// -------------------------------------------------------------------------
	// PASO 3: inyectar las dependencias ("armar el equipo")
	// -------------------------------------------------------------------------
	//
	// "Dependencia" = una pieza que necesita a otra para funcionar.
	// Este proyecto está dividido en 3 capas que se apoyan así:
	//
	//    transport (HTTP)  →  service (reglas)  →  store (base de datos)
	//
	// Acá creamos cada capa de afuera hacia adentro y le "enchufamos" la que
	// necesita. A esto se le llama INYECCIÓN DE DEPENDENCIAS: en vez de que
	// cada pieza cree sus propias dependencias por dentro, se las pasamos
	// desde afuera.
	//
	// ¿Para qué sirve en la práctica? Para poder cambiar una pieza sin tocar
	// las demás. Hoy le pasamos una base SQLite; mañana podrías pasarle otra
	// base de datos y TODO el código de las otras capas seguiría igual.
	//
	//Inyectar nuestras dependencias

	// 1) la capa de datos recibe la conexión a la base de datos
	bookStore := store.New(db)
	// 2) la capa de reglas recibe la de datos
	bookService := service.New(bookStore)
	// 3) la capa HTTP recibe la de reglas
	bookHandler := transport.New(bookService)

	// -------------------------------------------------------------------------
	// PASO 4: configurar las rutas (URLs)
	// -------------------------------------------------------------------------
	//
	// http.HandleFunc(ruta, función) = "cuando llegue una petición a esta
	// ruta, llama a esta función".
	//
	// Go distingue DOS rutas que parecen casi iguales:
	//
	//   "/books"   → exactamente /books          (traer todo, crear)
	//   "/books/"  → /books/1, /books/2, /books/9 (un libro concreto)
	//
	// Por eso el segundo tiene la barra al final: así captura el número de id
	// que viene después. Si estuvieran al revés, /books/1 no llegaría a
	// ningún handler y el servidor devolvería 404.
	//
	//COnfigurar rutas
	http.HandleFunc("/books", bookHandler.HandleBooks)
	http.HandleFunc("/books/", bookHandler.HandleBookByID)

	// Mensajes informativos para la consola. No forman parte de la API:
	// son ayuda para que al arrancar veas qué puede hacer el servidor.
	fmt.Println("Servidor ejecutandose en http:localhots:8080")
	fmt.Println("API Endpoints:")
	fmt.Println("Get	/books			-Obtener todos los libros")
	fmt.Println("POST	/books			-Crear un nuevo libro")
	fmt.Println("GET	/books/{id}		-Obtener un libro en especifico")
	fmt.Println("PUT	/books/{id}		-Actualizar un libro")
	fmt.Println("DELETE	/books/{id}		-Eliminar un libro")

	// -------------------------------------------------------------------------
	// PASO 5: encender el servidor
	// -------------------------------------------------------------------------
	//
	// http.ListenAndServe(":8080", nil) = "quedate escuchando peticiones en
	// el puerto 8080".
	//
	//   - ":8080" → en qué puerto escuchar. Por eso el curl usa
	//     http://localhost:8080 (localhost = tu propia máquina).
	//   - `nil` como segundo argumento → "usá el mux (repartidor de rutas)
	//     por defecto", que es el mismo donde registramos las rutas con
	//     HandleFunc más arriba.
	//
	// Esta línea BLOQUEA el programa: main() no termina nunca mientras el
	// servidor esté vivo (para eso está: tiene que quedarse escuchando).
	// Si falla (puerto ocupado, sin permisos...), devuelve un error y
	// log.Fatal lo imprime y detiene el proceso.
	//
	//Empezar y escuchar el servidor

	log.Fatal(http.ListenAndServe(":8080", nil))
}
