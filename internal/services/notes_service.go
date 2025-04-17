package services

import (
	"database/sql"
	"errors"
	"fmt"
	"notes-api/internal/models"
)

// NoteService предоставляет методы для работы с заметками
type NoteService struct {
	DB *sql.DB
}

func (s *NoteService) CreateNote(note *models.Note) error {
	query := `INSERT INTO notes (title, content, user_id) VALUES ($1, $2, $3) RETURNING id`
	err := s.DB.QueryRow(query, note.Title, note.Content, note.UserID).Scan(&note.ID)
	return err
}

func (s *NoteService) GetNotes(userID, page, limit int) ([]models.Note, error) {
	offset := (page - 1) * limit
	query := `SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.UserID, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *NoteService) GetNoteByID(noteID, userID int) (models.Note, error) {
	var note models.Note
	query := `SELECT id, title, content, user_id, created_at, updated_at FROM notes WHERE id = $1 AND user_id = $2`
	err := s.DB.QueryRow(query, noteID, userID).Scan(&note.ID, &note.Title, &note.Content, &note.UserID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return note, err
	}
	return note, nil
}

func (s *NoteService) UpdateNote(note *models.Note, userID int) (models.Note, error) {
	// Проверка, что пользователь является владельцем заметки
	existingNote, err := s.GetNoteByID(note.ID, userID)
	if err != nil {
		return existingNote, err
	}
	if existingNote.UserID != userID {
		return existingNote, errors.New("вы не можете редактировать эту заметку")
	}
	query := `UPDATE notes SET title = $1, content = $2 WHERE id = $3`
	_, err = s.DB.Exec(query, note.Title, note.Content, note.ID)
	if err != nil {
		return existingNote, err
	}
	// Получаем теги для обновленной заметки
	tags, err := s.GetTagsForNote(note.ID)
	if err != nil {
		return existingNote, err
	}
	note.Tags = tags
	return *note, nil
}

func (s *NoteService) DeleteNote(noteID, userID int) error {
	// Проверка, что пользователь является владельцем заметки
	existingNote, err := s.GetNoteByID(noteID, userID)
	if err != nil {
		return err
	}
	if existingNote.UserID != userID {
		return errors.New("вы не можете удалить эту заметку")
	}
	query := `DELETE FROM notes WHERE id = $1`
	_, err = s.DB.Exec(query, noteID)
	return err
}

func (s *NoteService) AddTags(noteID int, tags []models.Tag, userID int) error {
	// Проверка, что пользователь является владельцем заметки
	existingNote, err := s.GetNoteByID(noteID, userID)
	if err != nil {
		return err
	}
	if existingNote.UserID != userID {
		return errors.New("вы не можете добавлять теги к этой заметке")
	}
	for _, tag := range tags {
		query := `INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO NOTHING RETURNING id`
		var tagID int
		err := s.DB.QueryRow(query, tag.Name).Scan(&tagID)
		if err != nil {
			return err
		}
		// Связываем заметку с тегом
		if tagID > 0 {
			_, err = s.DB.Exec(`INSERT INTO note_tags (note_id, tag_id) VALUES ($1, $2)`, noteID, tagID)
			if err != nil {
				return err
			}
		}
	}
	// Получаем теги для обновленной заметки
	existingNote.Tags, err = s.GetTagsForNote(noteID)
	return err
}

func (s *NoteService) GetTagsForNote(noteID int) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name
		FROM tags t
		JOIN note_tags nt ON t.id = nt.tag_id
		WHERE nt.note_id = $1`
	rows, err := s.DB.Query(query, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (s *NoteService) GetNotesByTag(tag string) ([]models.Note, error) {
	query := `
		SELECT n.id, n.title, n.content, n.user_id, n.created_at, n.updated_at
		FROM notes n
		JOIN note_tags nt ON n.id = nt.note_id
		JOIN tags t ON nt.tag_id = t.id
		WHERE t.name = $1`
	rows, err := s.DB.Query(query, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.UserID, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	// Если заметок не найдено, можно вернуть ошибку
	if len(notes) == 0 {
		return nil, fmt.Errorf("тег не найден")
	}
	return notes, nil
}

func (s *NoteService) ShareNote(noteID, ownerID, userID int) error {
	// Проверка, что заметка принадлежит владельцу
	var owner int
	err := s.DB.QueryRow("SELECT user_id FROM notes WHERE id = $1", noteID).Scan(&owner)
	if err != nil {
		return fmt.Errorf("не удалось найти заметку: %w", err)
	}
	if owner != ownerID {
		return fmt.Errorf("вы не являетесь владельцем этой заметки")
	}
	// Проверка, что пользователь, которому передается доступ, не является владельцем
	if userID == owner {
		return fmt.Errorf("вы не можете передавать доступ к своей заметке")
	}
	// Передача доступа
	_, err = s.DB.Exec("INSERT INTO note_access (note_id, user_id) VALUES ($1, $2)", noteID, userID)
	if err != nil {
		return fmt.Errorf("не удалось передать доступ: %w", err)
	}
	return nil
}

func (s *NoteService) GetSharedNotes(userID int) ([]models.Note, error) {
	rows, err := s.DB.Query(`
		SELECT n.id, n.title, n.content, n.user_id, n.created_at, n.updated_at 
		FROM notes n 
		INNER JOIN note_access na ON n.id = na.note_id 
		WHERE na.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить доступные заметки: %w", err)
	}
	defer rows.Close()
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.UserID, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}
