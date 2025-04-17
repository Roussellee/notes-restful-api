package handlers

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"notes-api/internal/models"
	"notes-api/internal/services"
	"os"
	"strconv"
)

// CreateNote @Summary Создание заметки
// @Description Создает новую заметку
// @Tags notes
// @Accept json
// @Produce json
// @Param note body models.Note true "Заметка"
// @Success 201 {object} models.SuccessResponse "Созданная заметка"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /notes [post]
// @Security Bearer
func CreateNote(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var note models.Note
		if err := c.ShouldBindJSON(&note); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Извлечение user_id из токена
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен аутентификации"})
			return
		}
		note.UserID = userID // Устанавливаем user_id для заметки
		noteService := services.NoteService{DB: db}
		if err := noteService.CreateNote(&note); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании заметки"})
			return
		}

		c.JSON(http.StatusCreated, note)
	}
}

// GetNotes возвращает список заметок с пагинацией
// @Summary Получение списка заметок
// @Description Получает список заметок с пагинацией
// @Tags notes
// @Produce json
// @Param page query int false "Номер страницы"
// @Param limit query int false "Количество заметок на странице"
// @Success 200 {array} models.Note "Список заметок"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Router /notes [get]
// @Security Bearer
func GetNotes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен аутентификации"})
			return
		}
		pageStr := c.Query("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1 // По умолчанию первая страница
		}
		limitStr := c.Query("limit")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10 // По умолчанию 10 заметок на страницу
		}
		noteService := services.NoteService{DB: db}
		notes, err := noteService.GetNotes(userID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении заметок"})
			return
		}
		// Получаем теги для каждой заметки
		for i := range notes {
			tags, err := noteService.GetTagsForNote(notes[i].ID)
			if err == nil {
				notes[i].Tags = tags
			}
		}

		c.JSON(http.StatusOK, notes)
	}
}

// GetNoteByID @Summary Получение заметки по ID
// @Description Получает заметку по ID
// @Tags notes
// @Produce json
// @Param id path int true "ID заметки"
// @Success 200 {object} models.SuccessResponse "Заметка"
// @Failure 404 {object} models.ErrorResponse "Заметка не найдена или доступ запрещен"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 400 {object} models.ErrorResponse "id заметки должен быть формата int"
// @Router /notes/{id} [get]
// @Security Bearer
func GetNoteByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteIDStr := c.Param("id")
		noteID, err := strconv.Atoi(noteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id заметки должен быть формата int"})
			return
		}
		// Извлекаем userID из токена
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Вы должны быть авторизованы"})
			return
		}
		noteService := services.NoteService{DB: db}
		note, err := noteService.GetNoteByID(noteID, userID) // Передаем userID
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Заметка не найдена или доступ запрещен"})
			return
		}
		// Получаем теги для заметки
		tags, err := noteService.GetTagsForNote(note.ID)
		if err == nil {
			note.Tags = tags
		}

		c.JSON(http.StatusOK, note)
	}
}

// UpdateNote @Summary Обновление заметки
// @Description Обновляет существующую заметку
// @Tags notes
// @Accept json
// @Produce json
// @Param id path int true "ID заметки"
// @Param note body models.Note true "Обновленная заметка"
// @Success 200 {object} models.SuccessResponse "Обновленная заметка"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 404 {object} models.ErrorResponse "Заметка не найдена"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Router /notes/{id} [put]
// @Security Bearer
func UpdateNote(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var note models.Note
		if err := c.ShouldBindJSON(&note); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		noteIDStr := c.Param("id")
		noteID, err := strconv.Atoi(noteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id заметки должен быть формата int"})
			return
		}
		note.ID = noteID
		// Проверка, что пользователь является владельцем заметки
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Вы должны быть авторизованы"})
			return
		}
		noteService := services.NoteService{DB: db}
		updatedNote, err := noteService.UpdateNote(&note, userID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не можете редактировать эту заметку"})
			return
		}
		c.JSON(http.StatusOK, updatedNote) // Возвращаем обновленную заметку с тегами
	}
}

// DeleteNote @Summary Удаление заметки
// @Description Удаляет заметку по ID
// @Tags notes
// @Produce json
// @Param id path int true "ID заметки"
// @Success 200 {object} models.SuccessResponse "Успешный ответ"
// @Failure 404 {object} models.ErrorResponse "Заметка не найдена"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Router /notes/{id} [delete]
// @Security Bearer
func DeleteNote(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteIDStr := c.Param("id")
		noteID, err := strconv.Atoi(noteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "note_id must be an integer"})
			return
		}
		// Проверка, что пользователь является владельцем заметки
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен аутентификации"})
			return
		}
		noteService := services.NoteService{DB: db}
		if err := noteService.DeleteNote(noteID, userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не можете удалить эту заметку"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Заметка успешно удалена"})
	}
}

