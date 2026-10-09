package store

import (
	"api-rest/internal/model"
	"database/sql"
)

type Store interface {
	GetAll() ([]*model.Libro, error)
	GetByID(id int) (*model.Libro, error)
	Create(libro *model.Libro) (*model.Libro, error)
	Update(id int, libro *model.Libro) (*model.Libro, error)
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
		b := model.Libro{}
		if err := rows.Scan(&b.ID, &b.Titulo, &b.Autor); err != nil {
			return nil, err
		}

		libros = append(libros, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return libros, nil
}

func (s *store) GetByID(id int) (*model.Libro, error) {
	q := `SELECT id, title, author FROM books WHERE id = ?`

	b := model.Libro{}

	err := s.db.QueryRow(q, id).Scan(&b.ID, &b.Titulo, &b.Autor)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (s *store) Create(libro *model.Libro) (*model.Libro, error) {
	q := `INSERT INTO books (title, author) VALUES (?, ?)`
	resp, err := s.db.Exec(q, libro.Titulo, libro.Autor)
	if err != nil {
		return nil, err
	}

	id, err := resp.LastInsertId()

	if err != nil {
		return nil, err
	}

	libro.ID = int(id)
	return libro, nil
}

func (s *store) Update(id int, libro *model.Libro) (*model.Libro, error) {
	q := `UPDATE books SET title = ?, author = ? WHERE id = ?`
	_, err := s.db.Exec(q, libro.Titulo, libro.Autor, id)
	if err != nil {
		return nil, err
	}

	libro.ID = id
	return libro, nil
}

func (s *store) Delete(id int) error {
	q := `DELETE from books WHERE id = ?`
	_, err := s.db.Exec(q, id)

	if err != nil {
		return err
	}

	//TODO: Hacer esto
	return nil
}
