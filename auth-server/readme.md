# README

This code is a server written in Go language that handles HTTP requests for user registration, login, and logout, and retrieves user information. 

## Dependencies

This service depends on two external services:

- Redis: This is used as a cache to store session information. The service expects Redis to be running on port `6379`.

- PostgreSQL: This is used as the database to store user information. The service expects PostgreSQL to be running on port `5432`.

## Installation

To install and run the server, follow these steps:

1. Install Go language from [here](https://golang.org/dl/).

2. Clone the repository:

   ```
   git clone https://github.com/dgyurics/auth.git
   ```

3. Change directory to the server package:

   ```
   cd auth/auth-server
   ```

4. Build the server:

   ```
   go build -v -o authserver ./cmd/main.go
   ```

5. Run the server:

   ```
   ./authserver
   ```

   The server will be listening on port 8080 by default.

## Usage

The server handles the following endpoints:

- `GET /health`: a health check endpoint that returns HTTP 200 OK.
- `POST /register`: an endpoint for user registration. It expects a JSON object containing `username` (string) and `password` (string). If the registration is successful, it returns HTTP 201 Created. If the username already exists, it returns HTTP 409 Conflict.
- `POST /login`: an endpoint for user login. It expects a JSON object containing the following fields: `username` (string) and `password` (string). If the login is successful, it returns HTTP 200 OK and sets a session cookie. If the username or password is incorrect, it returns HTTP 401 Unauthorized Request.
- `POST /logout`: an endpoint for user logout. It invalidates the session cookie and removes the session from the Redis cache. It returns HTTP 200 OK. Optionally, you can include a query parameter `all` with a value of `true` to log out all sessions for the user.
- `GET /user`: a secure endpoint that retrieves user information. It expects a valid session cookie with a session ID. If the session is invalid or the cookie is missing, it returns HTTP 401 Unauthorized. If the session is valid, it returns HTTP 200 OK and a JSON object containing the user information (except for the password).

## License

This code is licensed under the MIT License. See the [LICENSE](https://github.com/dgyurics/auth/blob/master/LICENSE) file for details.
