package models

import "time"

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []Tag     `json:"tags,omitempty"` // Добавляем поле для тегов
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ShareNoteRequest struct {
	UserID int `json:"user_id"`
}

type SuccessResponse struct {
	Message string      `json:"message"` // Сообщение об успешном выполнении
	Data    interface{} `json:"data"`    // Данные, возвращаемые в ответе (например, заметка)
}
