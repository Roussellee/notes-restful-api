package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"notes-api/internal/models"
	"os"
	"strings"
	"time"
)

// UserService предоставляет методы для работы с пользователями
type UserService struct {
	DB *sql.DB
}

func (s *UserService) RegisterUser(user *models.User) error {
	// Проверяем, существует ли уже пользователь с таким именем
	var existingUser models.User
	query := `SELECT id FROM users WHERE username = $1`
	err := s.DB.QueryRow(query, user.Username).Scan(&existingUser.ID)

	if err == nil {
		// Если err == nil, значит, пользователь с таким именем уже существует
		return errors.New("пользователь с таким именем уже существует")
	} else if err != sql.ErrNoRows {
		// Если произошла другая ошибка, возвращаем её
		return err
	}

	// Если пользователя с таким именем не существует, продолжаем регистрацию
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	query = `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id, created_at`
	err = s.DB.QueryRow(query, user.Username, user.Password).Scan(&user.ID, &user.CreatedAt)
	return err
}

func (s *UserService) LoginUser(user *models.User) (string, error) {
	user.Username = strings.TrimSpace(user.Username)
	user.Password = strings.TrimSpace(user.Password)
	if user.Username == "" || user.Password == "" {
		return "", errors.New("имя пользователя и пароль не могут быть пустыми")
	}
	var storedUser models.User
	query := `SELECT id, password, created_at FROM users WHERE username = $1`
	err := s.DB.QueryRow(query, user.Username).Scan(&storedUser.ID, &storedUser.Password, &storedUser.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("неверные учетные данные")
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		fmt.Printf("Error comparing passwords: %v\n", err)
		return "", errors.New("неверные пароль")
	}
	token, err := generateJWT(storedUser.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) GetUserByID(userID int) (models.UserProfile, error) {
	var user models.UserProfile
	query := `SELECT id, username, created_at FROM users WHERE id = $1`
	err := s.DB.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

// generateJWT генерирует JWT для пользователя
func generateJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Токен действителен 24 часа
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
