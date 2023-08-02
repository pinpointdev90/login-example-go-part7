package main

import (
	"login-example/handler"
	"login-example/mail"
	"login-example/repository"
	"login-example/usecase"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func NewRouter(db *sqlx.DB, mailer mail.IMailer) *echo.Echo {
	e := echo.New()

	ur := repository.NewUserRepository(db)
	uu := usecase.NewUserUsecase(ur, mailer)
	uh := handler.NewUserHandler(uu)

	a := e.Group("/api/auth")
	a.POST("/register/initial", uh.PreRegister)
	a.POST("/register/complete", uh.Activate)

	return e
}