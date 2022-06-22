FROM mysql:latest

ENV MYSQL_ROOT_PASSWORD=password

COPY ./database_merchantapi.sql /tmp

CMD ["mysqld", "--init-file=/tmp/database_merchantapi.sql" ]