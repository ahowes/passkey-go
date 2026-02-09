# Standing up postgres in docker for local development

1. `docker pull postgres`
2. `docker volume create postgres_data`
3. `docker run --name postgres_container -e POSTGRES_PASSWORD=mysecretpassword -d -p 5432:5432 -v postgres_data:/var/lib/postgresql/data postgres`

## standing up postgres and pgadmin
`docker compose -f postgres-admin-compose.yaml up -d`

View server admin portal at [http://localhost:5050](http://localhost:5050)

TODO: 
[ ] create code to connect to the DB
[ ] Determine how we can stand up both DB and server in k8s