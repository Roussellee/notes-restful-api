package handlers

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"notes-api/internal/models"
	"notes-api/internal/services"
	"os"
)

// RegisterUser  @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param input body models.User true "Пользователь"
// @Success 201 {object} models.UserSuccess "Успешно зарегистрирован"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 409 {object} models.ErrorResponse "Пользователь с таким именем уже существует"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /register [post]
func RegisterUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userService := services.UserService{DB: db}
		if err := userService.RegisterUser(&user); err != nil {
			if err.Error() == "пользователь с таким именем уже существует" {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // Используем статус 409
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при регистрации пользователя"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
	}
}

// LoginUser   @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя и устанавливает cookie с токеном
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "Пользователь"
// @Success 200 {object} models.ErrorResponse "Успешно аутентифицирован"
// @Failure 400 {object} models.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Router /login [post]
func LoginUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userService := services.UserService{DB: db}
		token, err := userService.LoginUser(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Установите cookie с токеном
		c.SetCookie("tokenJWT", token, 3600, "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Успешная аутентификация"})
	}
}

// GetProfile @Summary Получение профиля пользователя
// @Description Получает профиль текущего пользователя
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.UserProfile "Профиль пользователя"
// @Failure 401 {object} models.ErrorResponse "Требуется токен аутентификации"
// @Failure 401 {object} models.ErrorResponse "Неверный токен"
// @Failure 500 {object} models.ErrorResponse "Ошибка при получении профиля"
// @Router /profile [get]
func GetProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("tokenJWT")
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Требуется токен аутентификации"})
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := int(claims["sub"].(float64))
			userService := services.UserService{DB: db}
			userProfile, err := userService.GetUserByID(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка при получении профиля"})
				return
			}
			c.JSON(http.StatusOK, userProfile)
		} else {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Неверный токен"})
		}
	}
}
