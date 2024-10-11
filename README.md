
# log-producer-aggregator

## Overview
- This project exposes a server with two endpoints: GET `/{short_url}` and POST `/shorten`.
- Upon recieving a short url, the server pushes the request to the worker pool and a worker makes a database request to retrieve the long url. It returns a redirect with it.
- Upon recieving a request with a long_url in the body, the server pushes the request to the worker pool and a worker hashes the url, if it already exists, returns the already existing for that long_url.

## Setup

1. Check the Config inside main.go is correct

2. Build the docker containers
```bash
docker-compose build
```

3. Start the container
```bash
docker-compose up -d
```

## Structure - aggregator
### api
- **`handlers.go`***
- Holds the logic for each endpoint.

- **`server.go`**
- Server, database and workerpool setup.

- **`ratelimiter.go`**
- Holds logic for implementing a rate limiter: handles limiting and bursting, cleans up ips on each request also.

### cmd
- **`main.go`**
- Main entry point to the application, handles connecting to the database, starting up the server and closing it based on signal.

### internal
- **`workerpool.go`**
- Workerpool logic, for workers and the pool.
- Logic for processing the job types that are passed through

### storage
- **`database.go`**
- Logic for setting up the database connection.
- Closing the database connection also.
- Handles retries with exponential backoff when connecting to the database.

- **`models.go`**
- Holds the database structure for postgres.
- Unique value for the short_url.

- **`operations.go`**
- A selection of functions for different database operations e.g. log fetching and storing to the database.
- Handles hashing the long_url, utilising retries and a nonce value.
- Transactions are used to store in the database due to concurrency with worker pools.

### utils
- **`log.go`**
- Utils for HTML logic, e.g. decoding body, passing a response back

- **`shared.go`**
- Shared strucs to use throughout the application

## Examples

1. Build the docker containers
```bash
docker-compose build
```

2. Start the container
```bash
docker-compose up -d
```

example post shorten request:
```bash
curl -X POST http://localhost:8005/shorten \ 
-H "Content-Type: application/json" \
-d '{"long_url": "https://www.example.com"}'
```

example get request /{short_url}:
```bash
curl -X GET http://localhost:8005/fe919a40
```