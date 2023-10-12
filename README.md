## Setting up the development environment

1. Refer to `.env.example` and create a `.env` file **with the same format** at the same directory level
as `.env.example`

2. Download and install dependencies.
```
$ go get .
```

3. Setup the PostgreSQL database (using pgAdmin or another tool)
   * Create the "onecvtest" database owned by "postgres"/root account
   * Add your local database URL to the .env created in step 1. 
   (Format: "user=postgres password=[PASSWORD] host=localhost port=5432 dbname=onecvtest")

4. Run `init_database.sql` via the Query Tool to set up the database tables and relations.

5. Then `go run .` to run.

6. Run `go test` to run the unit tests. 