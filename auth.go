package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// tasks
// switch from fiber to gin
// create tests for login, register, get users, health
//

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

const ADMIN_ROLE = "admin"
const USER_ROLE = "user"
const GUEST_ROLE = "guest"

const HOST = "127.0.0.1"
const PORT = "3000"

var ctx = context.Background()
var pool *sql.DB
var rdb *redis.Client

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func initLogging() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func init() {
	var err = godotenv.Load()
	checkErr(err)

	go initLogging()

	getEnv := os.Getenv

	verifyEnvirnmentVariables()

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

// TODO verify all required environment variables exist
func verifyEnvirnmentVariables() {
}

// TODO wrap into single transaction
func initTables() {
	_, err := pool.Exec("CREATE TABLE IF NOT EXISTS public.user (" +
		"id SERIAL PRIMARY KEY," +
		"username VARCHAR(100) UNIQUE," +
		"password VARCHAR(100) NOT NULL)")
	checkErr(err)

	pool.Exec("CREATE TYPE role_enum AS ENUM ('guest', 'user', 'admin')")
	// checkErr(err) // don't check error for above statement because CREATE TYPE [IF NOT EXISTS] syntax not available

	_, err = pool.Exec("CREATE TABLE IF NOT EXISTS public.role (" +
		"user_id INTEGER references public.user(id)," +
		"role role_enum NOT NULL," +
		"UNIQUE (user_id, role))")
	checkErr(err)

	_, err = pool.Exec("CREATE TABLE IF NOT EXISTS public.registration_code (" +
		"code char(5) NOT NULL," +
		"expiration TIMESTAMP," +
		"UNIQUE (code))")
	checkErr(err)
}

func main() {
	app := fiber.New()

	r := gin.Default()
	r.Use(errorHandler)
	r.POST("/register", register)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	rand.Seed(time.Now().UnixNano())

	app.Use(recover.New())

	// app.Post("/register", register)
	app.Post("/login", login)
	app.Post("/registration-code", generateRegistrationCode)
	app.Get("/users", getUsers)
	app.Get("/health", health)

	addr := HOST
	if PORT != "80" {
		addr = addr + ":" + PORT
	}
	app.Listen(addr)
}

// verify cookie exists + session is valid
func authenticate(c *fiber.Ctx) (string, error) {
	sessionId := c.Cookies("session", "")
	if sessionId == "" {
		return "", errors.New("unauthorized")
	}
	// check id exists in redis
	val, err := rdb.Get(ctx, sessionId).Result()
	if err != nil {
		return "", errors.New("unauthorized")
	}
	return val, nil
}

// verify user has the specified role
func authorize(userId string, role string) error {
	id, _ := strconv.Atoi(userId)
	sqlStatement := `SELECT user_id FROM public.role WHERE user_id = $1 AND role = $2`
	err := pool.QueryRow(sqlStatement, id, role).Scan(&userId)
	if err != nil {
		return errors.New("forbidden")
	}
	return nil
}

func getUsers(c *fiber.Ctx) error {
	// verify user has active session
	userId, err := authenticate(c)
	if err != nil {
		return handleErr(err)
	}
	// verify user has admin role
	err = authorize(userId, ADMIN_ROLE)
	if err != nil {
		return handleErr(err)
	}

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

func register(c *gin.Context) {
	var user User
	var registrationCode = c.Query("code")
	err := c.BindJSON(&user)
	checkErr(err)

	// verify registration code
	REGISTRATION_CODE_REQUIRED, _ := strconv.ParseBool(os.Getenv("REGISTRATION_CODE_REQUIRED"))
	if REGISTRATION_CODE_REQUIRED {
		if registrationCode == "nil" {
			c.AbortWithError(500, errors.New("registration code missing"))
			return
		}
		sqlStatement := `DELETE FROM public.registration_code WHERE code = $1 AND expiration > NOW()`
		res, err := pool.Exec(sqlStatement, registrationCode)
		if err != nil {
			c.Error(err)
		}
		count, err := res.RowsAffected()
		if err != nil {
			c.Error(err)
		}
		if count == 0 {
			c.Error(err)
		}
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// FIXME wrap inserts into single transaction

	sqlStatement := `INSERT INTO public.user (username, password) VALUES ($1, $2) RETURNING id, username`
	err = pool.QueryRow(sqlStatement, user.Username, hashedPass).Scan(&user.Id, &user.Username)
	if err != nil {
		c.AbortWithError(409, err)
		return
	}

	sqlStatement = `INSERT INTO public.role (user_id, role) VALUES ($1, $2)`
	_, err = pool.Exec(sqlStatement, user.Id, "user")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	sessionId := uuid.New().String()
	err = rdb.Set(ctx, sessionId, user.Id, 0).Err() // set ttl
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.SetCookie("session", sessionId, -1, "/", HOST, false, false) // FIXME
	c.Status(201)
}

func generateRegistrationCode(c *fiber.Ctx) error {
	var options = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	// verify user has active session
	userId, err := authenticate(c)
	if err != nil {
		return handleErr(err)
	}
	// verify user has admin role
	err = authorize(userId, ADMIN_ROLE)
	if err != nil {
		return handleErr(err)
	}

	go removeExpiredCodes()

	// generate unique registration code
	b := make([]rune, 5)
	for i := range b {
		b[i] = options[rand.Intn(len(options))]
	}

	// insert code into SQL with 30 minute expiration
	code := string(b)
	sqlStatement := `INSERT INTO public.registration_code (code, expiration) VALUES ($1, $2)`
	_, err = pool.Exec(sqlStatement, code, time.Now().Add(time.Minute*30))
	if err != nil {
		return handleErr(err)
	}

	return c.SendString(code)
}

func removeExpiredCodes() {
	sqlStatement := `DELETE FROM public.registration_code WHERE expiration < NOW()`
	_, err := pool.Exec(sqlStatement)
	if err != nil {
		ErrorLogger.Println(err.Error())
	}
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

func errorHandler(c *gin.Context) {
	c.Next()

	// for _, err := range c.Errors {
	// 	switch err.Err {
	// 	case "ErrNotFound":
	// 		c.JSON(-1, gin.H{"error": ErrNotFound.Error()})
	// 	}
	// }

	//c.Status(http.StatusInternalServerError)
	// status -1 doesn't overwrite existing status code
	c.Status(-1)
}

// FIXME
// refactor
// add ip address to log
func handleErr(err error) error {
	errMsg := err.Error()

	InfoLogger.Println(errMsg)

	if strings.Contains(errMsg, "pq: duplicate key") {
		return fiber.NewError(409)
	} else if strings.Contains(errMsg, "not the hash") {
		return fiber.NewError(401, errMsg)
	} else if strings.Contains(errMsg, "unauthorized") {
		return fiber.NewError(401, errMsg)
	} else if strings.Contains(errMsg, "forbidden") {
		return fiber.NewError(403, errMsg)
	} else if strings.Contains(
		errMsg, "missing registration code") ||
		strings.Contains(errMsg, "invalid registration code") {
		return fiber.NewError(400, errMsg)
	} else if strings.Contains(errMsg, "connection refused") {
		ErrorLogger.Println("503-STATUS-CODE:" + errMsg)
		return fiber.NewError(503, "error connecting to redis cache")
	} else if strings.Contains(errMsg, "bad connection") {
		ErrorLogger.Println("503-STATUS-CODE:" + errMsg)
		return fiber.NewError(503, "error connecting to postgresql db")
	} else if err == sql.ErrNoRows {
		return fiber.NewError(404, errMsg)
	} else {
		ErrorLogger.Println("500-STATUS-CODE:" + errMsg)
		return fiber.NewError(500, errMsg)
	}
}

func checkErr(err error) {
	if err != nil {
		ErrorLogger.Println(err.Error())
		panic(err)
	}
}
