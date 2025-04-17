# Notes API

## Описание проекта

Notes API - это RESTful API для системы управления заметками, написанный на GO с использованием Gin и PostgreSQL. API предоставляет возможности управления пользователями, включая регистрацию, аутентификацию и получение профиля. Также реализован функционал работы с заметками: создание, редактирование, удаление, просмотр списка заметок с пагинацией, фильтрацией заметок по тегам, передачей доступа к заметкам другому пользоателю и просмотр доступных заметок.

## Основные эндпоинты

- `POST /register` - регистрация пользователя
- `POST /login` - аутентификация
- `GET /profile` - получение профиля пользователя
- `POST /notes` - создание заметки
- `GET /notes` - получение списка заметок (с пагинацией)
- `GET /notes/{id}` - получение заметки по ID
- `PUT /notes/{id}` - редактирование заметки
- `DELETE /notes/{id}` - удаление заметки
- `POST /notes/{id}/tags` - добавление тегов к заметке
- `GET /notes?tags=example` - фильтрация заметок по тегам
- `POST /notes/{id}/share` - передача доступа к заметке другому пользователю
- `GET /shared-notes` - просмотр заметок, доступных текущему пользователю

## Установка и запуск

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/Roussellee/notes-restful-api.git
   cd notes-restful-api
2. Установите зависимости:
   ```
    go mod tidy
3. Создайте файл .env заполните его своими данными:
    ```
   POSTGRES_USER=postgres
    POSTGRES_PASSWORD=your_password
    POSTGRES_DB=your_db_name
    JWT_SECRET=your_secret_key
    PGADMIN_EMAIL=your_email
    PGADMIN_PASSWORD=your_password
4. Запустите приложение с помощью Docker Compose:
    ```
   docker-compose up --build
5. Откройте Swagger UI в браузере по адресу http://localhost:8080/swagger/index.html для просмотра документации API (Большинство запросов работает адекватно только через Postman).

## Контейнеризация

Проект контейнеризирован с использованием Docker и Docker Compose. Включает в себя сервисы для приложения и базы данных PostgreSQL.

## Тестирование

Для тестирования API вы можете использовать Postman. Примеры запросов приведены ниже:

- **Регистрация пользователя**:

```bash
POST http://localhost:8080/register '{"username": "testuser", "password": "testpass"}'
```
- **Аутентификация пользователя**:
```bash
POST http://localhost:8080/login '{"username": "testuser", "password": "testpass"}'
```
- **Создание заметки**:
```bash
POST http://localhost:8080/notes '{"title": "My Note", "content": "This is a note."}'
```
- **Получить заметки**:
```bash
GET http://localhost:8080/notes
```
