# Go CRUD API

A small REST API for creating, listing, reading, updating, and deleting users with MySQL.

## Requirements

- Go 1.27 or newer
- MySQL 8 or another compatible MySQL server

## Set up MySQL in Docker

The running container `my-sql-container` publishes host port `3307` to
container port `3306`. Since this Go app runs on the host, create its database
and table inside the container:

```powershell
docker exec -it my-sql-container mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS myapp; USE myapp; CREATE TABLE IF NOT EXISTS users (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);"
```

The command prompts for the MySQL root password configured for the container.
If your container has a different name or database user, substitute those values.

## Run

The app runs on the host, so connect to the published host port (`3307`), not
the container's internal port (`3306`):

```powershell
$env:MYSQL_DSN = "root:your-password@tcp(127.0.0.1:3307)/myapp?parseTime=true"
go run .
```

The API listens on `http://localhost:8080`. The server exits with an error if
`MYSQL_DSN` is missing or the database cannot be reached.
If you later run the Go app in Docker on the same Docker network, use the MySQL
container name and internal port instead, for example
`tcp(my-sql-container:3306)`.

## Test the API

### 1. Run the automated tests

These tests use a mock SQL driver. They check the handlers without making
requests to Docker or requiring MySQL credentials:

```powershell
go test ./...
```

### 2. Prepare MySQL

Make sure the MySQL container is running and the `myapp.users` table exists.
The setup command in [Set up MySQL in Docker](#set-up-mysql-in-docker) creates
both. Your container publishes host port `3307`.

### 3. Start the API

Open a PowerShell terminal in the project directory. Set the DSN to the actual
root password configured for your container, then run the server:

```powershell
$env:MYSQL_DSN = "root:your-password@tcp(127.0.0.1:3307)/myapp?parseTime=true"
go run .
```

Leave this terminal open. Startup should print `Connected to MySQL` and
`Server running on http://localhost:8080`. If either message is missing, see
[Troubleshooting](#troubleshooting).

### 4. Run the Postman collection

Import [`postman_collection.json`](postman_collection.json) into Postman:

1. Select **Import** and choose the collection file from this project.
2. Open the **Go CRUD API** collection. Its `baseUrl` collection variable is
	already set to `http://localhost:8080`.
3. Open **Runner**, select the collection, and run all requests in order.

The collection creates a user with a unique email, saves its ID, then lists,
reads, updates, and deletes it. Each request checks its expected status and
response; the final request confirms that the deleted user returns `404`.

### 5. Send CRUD requests with PowerShell

Open a second PowerShell terminal in the project directory. Run these commands
in order. They use a unique email address each time, so the create request can
be repeated without colliding with an earlier test.

Set the base URL and test email:

```powershell
$baseUrl = "http://localhost:8080"
$testEmail = "api-test-$([guid]::NewGuid().ToString('N'))@example.com"
```

**Create a user** with `POST /users`. The response should contain the new
user's generated `id`, `name`, and `email`; save the ID for the next requests.

```powershell
$created = Invoke-RestMethod -Method Post -Uri "$baseUrl/users" -ContentType "application/json" -Body (@{ name = "API Test User"; email = $testEmail } | ConvertTo-Json)
$created
$userId = $created.id
```

Expected status: `201 Created`.

**List users** with `GET /users`. The result should include the new user:

```powershell
Invoke-RestMethod -Method Get -Uri "$baseUrl/users" | Format-Table id, name, email, created_at
```

Expected status: `200 OK`.

**Read that user** with `GET /users/{id}`:

```powershell
Invoke-RestMethod -Method Get -Uri "$baseUrl/users/$userId"
```

Expected status: `200 OK`; the returned `id` should equal `$userId`.

**Update the user** with `PUT /users/{id}`. The API replaces the name and email
with the values in the JSON body:

```powershell
$updated = Invoke-RestMethod -Method Put -Uri "$baseUrl/users/$userId" -ContentType "application/json" -Body (@{ name = "Updated API Test User"; email = $testEmail } | ConvertTo-Json)
$updated
```

Expected status: `200 OK`; the returned name should be `Updated API Test User`.

**Delete the user** with `DELETE /users/{id}`:

```powershell
curl.exe -i -X DELETE "$baseUrl/users/$userId"
```

Expected status: `204 No Content`, with an empty response body. Confirm the
record is gone:

```powershell
curl.exe -i "$baseUrl/users/$userId"
```

Expected status: `404 Not Found` and response text `User not found`.

## API reference

All request and response bodies use JSON. `POST` and `PUT` require `name` and
`email` fields.

| Method | Path | Success status | Description |
| --- | --- | --- | --- |
| `GET` | `/users` | `200 OK` | List users |
| `POST` | `/users` | `201 Created` | Create a user |
| `GET` | `/users/{id}` | `200 OK` | Read one user |
| `PUT` | `/users/{id}` | `200 OK` | Replace a user's name and email |
| `DELETE` | `/users/{id}` | `204 No Content` | Delete a user |

Example JSON user:

```json
{
	"id": 1,
	"name": "API Test User",
	"email": "api-test@example.com",
	"created_at": "2026-10-08 12:00:00"
}
```

## Troubleshooting

- **`MYSQL_DSN must be set`**: Set `$env:MYSQL_DSN` in the same PowerShell
	terminal where you run `go run .`.
- **Database connection refused**: Confirm the container is running with
	`docker ps`, and that `docker port my-sql-container 3306/tcp` shows host port
	`3307`. From the host, the DSN must use `127.0.0.1:3307`.
- **Access denied**: Check the username and password in `MYSQL_DSN` against the
	credentials configured for the MySQL container.
- **Unknown database or table**: Run the MySQL setup command above for the
	correct container and database.
- **Port 8080 already in use**: Stop the process using port `8080` before
	starting the API.