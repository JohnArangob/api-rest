// Package transport — LA CAPA HTTP (donde entran y salen las peticiones).
//
// "Transport" = transporte. Es la FRONTERA entre el mundo exterior (curl,
// navegador, Postman, cualquier cliente) y la lógica de tu API.
//
// Su trabajo, en tres pasos:
//  1. Leer la petición: método (GET/POST/PUT/DELETE), URL y body JSON.
//  2. Traducirla en llamadas al service.
//  3. Tomar lo que devolvió el service y escribirlo de vuelta como JSON,
//     con el código de estado HTTP correspondiente.
//
// Lo que NO hace (a propósito): no valida reglas de negocio ni escribe SQL.
// Solo "traduce" entre HTTP y el service. Si acá metieras SQL, romperías la
// separación de capas y todo sería más difícil de cambiar después.
package transport

import (
	"api-rest/internal/model"
	"api-rest/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// BookHandler es la estructura que agrupa todo lo que los handlers necesitan.
// Hoy solo necesita el service: al handler no le importa cómo se guardan los
// datos, solo le pide operaciones sobre libros.
type BookHandler struct {
	service *service.Service
}

// New es el constructor: recibe el service (inyectado desde main.go en el
// PASO 3) y devuelve el handler, ya listo para registrar en las rutas.
func New(s *service.Service) *BookHandler {
	return &BookHandler{
		service: s,
	}
}

// HandleBooks atiende la ruta "/books" (SIN id): es decir, la colección
// completa de libros.
//
// ESTE es el formato de toda función manejadora (handler) en net/http:
//
//	w http.ResponseWriter → el "papel" donde escribís la respuesta que verá
//	                         el cliente: headers + código de estado + cuerpo.
//	                         Se escribe con w.Write, w.WriteHeader, etc.
//	r *http.Request        → la petición que llegó: método (r.Method),
//	                         URL (r.URL.Path), headers (r.Header) y cuerpo
//	                         (r.Body).
//
// Como esta misma URL soporta varios métodos (GET trae todo, POST crea),
// se hace un `switch` sobre r.Method para repartir. Si llega un método que
// no está contemplado, cae al `default` y responde 405.
func (h *BookHandler) HandleBooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// GET /books → traer todos los libros.
	case http.MethodGet:
		// Pedimos todos los libros al service.
		libros, err := h.service.ObtenTodosLosLIbros()
		if err != nil {
			// http.Error es un atajo: escribe el mensaje como cuerpo del
			// error y pone el código de estado. 500 = Internal Server Error,
			// es decir "falló algo de mi lado".
			//
			// El `return` es fundamental: SIEMPRE que respondas un error,
			// salí de la función. Si no, seguirías ejecutando el código de
			// abajo e intentarías escribir una segunda respuesta (lo cual
			// daría un error en runtime de Go, o una respuesta corrupta).
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ANTES de mandar el cuerpo, avisamos que lo que viene es JSON.
		// Si no setearas este header, el cliente recibiría texto plano y no
		// sabría que tiene que parsear JSON.
		w.Header().Set("Content-Type", "application/json")

		// json.NewEncoder(w).Encode(libros) convierte el slice de libros a
		// JSON y lo escribe directamente en la respuesta (en `w`).
		// Acá es donde los tags `json:"title"` del struct hacen efecto.
		json.NewEncoder(w).Encode(libros)

	// POST /books → crear un libro nuevo.
	case http.MethodPost:
		// Declaramos un Libro vacío que vamos a llenar con lo que mandó el
		// cliente.
		var libro model.Libro

		// json.NewDecoder(r.Body).Decode(&libro) LEE el cuerpo de la
		// petición y lo convierte en un struct Go.
		//
		// Se pasa &libro (puntero) porque Decode necesita ESCRIBIR dentro de
		// la variable, no modificar una copia.
		//
		// ¿Y si el body no es JSON válido o está vacío? Decode devuelve un
		// error y respondemos 400 (Bad Request) = "lo que mandaste está mal".
		if err := json.NewDecoder(r.Body).Decode(&libro); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Le pedimos al service que cree el libro (él valida las reglas).
		created, err := h.service.CrearLibro(libro)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 201 Created = "se creó un recurso nuevo". Es el código correcto
		// para un POST exitoso (mejor que el 200 que se usa en GET).
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(created)

	// Si llega PATCH, HEAD u otro método que no implementamos.
	default:
		// 405 Method Not Allowed = "esta ruta existe, pero no con ese método".
		http.Error(w, "Metodo no disponible", http.StatusMethodNotAllowed)
	}
}

