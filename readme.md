[![Tests](https://github.com/dgyurics/auth/actions/workflows/tests.yaml/badge.svg)](https://github.com/dgyurics/auth/actions/workflows/tests.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/dgyurics/auth)](https://goreportcard.com/report/github.com/dgyurics/auth)

# Simple, Fault-Tolerant, Distributed Authentication Service

This project consists of three services: `api-gateway`, `auth-server`, and `secure-server`. `api-gateway` is an Nginx server configured with the auth request module, acting as the entry point for all requests. `auth-server` is the authentication server which `api-gateway` calls using subrequests. `secure-server` is an HTTP server accessible to authorized users only.

## Instructions for Running Locally

To run the application locally, follow these steps:

1. From the root directory, run the command `docker compose build` to build the required images.

2. From the root directory, run the command `docker compose -p auth up -d` to start the required services.

3. In a separate terminal, create an account/user by running the following command:
   ```
   curl --location --request POST 'localhost:3000/auth/register' \
   --header 'Content-Type: application/json' \
   --data-raw '{ "username": "newuser", "password": "mypassword"}' \
   --include \
   --cookie-jar cookies.txt \
   --write-out '\nResponse status code: %{response_code}\n'
   ```
   This command will return a session cookie and store it in a file called `cookies.txt`.

4. Use the cookie obtained in the previous step to access the secure server by running the following command:
   ```
   curl --location --request GET 'localhost:3000/api/echo' \
   --cookie cookies.txt \
   --include \
   --write-out '\nResponse status code: %{response_code}\n'
   ```
   This command will use the `X-Session-ID` cookie from the `cookies.txt` file to access the secure server and return the response status code.

That's it! You have successfully authenticated and accessed a secure server using a distributed network of services.