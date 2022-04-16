# Go-Chi Sample Program

This is a simple Go program that uses the [Go-Chi](https://github.com/go-chi/chi) router and middleware to create a web server that responds to HTTP requests.

## Installation

1. Make sure you have Go installed on your system.
2. Clone the repository or download the source code.
3. Run `go build` to build the executable file.
4. Run the executable with `./executable-name` or `go run main.go`.

## Usage

The program starts a web server on port 8080 by default. You can change the port by modifying the `Port` constant in the source code.

The server responds to the following routes:

- `/health`: Returns an HTTP 200 OK response with the body "ok".
- `/echo`: Returns an HTTP 200 OK response with the body "echo".

## Dependencies

The program depends on the following third-party packages:

- [Go-Chi](https://github.com/go-chi/chi): A lightweight, idiomatic and composable router for building Go HTTP services.
- [Go-Chi/Middleware](https://github.com/go-chi/chi/tree/master/middleware): A collection of useful middleware for Go-Chi.

## License

This program is licensed under the MIT License. See the `LICENSE` file for details.