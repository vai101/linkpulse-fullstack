LinkPulse - Full-Stack URL Shortener 🔗
LinkPulse is a full-stack, microservice-based URL shortening application built with Go and React. It features an asynchronous analytics pipeline using AWS SQS to track click metrics in real-time without impacting redirect performance. The entire application is containerized with Docker for easy setup and deployment.

🚀 Live Demo
Frontend Dashboard: [Link to your deployed Vercel URL]

API Endpoint: [Link to your deployed Render URL]

🏛️ Architecture
The project is designed with a decoupled, microservice-based architecture to ensure scalability and resilience.

Frontend (React on Vercel): The user interacts with the dashboard to create new short URLs and view analytics.

API Server (Go on Render): A containerized Go application that handles POST requests to create new links and GET requests to redirect users.

Database (PostgreSQL on Supabase): The primary data store for URL mappings and click analytics.

Message Queue (AWS SQS): When a link is clicked, the API server sends a message to an SQS queue. This is a "fire-and-forget" operation that doesn't slow down the user's redirect.

Worker (Go on Render): A separate, containerized Go application that constantly listens to the SQS queue. When a new message appears, it processes the click and saves the analytic data to the database.

✨ Features
URL Shortening: A robust API to convert long URLs into unique, short codes.

Fast Redirects: A high-performance redirect service.

Asynchronous Click Analytics: A decoupled system to track clicks without latency.

Analytics Dashboard: A simple React frontend to create links and view click counts.

Fully Containerized: All three services (API, Worker, Dashboard) are containerized with Docker and orchestrated with Docker Compose for a simple, one-command local setup.

🛠️ Technology Stack
Frontend: React, Vite, CSS

Backend (API & Worker): Go (Golang)

Database: PostgreSQL (hosted on Supabase)

Message Queue: Amazon SQS (Simple Queue Service)

Containerization: Docker, Docker Compose

Deployment: Vercel (Frontend), Render (Backend)

🚀 Getting Started
Prerequisites
Git

Docker & Docker Compose

Node.js (for running the frontend outside of Docker if needed)

Running Locally
Clone the repository:

Bash

git clone https://github.com/your-username/linkpulse-fullstack.git
cd linkpulse-fullstack
Create your environment file:
Create a .env file in the root of the project by copying the template.

Bash

cp .env.example .env
Now, open the .env file and add your actual secret values from Supabase and AWS.

Run the application:
Use Docker Compose to build the images and start all three services.

Bash

docker-compose up --build
The application will be available at the following addresses:

Dashboard: http://localhost:5173

API Server: http://localhost:8080