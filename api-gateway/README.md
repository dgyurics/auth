# Nginx Load Balancing Configuration

This is a sample Nginx configuration file for load balancing between multiple upstream servers. The configuration includes:

- `worker_processes` directive to set the number of worker processes to 10, which matches the number of CPU cores on the server.
- `worker_connections` directive to set the maximum number of connections each worker process can handle to 1024.
- Definition of two upstream servers: `auth_servers` and `api_servers`.
- Configuration of the server to listen on both IPv4 and IPv6 addresses.
- Proxying of connections to the upstream servers using the `proxy_pass` directive.
- Configuration of timeouts for establishing connections and sending data to the upstream servers.

This configuration also includes a sample location block for authenticating user requests to the `/auth/user` endpoint, and another location block for handling requests to the `/auth/` and `/api/` endpoints, both of which proxy requests to the defined upstream servers. 

Please note that this configuration is meant to serve as a starting point and should be customized to meet the specific needs of your application.
