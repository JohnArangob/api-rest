package service

import (
	"api-rest/internal/model"
	"api-rest/internal/store"
	"errors"
)

type Service struct {
	store store.Store
}

func new(s store.Store) *Service {
	return &Service{
		store: s,
	}

}

func (s *Service) ObtenTodosLosLIbros() ([]*model.Libro, error) {

	libros, err := s.store.GetAll()
	if err != nil {
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