// AddTags добавляет теги к заметке
// @Summary Добавление тегов к заметке
// @Description Добавляет теги к заметке по ID
// @Tags notes
// @Accept json
// @Produce json
// @Param id path int true "ID заметки"
// @Param tags body []models.Tag true "Список тегов"
// @Success 200 {object} models.SuccessResponse "Обновленная заметка с тегами"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Запрещено"
// @Failure 404 {object} models.ErrorResponse "Заметка не найдена"
// @Router /notes/{id}/tags [post]
// @Security Bearer
func AddTags(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteIDStr := c.Param("id")
		noteID, err := strconv.Atoi(noteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "note_id must be an integer"})
			return
		}
		var tags []models.Tag
		if err := c.ShouldBindJSON(&tags); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Проверка, что пользователь является владельцем заметки
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен аутентификации"})
			return
		}
		noteService := services.NoteService{DB: db}
		if err := noteService.AddTags(noteID, tags, userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не можете добавлять теги к этой заметке"})
			return
		}
		// Получаем обновленную заметку с тегами
		updatedNote, err := noteService.GetNoteByID(noteID, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Заметка не найдена"})
			return
		}
		// Получаем теги для обновленной заметки
		updatedNote.Tags, err = noteService.GetTagsForNote(noteID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении тегов"})
			return
		}
		c.JSON(http.StatusOK, updatedNote) // Возвращаем обновленную заметку с тегами
	}
}

// GetNotesByTag возвращает заметки по тегу
// @Summary Получение заметок по тегу
// @Description Возвращает список заметок, связанных с определенным тегом
// @Tags notes
// @Produce json
// @Param tag query string true "Тег"
// @Success 200 {array} models.Note "Список заметок"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 404 {object} models.ErrorResponse "Тег не найден"
// @Router /notes/tag [get]
func GetNotesByTag(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := c.Query("tag")
		if tag == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Тег обязателен"})
			return
		}
		noteService := services.NoteService{DB: db}
		notes, err := noteService.GetNotesByTag(tag)
		if err != nil {
			if err.Error() == "тег не найден" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Тег не найден"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении заметок по тегу"})
			return
		}
		c.JSON(http.StatusOK, notes)
	}
}

// ShareNote - обработчик для передачи доступа к заметке
// @Summary Передача доступа к заметке
// @Description Делает заметку доступной для другого пользователя
// @Tags notes
// @Accept json
// @Produce json
// @Param id path int true "ID заметки"
// @Param requestBody body models.ShareNoteRequest true "ID пользователя, которому передается доступ"
// @Success 200 {object} map[string]string "Сообщение об успешной передаче доступа"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Запрещено"
// @Router /notes/{id}/share [post]
// @Security Bearer
func ShareNote(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteIDStr := c.Param("id")
		noteID, err := strconv.Atoi(noteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "note_id должен быть integer"})
			return
		}
		var requestBody struct {
			UserID int `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ownerID, err := getUserIDFromToken(c) // Получение ID владельца заметки из токена
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен аутентификации"})
			return
		}
		noteService := services.NoteService{DB: db}
		if err := noteService.ShareNote(noteID, ownerID, requestBody.UserID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()}) // Возвращаем конкретное сообщение об ошибке
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Доступ к заметке успешно передан"})
	}
}

// GetSharedNotes - возвращает список заметок, доступных текущему пользователю
// @Summary Получение списка доступных заметок
// @Description Возвращает заметки, к которым у пользователя есть доступ
// @Tags notes
// @Produce json
// @Success 200 {array} models.Note "Список доступных заметок"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Router /notes/shared-notes [get]
// @Security Bearer
func GetSharedNotes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getUserIDFromToken(c) // Получение ID пользователя из токена
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен аутентификации"})
			return
		}
		noteService := services.NoteService{DB: db}
		notes, err := noteService.GetSharedNotes(userID)
		if err != nil {
			log.Printf("Ошибка при получении доступных заметок: %v", err) // Логируем ошибку
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить доступные заметки"})
			return
		}
		c.JSON(http.StatusOK, notes)
	}
}

// Вспомогательная функция для извлечения user_id из токена
func getUserIDFromToken(c *gin.Context) (int, error) {
	tokenString, err := c.Cookie("tokenJWT")
	if err != nil {
		return 0, err
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrNotSupported
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int(claims["sub"].(float64)), nil
	}
	return 0, err
}
