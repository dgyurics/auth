package main

import (
	"database/sql"
	"log"

	"strconv"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var pool *sql.DB // database connection pool

func init() {
	connStr := "user=postgres password=postgres host=localhost port=5432 dbname=auth sslmode=disable"
	var err error
	// Opening a driver typically will not attempt to connect to the database.
	pool, err = sql.Open("postgres", connStr)
	checkErr(err)
}

func main() {
	app := fiber.New()

	app.Get("/", index)
	app.Post("/register", register)
	app.Get("/users", printUsers)
	app.Get("/health", health)
	log.Fatalln(app.Listen(":3000"))
}

func index(c *fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
}

func printUsers(c *fiber.Ctx) error {
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
	// return users as json
	return c.JSON(users)
	//return c.SendStatus(200)
}

func register(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	checkErr(err)
	log.Println(user.Username, user.Password)

	sqlStatement := `INSERT INTO public.user (username, password)
		VALUES ($1, $2) RETURNING id`
	id := -1
	err = pool.QueryRow(sqlStatement, user.Username, user.Password).Scan(&id)
	checkErr(err)

	// get username and password
	// insert into DB if not exists
	return c.SendString("user registered " + strconv.Itoa(id))
}

func login() string {
	// get username and password
	// check exists in DB
	// if exists return id
	message := "logging in"
	return message
}

func health(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func logout()          {}
func publicResource()  {}
func privateResource() {}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
