FROM postgres:16

# Initialize database from migration script
COPY up.sql /docker-entrypoint-initdb.d/1.sql