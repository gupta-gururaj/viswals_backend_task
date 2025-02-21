# Viswals Backend Test

## Overview
This project is designed to efficiently process CSV data using a message-driven architecture. The system consists of a producer that reads and publishes data to a RabbitMQ queue, and a consumer that processes the messages and stores them in Redis (for caching) and PostgreSQL (for persistence). Additionally, a REST API is provided for managing and retrieving stored records, ensuring high performance and scalability.

---

## System Components

### Producer
The producer is responsible for reading CSV files, transforming the data into JSON, and sending it to RabbitMQ for further handling.

#### Responsibilities:
- **Reading CSV Files** – Extracts raw data from a given CSV file.
- **Parsing Data** – Converts the extracted data into structured JSON format.
- **Publishing Messages** – Sends processed data to a RabbitMQ queue for processing by the consumer.

---

### Consumer
The consumer listens for messages from RabbitMQ, processes the data, and stores it in both Redis and PostgreSQL. It also provides REST API endpoints to interact with the stored data.

#### Responsibilities:
- **Listening for Messages** – Subscribes to RabbitMQ and retrieves incoming messages.
- **Processing Data** – Formats and prepares the data for storage.
- **Redis Caching** – Stores frequently accessed data in Redis to enhance performance.
- **PostgreSQL Storage** – Saves processed data in a relational database for persistence.
- **Providing APIs** – Exposes RESTful endpoints to manage and retrieve stored records.

---

## API Endpoints

| Endpoint         | Method | Description |
|-----------------|--------|-------------|
| `/users`       | POST   | Adds a new user to the system. |
| `/users/{id}`  | DELETE | Deletes a user from the database based on their ID. |
| `/users`       | GET    | Retrieves a list of users, with optional filtering by name and email.|
| `/users/{id}`  | GET    | Fetches detailed information for a specific user by ID. |
| `/users/sse`   | GET    | Streams user data in real-time via Server-Sent Events (SSE). |

---

## Running the Application

To run the application in a Docker environment, follow these steps:

1. **Start the Services**
   ```bash
   docker-compose up --build
   ```
2. **Access the API**
   - Once running, access the API at:
     ```
     http://localhost:8080
     ```


---

## Additional Information
- Ensure all required environment variables in `docker-compose.yml` are correctly set up before running the application.
- Redis is used for caching frequently accessed data, while PostgreSQL ensures long-term data persistence.
- The system is designed for scalability and can be extended to handle larger datasets or additional message brokers.

---

This documentation provides an overview of the backend system and its functionalities. For further details, refer to the source code and comments within the project files.

