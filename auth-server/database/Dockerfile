FROM postgres:12.9

# set environment variables
ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD postgres
ENV POSTGRES_DB postgres

# copy the init script
COPY init.sql /docker-entrypoint-initdb.d/init.sql

# expose the PostgreSQL port
EXPOSE 5432

# Define a mount point for persisting data
VOLUME [ "/var/lib/postgresql/data" ]
