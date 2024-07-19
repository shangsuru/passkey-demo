package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthParams struct {
	Email    string
	Password string
}

type AuthResponse struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

func sendError(ctx echo.Context, err string, code int) error {
	return ctx.JSON(code, AuthResponse{
		Status:       "error",
		ErrorMessage: err,
	})
}

func sendOK(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, AuthResponse{
		Status:       "ok",
		ErrorMessage: "",
	})
}
