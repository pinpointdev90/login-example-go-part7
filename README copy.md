## 今回の目標

- ユーザーの仮登録を実装する！

前回まではentity, repository, usecaseを作成しました。
今回はその残りです。

## 仮登録機能の確認
作業を行う前にもう一度それぞれの内容を確認しておきましょう

| パッケージ | 役割 | 機能 |
|:-----------|:------------|:------------|
| Repository       | DBとのやりとり        | ・emailからユーザーを取得する<br/>・ユーザーを仮登録で保存する<br/>・ユーザーを削除する         |
| Usecase     | 仮登録処理を行う	      | ・ユーザーがアクティブかどうか確認する<br/>・アクティブ（本登録）ならエラー<br/>・非アクティブ（仮登録）なら削除して仮登録処理をやり直す<br/>・本人確認トークンの作成<br/>・メール送信       |
| Handler       | リクエストボディの取得<br/>レスポンスの作成        | ・リクエストボディの取得<br/>・リクエストボディの検証<br/>・emailのフォーマット検証<br/>・パスワードの長さ検証（６〜２０文字）<br/>・レスポンスの作成         |

## Handler
- リクエストボディの取得
- リクエストボディの検証
    - emailのフォーマット検証
        - パスワードの長さ検証（６〜２０文字）
        - レスポンスの作成

検証を自前で実装するのは抜けやモレなど怖いのであるものを使います。

定番ですね。
ではやっていきましょう！

handler/user_handler.go
```
package handler

import (
	"login-example/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

type IUserHandler interface {
	PreRegister(c echo.Context) error
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
```

user_handler.go の解説
```
rb := struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6,lte=20"`
}{}
```
- go-playground/validatorはvalidateタグの内容を見て検証方法を判断します
- validate:"required,email"
    - required: 入力必須, email: メールアドレスのフォーマットチェック
- validate:"required,gte=6,lte=20"
    - required: 入力必須, gte=6: 6文字以上、lte=20: 20文字以内

はい、これで Handler も完成しました。

## エンドポイントを登録しよう！

早速、 /auth/register/initial を登録しますが、その前に諸々のechoの準備をします。
go-playground/validatorをダウンロードしましょう

```
$ go get github.com/go-playground/validator/v10
```

### validator.go

echoでgo-playground/validatorを使うための設定ファイルです。

validator.go
```
package main

import "github.com/go-playground/validator/v10"

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
```
### error_handler.go
エラーが返ってきた時のレスポンス作成を行ってます。
今回は全て500エラーで、エラーのメッセージをそのままレスポンスで返しています。
「自分の家の鍵はXXX社製で、型番はYYYだ！」と言ってるようなものなのでセキュリティ的にはよろしくないです。

error_handler.go
```
package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func customHTTPErrorHandler(err error, c echo.Context) {
	c.Logger().Error(err)
	
	// エラーの内容をそのまま返すのは本当はNG
	if err := c.JSON(http.StatusInternalServerError, echo.Map{
		"message": err.Error(),
	}); err != nil {
		c.Logger().Error(err)
	}
}
```

### router.go
router.go
```
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

	return e
}
```

### main.go

main.go
```
package main

import (
	"fmt"
	"login-example/db"
	"login-example/mail"

	"github.com/go-playground/validator/v10"
)

func main() {
	db, err := db.NewDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	mailer := mail.NewMailhogMailer()

	e := NewRouter(db, mailer)

	// error_handler.goの内容を登録してます。
	e.HTTPErrorHandler = customHTTPErrorHandler
	
	// validator.goの内容を登録してます。
	e.Validator = &CustomValidator{validator: validator.New()}

	e.Logger.Fatal(e.Start(":8000"))
}
```
はい、これで仮登録処理ができました。


## 確認しよう

本当にちゃんとできているか確認しましょう。
まずはサービスを起動します。
```
$ docker compose up -d
```

次にlocalhost:8000/api/auth/register/initialにリクエストを投げましょう。

```
$ curl -XPOST localhost:8000/api/auth/register/initial \
	-H 'Content-Type: application/json; charset=UTF-8' \
	-d '{"email": "test-user-1@example.xyz", "password": "foobar"}'
```

```
{"message":"ok"} # と出力されればOKです。
```

仮登録はできたようです。

次はメールが送信されているか確認しましょう。
ブラウザでlocalhost:8025にアクセスしてください。
次のような画面になるはずです。

![d82ac1ad59e3-20230703.png](https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/3507088/6e54d806-4030-a024-da4b-1431c5bf2817.png)

![684ad3040507-20230703.png](https://qiita-image-store.s3.ap-northeast-1.amazonaws.com/0/3507088/331c50dd-784d-fa88-c15e-1761a5b7b89b.png)

はい、ちゃんとメールも送信されてます。
トークンもちゃんと送信されてますね。

最後にMySQLにちゃんとデータが保存されているか確認しましょう。
・emailは正しく保存されているか？
・passwordはちゃんとハッシュ化されているか？
・メールで送信されているトークンと保存されたユーザーのトークンはちゃんと一致するか？
・stateはinactiveになっているか
確認していきましょう

```
$ docker compose exec db mysql -u login-user -plogin-pass login-db

mysql> SELECT * FROM user\G;
*************************** 1. row ***************************
            id: 100002
         email: test-user-1@example.xyz
      password: $2a$10$wqC8qxr17LIBAEO4oKSLNeQLvMqvl.PFJmddWgKrbO5mVWloAdBme
          salt: BTahFEsnC7qaHkYro1AteRAZVKm8XF
         state: inactive
activate_token: rCT0J3dr
    updated_at: 2023-07-03 04:56:36.991717
    created_at: 2023-07-03 04:56:36.991717
1 row in set (0.01 sec)

ERROR: 
No query specified
```

ちゃんと保存されています。

## まとめ

今回やったこと

- handlerの作成
- validatorの登録
- error_handlerの登録
- routerの登録


今日は以上です。
ありがとうございました。

また、今のディレクトリはこんな感じです。

```
  ├── .air.toml
  ├── _tools
  │   └── mysql
  │       ├── conf.d
  │       │   └── my.cnf
  │       └── init.d
  │           └── init.sql
  ├── Dockerfile
  ├── db
  │   └── db.go
  ├── docker-compose.yml
  ├── entity
  │   └── user.go
+ ├── error_handler.go
  ├── go.mod
  ├── go.sum
+ ├── handler
+ │   └── user_handler.go
  ├── mail
  │   └── mailer.go
+ ├── main.go
+ ├── router.go
+ ├── validator.go
  ├── repository
  │   └── user_repository.go
  └── usecase
      └── user_usecase.go
```