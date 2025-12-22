# Go REST DB API

A simple Go-based REST API to connect and query multiple database types (MySQL, PostgreSQL, SQL Server).

## Prerequisites

- [Go](https://golang.org/dl/) (version 1.18 or higher recommended)
- Access to a supported database (MySQL, PostgreSQL, or SQL Server)

## Setup

Before building the project, you need to initialize the Go module and download dependencies:

```bash
go mod init github.com/user/go-db-rest-api
go mod tidy
```

## Command-Line Arguments

You can configure a default database connection using command-line flags when starting the API.

| Flag | Description | Example |
|------|-------------|---------|
| `-db-type` | Database type (`mysql`, `postgres`, `sqlserver`) | `-db-type=postgres` |
| `-db-host` | Database host address | `-db-host=localhost` |
| `-db-port` | Database port | `-db-port=5432` |
| `-db-user` | Database username | `-db-user=admin` |
| `-db-password` | Database password | `-db-password=secret` |
| `-db-name` | Database name | `-db-name=mydb` |
| `-port` | Server port | `-port=9090` |

## Build Instructions

### Windows
Open PowerShell or Command Prompt and run:
```powershell
go build -o api.exe main.go
```

### macOS / Ubuntu
Open Terminal and run:
```bash
go build -o api main.go
```

## Run Instructions

### Windows (PowerShell)
Run the binary with flags:
```powershell
.\api.exe -db-type="postgres" -db-host="localhost" -db-port=5432 -db-user="user" -db-password="password" -db-name="dbname"
```

### macOS / Ubuntu
Run the binary with flags:
```bash
./api -db-type=postgres -db-host=localhost -db-port=5432 -db-user=user -db-password=password -db-name=dbname
```

## API Endpoints

- `POST /connect`: Establish a persistent connection.
- `POST /query`: Execute a SQL query (stateless, handles its own connection).
- `DELETE /disconnect/:id`: Close a specific persistent connection.
- `GET    /health`: Check API status.

## Troubleshooting

### Error: "address already in use"
This error means another process is already using the port you specified.

#### macOS / Ubuntu
To find the process ID (PID) using the port (e.g., 9090):
```bash
lsof -i :9090
```
To kill the process:
```bash
kill -9 <PID>
```

#### Windows (PowerShell)
To find the process ID (PID) using the port (e.g., 9090):
```powershell
netstat -ano | findstr :9090
```
To kill the process:
```powershell
Stop-Process -Id <PID> -Force
```

### Query Example (Stateless)
```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "type": "postgres",
    "host": "localhost",
    "port": 5432,
    "user": "user",
    "password": "password",
    "dbname": "mydb",
    "query": "SELECT * FROM users WHERE id = 1"
  }'
```
