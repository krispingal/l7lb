# L7 Load Balancer

![build passing](https://github.com/krispingal/l7lb/actions/workflows/build.yml/badge.svg?event=push)

This is a custom Layer 7 (L7) load balancer implemented in Go (Golang) that supports HTTP traffic routing, SSL termination, path-based routing, request latency tracking, and rate limiting. It can balance requests across multiple backend servers and manage different backend groups per API endpoint.

## Features

1. Round-robin Load Balancing: Distributes incoming requests evenly across multiple backend servers.
2. Health Checks: Periodic health checks for each backend server to ensure requests are only routed to healthy servers.
3. SSL Termination: Terminates SSL connections and forwards the unencrypted requests to backend servers.
4. Path-based Routing: Routes requests based on URL paths, allowing different backend groups to handle different API endpoints.
5. Request Latency Tracking: Logs request latency and response codes for each request.
6. Rate Limiting: Limits the number of requests from each client IP within a defined time window using token bucket rate limiting algorithm.

## Usage

### Running the load balancer

```sh
go install
go run cmd/api/main.go
```

The load balancer will start on port 8443. You can modify the configuration in main.go to adjust backend groups and routes.

### Running backend servers

Change directory into `backends` directory and spin up servers.

```sh
BACKEND_RESPONSE="Hello from backend 1" go run backend_server.go --port 8081
```

Repeat this for the rest of the servers with different port numbers and backend responses.

### Testing the load balancer

#### Basic HTTP request

```sh
curl -k https://localhost:8443/apiA
```

#### Test rate limiting

```sh
for i in {1..110}; do
  curl -k https://localhost:8443/apiA
done
```

## Future Enhancements

1. Caching: Add support for caching frequent responses to reduce load on backend servers.
1. Session Persistence: Implement sticky sessions to route requests from the same client to the same backend.
1. Different routing strategies: Use different routing strategies like least connections or weighted round robin.
1. Circuit breaker: Implement a circuit breaker to stop routing requests to servers that consecutively failures, until it recovers.
1. Request retry policies
1. Dynamic backend registration/removal
