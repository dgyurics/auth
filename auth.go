package main

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"

	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var ctx = context.Background()
var pool *sql.DB
var rdb *redis.Client

func init() {
	var err error
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	connStr := "user=postgres password=postgres host=localhost port=5432 dbname=auth sslmode=disable"
	pool, err = sql.Open("postgres", connStr) // opens a connection pool, safe for use by multiple goroutines
	checkErr(err)
}

func main() {
	app := fiber.New()

	app.Post("/register", register)
	app.Post("/login", login)
	app.Get("/users", getUsers)
	app.Get("/health", health)

	app.Listen(":3000")
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

	sqlStatement := `SELECT id, username, password FROM public.user WHERE username = $1`
	var hashedPassword []byte
	err = pool.QueryRow(sqlStatement, user.Username).Scan(&user.Id, &user.Username, &hashedPassword)
	if err != nil {
		handleErr(err, c)
		return nil
	}
	// generate session id
	// user.Password = nil
	// store session id in redis, with user account

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(user.Password))
	if err != nil {
		handleErr(err, c)
		return nil
	}

	c.Cookie(&fiber.Cookie{
		Name:  "userId",
		Value: strconv.Itoa(user.Id),
	})
	return c.SendStatus(200)
}

func register(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	checkErr(err)

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		handleErr(err, c)
		return nil
	}

	sqlStatement := `INSERT INTO public.user (username, password) VALUES ($1, $2) RETURNING id, username`
	err = pool.QueryRow(sqlStatement, user.Username, hashedPass).Scan(&user.Id, &user.Username)
	if err != nil {
		handleErr(err, c)
		return nil
	}
	// generate session id
	// user.Password = nil
	// store session id in redis, with user account
	c.Cookie(&fiber.Cookie{
		Name:  "userId",
		Value: strconv.Itoa(user.Id),
	})
	return c.SendStatus(201)
}

func health(c *fiber.Ctx) error {
	ctx, cancelCtx := context.WithDeadline(ctx, time.Now().Add(1000*time.Millisecond))
	defer cancelCtx()

	// ping cache
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return c.Status(503).SendString("cache unavailable")
	}

	// ping db
	err = pool.Ping()
	if err != nil {
		return c.Status(503).SendString("database unavailable")
	}
	return c.SendStatus(200)
}

func handleErr(err error, c *fiber.Ctx) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "pq: duplicate key") {
		c.Status(409).SendString(errMsg)
	} else if strings.Contains(errMsg, "not the hash") {
		c.Status(401).SendString(errMsg)
	} else if strings.Contains(errMsg, "connection refused") {
		c.Status(503).SendString("database unavailable")
	} else if err == sql.ErrNoRows {
		c.Status(404).SendString(errMsg)
	} else {
		c.Status(500).SendString(errMsg)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
