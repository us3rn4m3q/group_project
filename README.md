Website Parsing Microservices

This project implements a microservice-based architecture for parsing data from various websites. The system is built with Docker and consists of multiple independent services.

 Features

Parallel parsing of multiple websites

REST API for managing parsers

Persistent storage of parsing history and results

Full containerization of all components

Centralized routing through API Gateway

Scalable architecture

🏗 Architecture
├── api-gateway/          # Main service (HTTP routing, aggregation)
├── bitshop/              # Microservice 1 (data parsing)
├── jetman/               # Microservice 2 (data parsing)
├── xcore/                # Microservice 3 (data parsing)
├── ram/                  # Microservice 4 (data parsing)
├── docker-compose.yml    # Docker configuration
└── README.md             # This file


Docker

Docker Compose

Go 1.23.2+

Run the project

Build and start containers:

docker-compose up --build


Check if the server is running:

curl http://localhost/api/v1/status


Example requests:

curl -X POST "http://localhost/api/v1/search?query=..."
curl -X POST "http://localhost/api/v1/searchAll?query=..."

 API Endpoints

POST /api/v1/search — Search a single service

POST /api/v1/searchAll — Aggregate results from all services

Area for Improvements

Use configuration files instead of hardcoded values

Host databases in separate Docker containers instead of a single instance

Add centralized logging and monitoring (e.g., Prometheus + Grafana)

Implement retries and fault tolerance for failed parsers

Introduce message queue (e.g., Kafka, RabbitMQ) for asynchronous processing

Add authentication and rate-limiting to the API Gateway

Provide Helm charts for Kubernetes deployment
