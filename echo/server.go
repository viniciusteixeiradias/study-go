package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type User struct {
	Id       *int   `json:"id,omitempty"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Tables struct {
	Tables map[string][]User `json:"tables"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func openDB() (Tables, error) {
	jsonFile, err := os.Open("database.json")

	if err != nil {
		return Tables{}, fmt.Errorf("failed to open database.json: %w", err)
	}

	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)

	if err != nil {
		return Tables{}, fmt.Errorf("failed to read database.json: %w", err)
	}

	var tables Tables

	if err := json.Unmarshal(byteValue, &tables); err != nil {
		return Tables{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tables, nil
}

func handleErrorResponse(message string) ErrorResponse {
	return ErrorResponse{Error: message}
}

func get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return c.JSON(http.StatusBadRequest, handleErrorResponse("Invalid ID format"))
	}

	tables, err := openDB()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, handleErrorResponse("Error opening database"))
	}

	users, found := tables.Tables["user"]

	if !found {
		return c.JSON(http.StatusNotFound, handleErrorResponse("Table 'user' not found"))
	}

	var foundUser *User
	for _, user := range users {
		if *user.Id == id {
			foundUser = &user
			break
		}
	}

	if foundUser == nil {
		return c.JSON(http.StatusNotFound, handleErrorResponse("User not found"))
	}

	return c.JSON(http.StatusOK, foundUser)
}

func list(c echo.Context) error {
	name := c.QueryParam("name")

	tables, err := openDB()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error opening database")
	}

	users, found := tables.Tables["user"]
	if !found {
		return c.String(http.StatusNotFound, "Table 'user' not found")
	}

	if name == "" {
		return c.JSON(http.StatusOK, users)
	}

	var foundUsers []User
	for _, user := range users {
		if strings.Contains(user.Name, name) {
			foundUsers = append(foundUsers, user)
		}
	}

	if len(foundUsers) == 0 {
		return c.String(http.StatusNotFound, fmt.Sprintf(`{"error": "User with name like %d not found"}`, name))
	}

	return c.JSON(http.StatusOK, foundUsers)
}

func create(c echo.Context) (err error) {
	user := new(User)

	tables, err := openDB()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error opening database")
	}

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	tables.Tables["user"] = append(tables.Tables["user"], *user)
	updatedJSON, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	err = os.WriteFile("database.json", updatedJSON, 0644)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, user)
}

func update(c echo.Context) (err error) {
	user := new(User)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID format")
	}

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	tables, err := openDB()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error opening database")
	}

	users, found := tables.Tables["user"]
	if !found {
		return c.String(http.StatusNotFound, "Table 'user' not found")
	}

	var foundUser *User
	for i, u := range users {
		if *u.Id == id {
			users[i].Name = user.Name
			users[i].Email = user.Email
			users[i].Password = user.Password
			foundUser = &users[i]
			break
		}
	}

	if foundUser == nil {
		return c.String(http.StatusNotFound, fmt.Sprintf(`{"error": "User with ID %d not found"}`, id))
	}

	// Update the 'user' array in the tables
	tables.Tables["user"] = users

	// Marshal the updated tables back to JSON
	updatedJSON, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Write the updated JSON back to the file
	err = os.WriteFile("database.json", updatedJSON, 0644)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, foundUser)
}

func delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID format")
	}

	tables, err := openDB()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error opening database")
	}

	users, found := tables.Tables["user"]
	if !found {
		return c.String(http.StatusNotFound, "Table 'user' not found")
	}

	var updatedUsers []User
	var deletedUser *User
	for _, u := range users {
		if *u.Id == id {
			// Skip the user with the specified ID (effectively deleting it)
			deletedUser = &u
			continue
		}
		updatedUsers = append(updatedUsers, u)
	}

	if deletedUser == nil {
		return c.String(http.StatusNotFound, fmt.Sprintf(`{"error": "User with ID %d not found"}`, id))
	}

	// Update the 'user' array in the tables
	tables.Tables["user"] = updatedUsers

	// Marshal the updated tables back to JSON
	updatedJSON, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Write the updated JSON back to the file
	err = os.WriteFile("database.json", updatedJSON, 0644)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, deletedUser)
}

func routes(g *echo.Group) {
	g.GET("/", list)
	g.GET("/:id", get)
	g.POST("/", create)
	g.PUT("/:id", update)
	g.DELETE("/:id", delete)
}

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	routes(e.Group("/users"))

	e.Logger.Fatal(e.Start(":3000"))
}
