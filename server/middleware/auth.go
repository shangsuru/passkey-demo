package middleware

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
)

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sess, _ := session.Get("auth", ctx)
		if userID, ok := sess.Values["user"]; !ok || userID == nil {
			return ctx.Redirect(http.StatusFound, "/")
		}

		return next(ctx)
	}
}
