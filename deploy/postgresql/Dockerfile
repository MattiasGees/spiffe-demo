FROM postgres:16.3
COPY init-user-db.sh /docker-entrypoint-initdb.d/init-user-db.sh
COPY set-pg_hba.sh /docker-entrypoint-initdb.d/set-pg_hba.sh
COPY pg_hba.conf /pg_hba.conf
