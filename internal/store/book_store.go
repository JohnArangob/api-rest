package store

import (
	"api-rest/internal/model"
	"database/sql"
)

type Store interface {
	GetAll() ([]*model.Libro, error)
	GetByID(id int) (model.Libro, error)
	Create(libro *model.Libro) (*model.Libro, error)
	Update(id int, libro *model.Libro) (model.Libro, error)
	Delete(id int) error
}

type store struct {
	db *sql.DB
}

func New(db *sql.DB) Store {
	return &store{db: db}
}

func (s *store) GetAll() ([]*model.Libro, error) {
	q := `SELECT id, title, author FROM books`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var libros []*model.Libro
	for rows.Next() {
		var b *model.Libro
		if err := rows.Scan(&b.ID, &b.Titulo, &b.Autor); err != nil {
			return nil, err
		}

		libros = append(libros, b)
	}

	return libros, nil
}

func (s *store) GetByID(id int) (*model.Libro, error) {
	q := `SELECT id, title, author FROM books WHERE id = ?`

	var b *model.Libro

	err := s.db.QueryRow(q, id).Scan(&b.ID, &b.Titulo, &b.Autor)
	if err != nil {
		return nil, err
	}

	return b, nil
}
