# NASA Asteroid Tracker

### A Cloud-Native Microservices Pipeline

![Go](https://img.shields.io/badge/Go-1.24.0-00ADD8?logo=go)
![Python](https://img.shields.io/badge/Python-3.9-3776AB?logo=python)
![Kubernetes](https://img.shields.io/badge/Kubernetes-v1.27-326CE5?logo=kubernetes)
![AWS](https://img.shields.io/badge/AWS-SQS-232F3E?logo=amazon-aws)
![Postgres](https://img.shields.io/badge/PostgreSQL-15-4169E1?logo=postgresql)

## Overview
This project is a distributed microservices system that tracks Near-Earth Objects in real-time. It ingests data from NASA's API, calculates impact trajectories using a custom physics engine, and persists the data for analysis.

The system is containerized with Docker and orchestrated via Kubernetes, utilizing AWS SQS for asynchronous inter-service communication.

---

## Architecture
This project uses Event-Driven Architecture to decouple ingestion from processing.

```mermaid
graph LR
    A[NASA API] -->|JSON| B(Ingestion Service <br/> Python)
    B -->|Raw Data| D[(Postgres DB)]
    B -->|Message| C{AWS SQS <br/> Queue}
    C -->|Trigger| E(Simulation Engine <br/> Go)
    E -->|Physics Data| D
    User((User)) -->|HTTP Request| F[API Gateway]
    F -->|Route| G(API Service <br/> Go)
    G -->|Query| D
```

| Database Verification |
| --------------------- | 
<img width="1481" alt="DB Proof" src="https://github.com/user-attachments/assets/23ea964f-3da0-4910-ac57-c823d78e0c4d" />

| Simulation Logs       |
| --------------------- | 
<img width="1477" height="379" alt="Simulation Proof" src="https://github.com/user-attachments/assets/dc3ad412-86eb-449d-b329-24cf7a2da793" />

| API-Gateway           |
| --------------------- | 
<img width="1394" height="157" alt="image" src="https://github.com/user-attachments/assets/a7e70a87-1068-4c70-8224-7532d76da107" />

## 🛠️ Installation & Setup

> **Note:** To run this project locally, you will need a NASA API Key and AWS Credentials (SQS access).

### 1. Prerequisites
* Docker Desktop & Kubernetes
* Go 1.24.0+ & Python 3.9+
* `kubectl` CLI tool

### 2. Environment Configuration
Create a `.env` file in the `services/ingestion-service` directory:
```ini
NASA_API_KEY=YOUR_KEY_HERE
AWS_ACCESS_KEY_ID=YOUR_AWS_KEY
AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET
DB_CONNECTION_STRING=postgres://...
# Deploy all microservices to Kubernetes
kubectl apply -f infrastructure/k8s/


