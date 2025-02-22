# rcoder
RCoder: A Simple Remote Code Execution Platform

# Project Title: Real-Time Code Execution Platform

## Overview

This project is a real-time code execution platform that enables users to write, execute, and view the output of code snippets directly from a web-based interface. The system is built with a React frontend, a Go backend, and utilizes Apache Kafka for managing code execution tasks. Docker Compose is employed to orchestrate the multi-container setup, ensuring seamless integration and deployment of all services.

## Architecture

The platform comprises the following components:

1. **Frontend**: A React application that provides an interactive code editor (integrated with Monaco Editor) for users to write and submit code.

2. **Backend**: A Go server that exposes RESTful APIs to handle code execution requests from the frontend. It communicates with Kafka to manage the distribution and processing of code execution tasks.

3. **Kafka**: Serves as the message broker, facilitating communication between the backend and language-specific executors.

4. **Language Executors**: Microservices responsible for executing code in specific programming languages. Each executor listens to a designated Kafka topic, executes the received code, and sends the output back through another Kafka topic.

## Prerequisites

Before running the project, ensure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/)

- [Docker Compose](https://docs.docker.com/compose/install/)

## Setup and Running the Application

Follow these steps to set up and run the application:

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/hdarweesh/rcoder.git
   cd rcoder
   ```


2. **Configure Environment Variables**:

   Create a `.env` file in the project root directory to define necessary environment variables. Refer to the `.env.example` file for the required variables.

3. **Build and Start Services**:

   Use Docker Compose to build and start all services:

   ```bash
   docker-compose up --build
   ```


   This command will build the Docker images (if not already built) and start the containers for the frontend, backend, Kafka, Zookeeper, and language executors.

4. **Access the Application**:

   Once all services are running:

   - **Frontend**: Navigate to `http://localhost:3000` in your web browser to access the code editor interface.

   - **Backend API**: Available at `http://localhost:8080`.

   - **Kafka**: Running internally; no direct access required.

## Project Structure

The repository is organized as follows:

```
realtime-code-execution-platform/
├── backend/
│   ├── Dockerfile
│   ├── main.go
│   └── ...
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── src/
│   └── ...
├── executors/
│   ├── python-executor/
│   │   ├── Dockerfile
│   │   └── app.py
│   ├── javascript-executor/
│   │   ├── Dockerfile
│   │   └── app.js
│   └── ...
├── docker-compose.yml
└── README.md
```

- **backend/**: Contains the Go backend server code and its Dockerfile.

- **frontend/**: Holds the React frontend application and its Dockerfile.

- **executors/**: Includes subdirectories for each language executor, each with its own code and Dockerfile.

- **docker-compose.yml**: Defines the multi-container Docker application, specifying how the services are connected and configured.

## Configuration Details

- **Frontend**:

  - Runs on port 3000.

  - Communicates with the backend at `http://backend:8080` (as defined in Docker Compose networking).

- **Backend**:

  - Runs on port 8080.

  - Exposes endpoints to handle code execution requests.

  - Publishes code execution tasks to Kafka topics.

- **Kafka and Zookeeper**:

  - Kafka broker listens on port 9092.

  - Zookeeper listens on port 2181.

  - Configured within the Docker Compose file to facilitate inter-service communication.

- **Language Executors**:

  - Each executor subscribes to a specific Kafka topic to receive code execution tasks.

  - After execution, results are sent back to a designated Kafka topic for the backend to process.

## Adding Support for New Languages

To add a new language executor:

1. **Create a New Executor Directory**:

   In the `executors/` directory, create a folder named after the new language (e.g., `ruby-executor/`).

2. **Implement the Executor Service**:

   Within this new directory:

   - Develop a script (e.g., `app.rb`) that listens to the appropriate Kafka topic, executes the received code, and returns the output.

   - Create a `Dockerfile` to containerize this executor service.

3. **Update Docker Compose**:

   Modify the `docker-compose.yml` file to include the new executor service, ensuring it connects to the appropriate network and Kafka broker.

4. **Configure Kafka Topics**:

   Ensure that the new executor listens to a unique Kafka topic for incoming code and publishes results to another topic.
