package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID        int       `json:"_id"`
	Completed bool      `json:"completed"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func main() {
	fmt.Println("hello world")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	MYSQL_URI := os.Getenv("MYSQL_URI")
	var err error
	db, err = sql.Open("mysql", MYSQL_URI)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MySQL")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id INT AUTO_INCREMENT,
		completed BOOLEAN,
		body TEXT NOT NULL,
		created_at DATETIME,
		PRIMARY KEY (id)
	);`)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5001"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/To_do_task/dist")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT id, completed, body, created_at FROM todos")
	if err != nil {
		return err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Completed, &todo.Body, &todo.CreatedAt); err != nil {
			return err
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	todo.CreatedAt = time.Now()
	result, err := db.Exec("INSERT INTO todos (completed, body, created_at) VALUES (?, ?, ?)", todo.Completed, todo.Body, todo.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = int(id)

	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	completed := true

	result, err := db.Exec("UPDATE todos SET completed = ? WHERE id = ?", completed, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}
