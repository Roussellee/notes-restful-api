package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type UserProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type UserSuccess struct {
	Username string `json:"username"`
	Password string `json:"json"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

//type errorResponse struct {
//	Error string `json:"error"`
//}
//
//type statusResponse struct {
//	Status string `json:"status"`
//}
//
//func newErrorResponse(c *gin.Context statusCode int, message string)  {
//	logrus.Error(message)
//	c.AbortWithError(statusCode, errorResponse{message})
//}
