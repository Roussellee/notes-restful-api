package routes

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"notes-api/internal/handlers"
)

func SetupRoutes(router *gin.Engine, db *sql.DB) {
	// Регистрация пользователя
	router.POST("/register", handlers.RegisterUser(db))
	// Аутентификация
	router.POST("/login", handlers.LoginUser(db))
	// Получение профиля пользователя
	router.GET("/profile", handlers.GetProfile(db))
	// Заметки
	router.POST("/notes", handlers.CreateNote(db))
	router.GET("/notes", handlers.GetNotes(db)) // Пагинация
	router.GET("/notes/:id", handlers.GetNoteByID(db))
	router.PUT("/notes/:id", handlers.UpdateNote(db))
	router.DELETE("/notes/:id", handlers.DeleteNote(db))
	router.POST("/notes/:id/tags", handlers.AddTags(db))
	router.GET("/notes/tags", handlers.GetNotesByTag(db))
	router.POST("/notes/:id/share", handlers.ShareNote(db))  // Новый маршрут для передачи доступа
	router.GET("/shared-notes", handlers.GetSharedNotes(db)) // Новый маршрут для просмотра доступных заметок
	// Добавляем обработчик для главной страницы
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Привет, мир!!!") // Отправляем ответ "Привет, мир!"
	})
}