// HandleBookByID atiende la ruta "/books/" CON id: /books/1, /books/2, etc.
// Aquí se concentran GET (traer uno), PUT (actualizar) y DELETE (borrar).
func (h *BookHandler) HandleBookByID(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path es la parte de la URL después del dominio, ej: "/books/1".
	// strings.TrimPrefix quita el prefijo "/books/" y deja solo "1".
	idStr := strings.TrimPrefix(r.URL.Path, "/books/")

	// strconv.Atoi convierte un string a int ("1" → 1).
	// Atoi = "ASCII to Integer". Si el texto no es un número (ej: /books/abc),
	// devuelve un error.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 400 Bad Request = "el id de la URL no es un número válido".
		http.Error(w, "No lo encontre", http.StatusBadRequest)
		// NOTA DE APRENDIZAJE (no se tocó el código a propósito): acá falta un
		// `return`. Sin él, la ejecución sigue igual y entra al switch de
		// abajo con id = 0. En la práctica, /books/abc en GET devolvería el
		// error 400 pero el código intentaría además buscar el libro id=0.
		// Sumo esto para que lo veas: es un bug clásico de olvidarse del
		// return después de http.Error.
	}

	// Nuevamente, un switch sobre el método HTTP.
	switch r.Method {
	// GET /books/{id} → traer un libro en particular.
	case http.MethodGet:
		libro, err := h.service.ObtenLibrosPorID(id)
		if err != nil {
			// Cualquier error acá lo traducimos a 404 (Not Found): si el
			// store falló al buscar, lo más probable es que ese id no exista.
			http.Error(w, "No lo encontramos", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-type", "application/json")
		// Encode convierte el struct a JSON y lo escribe en la respuesta.
		json.NewEncoder(w).Encode(libro)

	// PUT /books/{id} → reemplazar/actualizar un libro.
	//
	// PUT significa "actualiza este recurso con lo que te mando". Se diferencia
	// de POST: POST crea algo nuevo en una colección, PUT modifica algo que
	// ya existe en una URL concreta.
	case http.MethodPut:
		// Igual que en POST: leemos el body JSON y lo volcamos en un Libro.
		// Si el JSON está mal, 400.
		var libro model.Libro
		if err := json.NewDecoder(r.Body).Decode(&libro); err != nil {
			http.Error(w, "input invalido", http.StatusBadRequest)
			return
		}

		// Mandamos el id (de la URL) y el libro (del body) al service.
		// El service valida la regla de negocio (título no vacío) y delega en
		// el store.
		updated, err := h.service.UpdateAlLibro(id, libro)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Devolvemos el libro actualizado en formato JSON.
		// Si `updated` viniera como nil (sin error), json.Encode escribiría
		// literalmente `null` como cuerpo de la respuesta. Ese "null" es lo
		// que salía en tu curl cuando el store no devolvía el puntero.
		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(updated)

	// DELETE /books/{id} → borrar un libro.
	case http.MethodDelete:
		if err := h.service.RemoverLibro(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 204 No Content = "se ejecutó bien y NO tengo cuerpo que devolver".
		// Es lo correcto para un DELETE: el recurso ya no existe, no hay nada
		// que mandar de vuelta. Por eso acá no se llama a json.Encode.
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method no disponible", http.StatusMethodNotAllowed)
	}

}
