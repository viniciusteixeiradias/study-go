package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Passowrd string `json:"password"`
}

func saveUser(c echo.HandlerFunc) {}

func get(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

func list(c echo.Context) error {
	name := c.QueryParam("name")
	return c.String(http.StatusOK, name)
}

func create(c echo.Context) (err error) {
	user := new(User)

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	return c.JSON(http.StatusOK, user)
}

func update(c echo.Context) (err error) {
	id := c.Param("id")
	user := new(User)

	if id == "" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	return c.JSON(http.StatusOK, user)
}

func delete(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/users", list)

	e.GET("/users/:id", get)

	e.POST("/users", create)

	e.PUT("/users/:id", update)

	e.DELETE("/users/:id", delete)

	e.Logger.Fatal(e.Start(":3000"))
}
