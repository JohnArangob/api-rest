package service

import (
	"api-rest/internal/model"
	"api-rest/internal/store"
	"errors"
)

type Logger interface {
	Log(msg, error string)
}
type Service struct {
	store  store.Store
	logger Logger
}

func new(s store.Store) *Service {
	return &Service{
		store:  s,
		logger: nil,
	}

}

func (s *Service) ObtenTodosLosLIbros() ([]*model.Libro, error) {
	s.logger.Log("Estamos obteniendo los libros", "")
	libros, err := s.store.GetAll()
	if err != nil {
		s.logger.Log("El error es %v\n", err.Error())
		return nil, err
	}
	return libros, nil
}

func (s *Service) ObtenLibrosPorID(id int) (*model.Libro, error) {
	return s.store.GetByID(id)
}

func (s *Service) CrearLibro(libro model.Libro) (*model.Libro, error) {

	return s.store.Create(&libro)
}

func (s *Service) UpdateAlLibro(id int, libro model.Libro) (*model.Libro, error) {
	if libro.Titulo == "" {
		return nil, errors.New("Necesitamos el titulo")
	}

	return s.store.Update(id, &libro)
}

func (s *Service) RemoverLibro(id int) error {
	return s.store.Delete(id)
}
