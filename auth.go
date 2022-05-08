package main

import (
	"database/sql"
	"log"
	"strconv"

	"strings"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var pool *sql.DB

func init() {
	connStr := "user=postgres password=postgres host=localhost port=5432 dbname=auth sslmode=disable"
	var err error
	pool, err = sql.Open("postgres", connStr) // opens a connection pool, safe for use by multiple goroutines
	checkErr(err)
}

func main() {
	app := fiber.New()

	app.Get("/", index)
	app.Post("/register", register)
	app.Post("/login", login)
	app.Get("/users", getUsers)
	app.Get("/health", health)

	app.Listen(":3000")
}

func index(c *fiber.Ctx) error {
	return c.SendStatus(404)
}

func getUsers(c *fiber.Ctx) error {
	rows, err := pool.Query("SELECT id, username, password FROM public.user")
	checkErr(err)

	var users []*User
	for rows.Next() {
		u := new(User)
		err = rows.Scan(&u.Id, &u.Username, &u.Password)
		checkErr(err)
		log.Println(u.Id, u.Username, u.Password)
		users = append(users, u)
	}
	defer rows.Close()
	return c.JSON(users)
}

func login(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	checkErr(err)

	sqlStatement := `SELECT id, username, password FROM public.user WHERE username = $1 AND password = $2`
	err = pool.QueryRow(sqlStatement, user.Username, user.Password).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		handleErr(err, c)
		return nil
	}

	c.Cookie(&fiber.Cookie{
		Name:  "user",
		Value: strconv.Itoa(user.Id),
	})
	return c.SendStatus(200)
}

func register(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	checkErr(err)

	sqlStatement := `INSERT INTO public.user (username, password) VALUES ($1, $2) RETURNING id`
	err = pool.QueryRow(sqlStatement, user.Username, user.Password).Scan(&user.Id)
	if err != nil {
		handleErr(err, c)
		return nil
	}
	return c.Status(201).SendString("user registered id:" + strconv.Itoa(user.Id))
}

func health(c *fiber.Ctx) error {
	// ping database
	// ping redis
	return c.SendStatus(200)
}

func handleErr(err error, c *fiber.Ctx) {
	if strings.Contains(err.Error(), "pq: duplicate key") {
		c.Status(409).SendString(err.Error())
	} else if strings.Contains(err.Error(), "connection refused") {
		c.Status(503).SendString("database unavailable")
	} else if err == sql.ErrNoRows {
		c.Status(404).SendString(err.Error())
	} else {
		c.Status(500).SendString(err.Error())
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
