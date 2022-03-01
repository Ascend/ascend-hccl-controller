FROM {{HARBOR_IP}}:{{HARBOR_HTTPS_PORT}}/mysql/mysql-server:8.0.26
COPY init.sql /docker-entrypoint-initdb.d/
COPY my.cnf /etc/
