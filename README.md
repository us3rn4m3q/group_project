#  Website Parsing Microservices

This project implements a **microservice-based architecture** for parsing data from various websites.  
The system is built with **Docker** and consists of multiple independent services.

---

##  Features

-  Parallel parsing of multiple websites  
-  REST API for managing parsers  
-  Persistent storage of parsing history and results  
-  Full containerization of all components  
-  Centralized routing through API Gateway  
-  Scalable architecture  

---

## 🏗 Architecture

├── api-gateway/ # Main service (HTTP routing, aggregation)
├── bitshop/ # Microservice 1 (data parsing)
├── jetman/ # Microservice 2 (data parsing)
├── xcore/ # Microservice 3 (data parsing)
├── ram/ # Microservice 4 (data parsing)
├── docker-compose.yml # Docker configuration
└── README.md # This file

yaml
Copy
Edit

**Tech stack**:  
- 🐳 Docker  
- ⚙️ Docker Compose  
- 💻 Go 1.23.2+  

---

## ▶️ Getting Started

### 1. Build and start containers
bash
docker-compose up --build 
2. Check server status
bash
Copy
Edit
curl http://localhost/api/v1/status
3. Example API requests
bash
Copy
Edit
# Search in a single service
curl -X POST "http://localhost/api/v1/search?query=..."

# Aggregate results from all services
curl -X POST "http://localhost/api/v1/searchAll?query=..."
 API Endpoints
Method	Endpoint	Description
POST	/api/v1/search?query=...	Search in a single service
POST	/api/v1/searchAll?query=...	Aggregate results from all services

Improvements
 Use configuration files instead of hardcoded values

 Host databases in separate Docker containers instead of a single instance

 Add centralized logging & monitoring (e.g., Prometheus + Grafana)

 Implement retries and fault tolerance for failed parsers

 Introduce a message queue (e.g., Kafka, RabbitMQ) for async processing

 Add authentication & rate limiting in API Gateway

 Provide Helm charts for Kubernetes deployment

 License
MIT License. Feel free to use and contribute.

arduino
Copy
Edit
