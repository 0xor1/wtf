FROM mariadb:10.5.9 AS builder

# That file does the DB initialization but also runs mysql daemon, by removing the last line it will only init
RUN ["sed", "-i", "s/exec \"$@\"/echo \"not running $@\"/", "/usr/local/bin/docker-entrypoint.sh"]

ENV MYSQL_ROOT_PASSWORD=root

# generic sqls for user pkg tests
COPY sql/users.sql /docker-entrypoint-initdb.d/users.sql
COPY sql/data.sql /docker-entrypoint-initdb.d/data.sql

# games data sqls
COPY cmd/games/sql/data.sql /docker-entrypoint-initdb.d/games_data.sql

# todo sqls
COPY cmd/todo/sql/users.sql /docker-entrypoint-initdb.d/todo_users.sql
COPY cmd/todo/sql/data.sql /docker-entrypoint-initdb.d/todo_data.sql

RUN ["/usr/local/bin/docker-entrypoint.sh", "mysqld", "--datadir", "/initialized-db", "--aria-log-dir-path", "/initialized-db"]

FROM mariadb:10.5.9

COPY --from=builder /initialized-db /var/lib/mysql