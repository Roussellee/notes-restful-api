package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

func InitDB() *sql.DB {
	//Используем переменные окружения для подключения к БД
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	var db *sql.DB
	var err error
	// Попытка подключиться к базе данных с задержкой
	for retries := 0; retries < 5; retries++ {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("Ошибка при подключении к базе данных: %v", err)
			time.Sleep(2 * time.Second) // Задержка перед следующей попыткой
			continue
		}
		if err = db.Ping(); err == nil {
			log.Println("Успешно подключено к базе данных")
			return db
		}
		log.Printf("Ошибка при проверке соединения с базой данных: %v", err)
		time.Sleep(2 * time.Second) // Задержка перед следующей попыткой
	}
	log.Fatalf("Не удалось подключиться к базе данных после нескольких попыток: %v", err)
	return nil
}

func InitSchema(db *sql.DB) {
	// Проверка и создание таблицы пользователей
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(100) UNIQUE NOT NULL,
        password VARCHAR(100) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	// Проверка и создание таблицы заметок
	createNotesTable := `
    CREATE TABLE IF NOT EXISTS notes (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	// Создание функции триггера для обновления поля updated_at
	createUpdateTimestampFunction := `
    CREATE OR REPLACE FUNCTION update_timestamp()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = CURRENT_TIMESTAMP;
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;`
	// Удаление триггера, если он существует, перед его созданием
	dropUpdateTimestampTrigger := `
    DROP TRIGGER IF EXISTS update_notes_timestamp ON notes;`
	// Создание триггера для таблицы notes
	createUpdateTimestampTrigger := `
    CREATE TRIGGER update_notes_timestamp
    BEFORE UPDATE ON notes
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();`
	// Проверка и создание таблицы тегов
	createTagsTable := `
    CREATE TABLE IF NOT EXISTS tags (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50) UNIQUE NOT NULL
    );`
	// Проверка и создание таблицы для связи заметок и тегов
	createNoteTagsTable := `
    CREATE TABLE IF NOT EXISTS note_tags (
        note_id INT REFERENCES notes(id) ON DELETE CASCADE,
        tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
        PRIMARY KEY (note_id, tag_id)
    );`
	// Проверка и создание таблицы для доступа к заметкам
	createNoteAccessTable := `
    CREATE TABLE IF NOT EXISTS note_access (
        note_id INT REFERENCES notes(id) ON DELETE CASCADE,
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        PRIMARY KEY (note_id, user_id)
    );`
	// Выполнение SQL-запросов для создания таблиц и триггеров
	tables := []string{
		createUsersTable,
		createNotesTable,
		createTagsTable,
		createNoteTagsTable,
		createNoteAccessTable,
		createUpdateTimestampFunction,
		dropUpdateTimestampTrigger,
		createUpdateTimestampTrigger,
	}
	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			log.Fatalf("Ошибка при создании таблицы или триггера: %v", err)
		}
	}
	log.Println("Таблицы и триггеры успешно созданы или уже существуют")
}
