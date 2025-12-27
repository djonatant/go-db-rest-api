# Application Usage Guide

This application is a REST API database connector. It is distributed as a single binary executable.

## 1. Download

Download the appropriate binary for your operating system from the release bucket.

- **Windows**: `app-windows-amd64.exe`
- **Mac (Intel)**: `app-darwin-amd64`
- **Mac (Apple Silicon)**: `app-darwin-arm64`
- **Linux (Ubuntu)**: `app-linux-amd64`

## 2. Basic Usage

Run the application from your command line (Terminal on Mac/Linux, Command Prompt or PowerShell on Windows).

### Flags
The application accepts the following optional flags to configure a default database connection on startup:

| Flag | Description | Example |
|------|-------------|---------|
| `--port` | Server port (default: 8080) | `--port 3000` |
| `--db-type` | Database type (`mysql`, `postgres`, `sqlserver`) | `--db-type postgres` |
| `--db-host` | Database host address | `--db-host localhost` |
| `--db-port` | Database port | `--db-port 5432` |
| `--db-user` | Database username | `--db-user myuser` |
| `--db-password` | Database password | `--db-password secret` |
| `--db-name` | Database name | `--db-name mydb` |

---

## 3. Platform Specific Instructions

### 🍎 macOS

1.  **Open Terminal** and navigate to the folder where you downloaded the file.
2.  **Make executable**:
    ```bash
    chmod +x app-darwin-arm64  # for Apple Silicon
    # or
    chmod +x app-darwin-amd64  # for Intel Mac
    ```
3.  **Run**:
    ```bash
    ./app-darwin-arm64 --port 8080
    ```
    *Note: If you see a security warning ("developer cannot be verified"), you may need to go to System Settings > Privacy & Security and allow the app to run, or right-click the file in Finder, select Open, and confirm.*

### 🐧 Linux (Ubuntu)

1.  **Open Terminal**.
2.  **Make executable**:
    ```bash
    chmod +x app-linux-amd64
    ```
3.  **Run**:
    ```bash
    ./app-linux-amd64 --port 8080
    ```

### 🪟 Windows

1.  **Open PowerShell** or **Command Prompt**.
2.  Navigate to the download location.
3.  **Run**:
    ```powershell
    .\app-windows-amd64.exe --port 8080
    ```
    *Note: You may need to "Unblock" the file in Properties if Windows Defender SmartScreen allows it.*

---

## 4. Examples

**Start server only (connect via API later):**
```bash
./app-linux-amd64 --port 8080
```

**Start and connect to PostgreSQL:**
```bash
./app-linux-amd64 \
  --port 8080 \
  --db-type postgres \
  --db-host localhost \
  --db-port 5432 \
  --db-user postgres \
  --db-password mysecret \
  --db-name psql_db
```
