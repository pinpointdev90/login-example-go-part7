package handler

import (
	"login-example/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

type IUserHandler interface {
	PreRegister(c echo.Context) error
	Activate(c echo.Context) error
}

type userHandler struct {
	uu usecase.IUserUsecase
}

func NewUserHandler(uu usecase.IUserUsecase) IUserHandler {
	return &userHandler{uu: uu}
}

func (h *userHandler) PreRegister(c echo.Context) error {
	// リクエストボディを受け取るための構造体を作成します
	rb := struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,gte=6,lte=20"`
	}{}

	// リクエストボディの中身をrbに書き込みます
	if err := c.Bind(&rb); err != nil {
		return err
	}
	// validateタグの内容通りかどうか検証します。
	if err := c.Validate(rb); err != nil {
		return err
	}

	// context.ContextをPreRegisterに渡す必要があるので、echo.Contextから取得します。
	ctx := c.Request().Context()

	_, err := h.uu.PreRegister(ctx, rb.Email, rb.Password)
	if err != nil {
		return err
	}

	// 仮登録が完了したメッセージとしてokとクライアントに返します。
	return c.JSON(http.StatusOK, echo.Map{
		"message": "ok",
	})
}

func (h *userHandler) Activate(c echo.Context) error {
	rb := struct {
		Email string `json:"email" validate:"required,email"`
		Token string `json:"token" validate:"required,len=8"`
	}{}
	if err := c.Bind(&rb); err != nil {
		return err
	}
	if err := c.Validate(rb); err != nil {
		return err
	}

	ctx := c.Request().Context()

	if err := h.uu.Activate(ctx, rb.Email, rb.Token); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "activate ok",
	})
}