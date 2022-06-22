FROM mysql:latest

ENV MYSQL_ROOT_PASSWORD=password

COPY ./database_merchantapi.sql /docker-entrypoint-initdb.d/

CMD ["mysqld"]