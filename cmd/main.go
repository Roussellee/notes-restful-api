package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	_ "notes-api/docs"
	"notes-api/internal/config"
	"notes-api/internal/database"
	"notes-api/internal/routes"
)

// Notes RESTful API
// @version 1.0
// @description Notes API - это RESTful API для системы управления заметками, написанный на Go с использованием Gin и PostgreSQL + PgAmdmin4.
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer {token}"

func main() {
	// Загрузка конфигурации
	config.LoadConfig()
	// Инициализация базы данных
	db := database.InitDB()
	defer db.Close()
	database.InitSchema(db)
	// Настройка маршрутизатора
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Настройка маршрутов
	routes.SetupRoutes(router, db)
	log.Println("Сервер запущен на порту 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
