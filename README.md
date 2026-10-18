# Load Balancing Go Services with Nginx

A minimal setup demonstrating HTTP load balancing and automatic failover using Go, Nginx, and Docker Compose.

A single Go HTTP service returning a JSON list of users is deployed across 5 separate containers (`8001`–`8005`), fronted by an Nginx reverse proxy on port `8080`. Every incoming request is logged to `./logs` so you can inspect traffic routing and failover behavior in real time.

## Architecture

```text
                        +----------------------+
   Browser / Client --->| Nginx Load Balancer  |
   http://localhost:8080|     (Port: 8080)     |
                        +----------+-----------+
                                   | Round-Robin + Failover
         +--------------+----------+---+--------------+--------------+
         v              v              v              v              v
   +-----------+  +-----------+  +-----------+  +-----------+  +-----------+
   | go-app-1  |  | go-app-2  |  | go-app-3  |  | go-app-4  |  | go-app-5  |
   | Port:8001 |  | Port:8002 |  | Port:8003 |  | Port:8004 |  | Port:8005 |
   +-----+-----+  +-----+-----+  +-----+-----+  +-----+-----+  +-----+-----+
         +--------------+--------------+--------------+--------------+
                                       |
                                       v
                             ./logs/hits.log
                             ./logs/nginx-balancer.log
```

## Project Structure

```text
.
├── Dockerfile              # Multi-stage build for the Go service
├── docker-compose.yml      # Runs 5 Go containers (8001-8005) + Nginx (8080)
├── go.mod                  # Go module definition
├── main.go                 # HTTP server returning user JSON and writing hit logs
├── script.sh               # Helper script to send N requests and summarize distribution
├── nginx/
│   └── nginx.conf          # Upstream cluster and failover configuration
└── logs/
    ├── hits.log            # Application-level hit log (which server handled the request)
    └── nginx-balancer.log  # Nginx upstream access log
```

## Getting Started

Build and start all containers in detached mode:

```bash
docker compose up -d --build
```

Verify that all 6 containers are running:

```bash
docker compose ps
```

| Container | Service | Host Port | Description |
| :--- | :--- | :--- | :--- |
| `nginx-load-balancer` | `nginx` | `http://localhost:8080` | Nginx reverse proxy / load balancer |
| `go-app-1` | `app1` | `http://localhost:8001` | Go instance 1 |
| `go-app-2` | `app2` | `http://localhost:8002` | Go instance 2 |
| `go-app-3` | `app3` | `http://localhost:8003` | Go instance 3 |
| `go-app-4` | `app4` | `http://localhost:8004` | Go instance 4 |
| `go-app-5` | `app5` | `http://localhost:8005` | Go instance 5 |

## Usage & Testing

### 1. Browser or curl
Open `http://localhost:8080` (or `http://localhost:8080/users`) in your browser and refresh multiple times. The `served_by` and `server_port` fields will cycle through `app-1` (`8001`) to `app-5` (`8005`).

```json
{
  "status": "success",
  "served_by": "app-1",
  "server_port": "8001",
  "timestamp": "2026-10-18T06:20:00Z",
  "data": [
    {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@example.com",
      "role": "Backend Engineer"
    }
  ]
}
```

### 2. Inspect Logs
Tail the log files in the `./logs` directory to watch which backend instance serves each request:

```bash
tail -f logs/hits.log
```

Output example:
```text
[2026-10-18 06:21:01] HIT -> Server: app-1  (Port: 8001) | Method: GET  | Path: /  | Client: 172.20.0.7:45112
[2026-10-18 06:21:02] HIT -> Server: app-2  (Port: 8002) | Method: GET  | Path: /  | Client: 172.20.0.7:38910
[2026-10-18 06:21:03] HIT -> Server: app-3  (Port: 8003) | Method: GET  | Path: /  | Client: 172.20.0.7:50124
```

For Nginx-level upstream logs:
```bash
tail -f logs/nginx-balancer.log
```

### 3. Send Bulk Requests (`script.sh`)
Use `script.sh` to send `n` consecutive requests to `http://localhost:8080/users` and print a distribution summary.

Interactive prompt:
```bash
./script.sh
```

Or pass the request count directly as an argument:
```bash
./script.sh 25
```

### 4. Failover Testing
Stop one or more backend containers to verify that Nginx automatically routes traffic around unhealthy instances:

```bash
docker stop go-app-2
```

Refresh `http://localhost:8080` or run `./script.sh 20`. Requests will continue succeeding without errors, skipping `app-2`.

Bring the container back online:

```bash
docker start go-app-2
```

## Nginx Configuration Notes

The upstream pool is configured in `nginx/nginx.conf`:

```nginx
upstream go_backend_cluster {
    server app1:8001 max_fails=2 fail_timeout=5s;
    server app2:8002 max_fails=2 fail_timeout=5s;
    server app3:8003 max_fails=2 fail_timeout=5s;
    server app4:8004 max_fails=2 fail_timeout=5s;
    server app5:8005 max_fails=2 fail_timeout=5s;
}
```

- **Round-Robin (default)**: Requests are distributed evenly in sequential order.
- **`max_fails=2 fail_timeout=5s`**: Marks an upstream server as unavailable for 5 seconds after 2 failed connection attempts.
- **`proxy_next_upstream`**: Forwards the request to the next available server if the selected upstream times out or drops the connection.

To test other load balancing algorithms, update `upstream go_backend_cluster` in `nginx/nginx.conf` and run `docker compose restart nginx`:
- `least_conn;` — Routes to the server with the fewest active connections.
- `ip_hash;` — Routes requests from the same client IP to the same backend server.
- `server app1:8001 weight=3;` — Assigns higher traffic weight to specific servers.

## Stopping the Containers

```bash
docker compose down
```
