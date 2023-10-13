## Setting up the development environment
1. Clone the repository:
```
git clone https://github.com/Eugene-Lek/onecv-go-backend.git
cd onecv-go-backend
```

2. Refer to `.env.example` and create a `.env` file **with the same format** at the same directory level
as `.env.example`

3. Download and install dependencies.
```
$ go get .
```

4. Setup the PostgreSQL database (using pgAdmin or another tool)
   * Create the "onecvtest" database owned by "postgres"/root account
   * Add your local database URL to the .env created in step 1. 
   (Format: "user=postgres password=[PASSWORD] host=localhost port=5432 dbname=onecvtest")

5. Run `init_database.sql` via the Query Tool to set up the database tables and relations.

6. Run the API server:
```
$ go run .
```

7. Run the unit tests:
```
$ go test
```