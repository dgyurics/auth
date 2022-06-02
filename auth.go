package main

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// tasks
// improve error handling function
// refactor code
// add multi role support
// add registration code option
// add logging
// create tests for login, register, get users, health
// deploy via kubernetes
// switch from fiber to gin

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var ctx = context.Background()
var pool *sql.DB
var rdb *redis.Client

func init() {
	var err = godotenv.Load()
	checkErr(err)

	getEnv := os.Getenv

	db, _ := strconv.Atoi(getEnv("REDIS_DB"))
	rdb = redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_HOST") + ":" + getEnv("REDIS_PORT"),
		Password: getEnv("REDIS_PASSWORD"), // no password set
		DB:       db,                       // use default DB
	})

	connStr := "user=" + getEnv("PG_USER") +
		" password=" + getEnv("PG_PASSWORD") +
		" host=" + getEnv("PG_HOST") +
		" port=" + getEnv("PG_PORT") +
		" dbname=" + getEnv("PG_DB") +
		" sslmode=" + getEnv("PG_SSLMODE")
	pool, err = sql.Open("postgres", connStr) // opens a connection pool, safe for use by multiple goroutines
	checkErr(err)
	initTables()
}

func initTables() {
	_, err := pool.Exec("CREATE TABLE IF NOT EXISTS public.user (id SERIAL PRIMARY KEY, username VARCHAR(100) UNIQUE, password VARCHAR(100) NOT NULL)")
	checkErr(err)
}

func main() {
	app := fiber.New()

	app.Use(recover.New())

	app.Post("/register", register)
	app.Post("/login", login)
	app.Get("/users", getUsers)
	app.Get("/health", health)

	app.Listen("127.0.0.1:3000")
}

func getUsers(c *fiber.Ctx) error {
	rows, err := pool.Query("SELECT id, username FROM public.user")
	checkErr(err)

	var users []*User
	for rows.Next() {
		u := new(User)
		err = rows.Scan(&u.Id, &u.Username)
		checkErr(err)
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
		return handleErr(err)
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(user.Password))
	if err != nil {
		return handleErr(err)
	}

	sessionId := uuid.New().String()
	err = rdb.Set(ctx, sessionId, user.Id, 0).Err() // set ttl
	if err != nil {
		return handleErr(err)
	}

	c.Cookie(&fiber.Cookie{
		Name:  "session",
		Value: sessionId,
	})
	return c.SendStatus(200)
}

func register(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	checkErr(err)

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return handleErr(err)
	}

	sqlStatement := `INSERT INTO public.user (username, password) VALUES ($1, $2) RETURNING id, username`
	err = pool.QueryRow(sqlStatement, user.Username, hashedPass).Scan(&user.Id, &user.Username)
	if err != nil {
		return handleErr(err)
	}

	sessionId := uuid.New().String()
	err = rdb.Set(ctx, sessionId, user.Id, 0).Err() // set ttl
	if err != nil {
		return handleErr(err)
	}

	c.Cookie(&fiber.Cookie{
		Name:  "session",
		Value: sessionId,
	})

	return c.SendStatus(201)
}

func health(c *fiber.Ctx) error {
	ctx, cancelCtx := context.WithDeadline(ctx, time.Now().Add(1000*time.Millisecond))
	defer cancelCtx()

	// ping cache
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return handleErr(err)
	}

	// ping db
	err = pool.Ping()
	if err != nil {
		return handleErr(err)
	}
	return c.SendStatus(200)
}

func handleErr(err error) error {
	errMsg := err.Error()
	if strings.Contains(errMsg, "pq: duplicate key") {
		return fiber.NewError(409)
	} else if strings.Contains(errMsg, "not the hash") {
		return fiber.NewError(401, errMsg)
	} else if strings.Contains(errMsg, "connection refused") {
		return fiber.NewError(503, "error connecting to redis cache")
	} else if strings.Contains(errMsg, "bad connection") {
		return fiber.NewError(503, "error connecting to postgresql db")
	} else if err == sql.ErrNoRows {
		return fiber.NewError(404, errMsg)
	} else {
		return fiber.NewError(500, errMsg)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
