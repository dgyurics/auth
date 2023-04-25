# FROM nginx:stable-alpine3.17
FROM nginx:1.21.3

# Copy the nginx configuration file to the container
COPY nginx.conf /etc/nginx/nginx.conf

# Expose ports 80 and 443 to the outside world
EXPOSE 80
EXPOSE 443

# Start nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]
