# FROM nginx:stable-alpine3.17
FROM nginx:1.21.3

# Copy the nginx configuration file to the container
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
EXPOSE 443

# Start nginx in the foreground
CMD ["nginx", "-g", "daemon off;"]
