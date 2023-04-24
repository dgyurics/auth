# Reverse Proxy Configuration using NGINX

This configuration file demonstrates how to use NGINX as a reverse proxy to load balance traffic between multiple instances of an application server. The configuration file is written in the NGINX configuration syntax and provides an example of how to configure a reverse proxy to distribute HTTP requests to a pool of application servers.

## Prerequisites

Before using this configuration file, ensure that the following software is installed and running on your system:

- NGINX

## Configuration Details

The configuration file specifies the following details:

- `worker_connections`: This directive sets the maximum number of simultaneous connections that can be handled by the worker process.
- `upstream`: This block specifies the pool of application servers that will receive the proxied connections. In this example, the `auth_servers` pool contains a single server listening on port 8080.
- `server`: This block defines the configuration for the server. It specifies that the server should listen on both IPv6 and IPv4 addresses on port 80. The `location` block then proxies all incoming HTTP requests to the `auth_servers` pool using the `proxy_pass` directive.

## Usage

To use this configuration file, follow these steps:

1. Install NGINX on your system.
2. Copy the contents of the configuration file into the NGINX configuration file (`nginx.conf`).
3. Save the configuration file and restart the NGINX service to apply the changes.

## Notes

- You can modify the `upstream` block to add or remove servers from the pool as needed.
- Ensure that the application servers in the `upstream` block are configured to handle the incoming requests from NGINX.