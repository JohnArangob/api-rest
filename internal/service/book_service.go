// Package service — LA CAPA DE REGLAS DE NEGOCIO (lógica de negocio).
//
// ¿Qué hace? Se sienta EN EL MEDIO entre el mundo HTTP (transport) y el
// mundo de la base de datos (store):
//
//	transport  →  service  →  store
//	(HTTP)        (reglas)    (SQL)
//
// Aquí viven las decisiones de TU aplicación que no dependen ni de HTTP ni de
// SQL. Ejemplo: "un libro sin título no se acepta". Esa regla es de la
// aplicación, no del navegador ni de la base de datos, por eso vive acá.
//
// ¿Por qué separarlo en su propia capa? Porque si mezclaras las reglas con el
// SQL, cambiar la base de datos te obligaría a reescribir las reglas, y
// viceversa. Separándolas, cada parte cambia sin romper a la otra.
package service

import (
	"api-rest/internal/model"
	"api-rest/internal/store"
	"errors"
)

// Service es el "servicio" de libros. Solo guarda una referencia al store
// (la capa de datos) que va a usar para trabajar.
type Service struct {
	store store.Store
}

// New es el constructor del Service: recibe el store que necesita y devuelve
// el Service ya listo para usar. Este es el argumento que se "inyecta" desde
// main.go (PASO 3).
func New(s store.Store) *Service {
	return &Service{
		store: s,
	}

}

// (s *Service) que aparece antes de cada función es el "receptor" (receiver):
// es como el "this" de otros lenguajes. Significa que esta función es un
// MÉTODO del Service, y dentro de ella `s` es el propio Service.

// ObtenTodosLosLIbros le pide al store TODOS los libros y los devuelve.
//
// Leer su firma es importante para entender Go:
//
//	([]*model.Libro, error)
//
//	  - []*model.Libro → un "slice" (lista) de punteros a Libro. Cada
//	    elemento es un libro de la base de datos. Se usa puntero
//	    (*model.Libro) para no copiar el struct completo en memoria.
//	  - error → si algo falló, acá viene el motivo; si todo salió bien es nil.
//
// PATRÓN CLÁSICO DE GO: si hay error, devolvemos (nil, err) y listo. Quien
// llamó (el handler) es el que decide qué responder al cliente. Go no usa
// excepciones: los errores son valores que se pasan por la vuelta.
func (s *Service) ObtenTodosLosLIbros() ([]*model.Libro, error) {

	libros, err := s.store.GetAll()
	if err != nil {
		// Acá NO imprimimos ni respondemos nada: solo propagamos el error
		// hacia arriba, para que el handler lo traduzca en un código HTTP.
		return nil, err
	}
	return libros, nil
}

// ObtenLibrosPorID busca UN libro por su id.
//
// Es una función "delgada": no tiene reglas que validar, así que simplemente
// delega el pedido en el store. Igual existe como método propio del service
// porque es la puerta de entrada que usa el handler.
func (s *Service) ObtenLibrosPorID(id int) (*model.Libro, error) {
	return s.store.GetByID(id)
}

// CrearLibro recibe un libro llegado desde el body JSON y le pide al store
// que lo guarde en la base de datos.
//
// Nota sobre valores vs punteros: `libro` se pasa por VALOR, es decir que
// service trabaja con una COPIA de lo que mandó el cliente. Después se manda
// &libro (la dirección de esa copia) al store, que es quien la modifica
// (le agunta el id) y la devuelve.
func (s *Service) CrearLibro(libro model.Libro) (*model.Libro, error) {

	return s.store.Create(&libro)
}

// UpdateAlLibro actualiza un libro existente.
//
// ACÁ VIVE UNA REGLA DE NEGOCIO: si el título viene vacío, rechazamos la
// operación ANTES de tocar la base de datos. Ese es exactamente el trabajo
// de esta capa: validar antes de persistir. Si esta validación no existiera,
// podrías guardar libros sin título en la BD.
//
//   - id    → viene de la URL (/books/1), lo parsea el handler.
//   - libro → viene del body JSON, con los campos nuevos.
func (s *Service) UpdateAlLibro(id int, libro model.Libro) (*model.Libro, error) {
	if libro.Titulo == "" {
		// errors.New crea un error con un mensaje legible. Ese mensaje es el
		// que eventualmente verá el usuario en la respuesta del servidor.
		return nil, errors.New("Necesitamos el titulo")
	}

	return s.store.Update(id, &libro)
}

// RemoverLibro borra un libro por id. Solo devuelve un error, porque no hay
// datos que devolver: el libro ya no existirá después de la operación.
func (s *Service) RemoverLibro(id int) error {
	return s.store.Delete(id)
}
