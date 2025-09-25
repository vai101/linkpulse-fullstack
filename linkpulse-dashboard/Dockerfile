# linkpulse-dashboard/Dockerfile

# --- Stage 1: The "builder" ---
# Start with a Node.js image to build our React app
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the source code
COPY . .

# Run the build script to create optimized static files
RUN npm run build

# --- Stage 2: The "server" ---
# Start with a lightweight Nginx web server image
FROM nginx:stable-alpine

# Copy ONLY the built static files from the builder's /app/dist directory
# into the default directory Nginx serves files from.
COPY --from=builder /app/dist /usr/share/nginx/html

# Nginx listens on port 80 by default
EXPOSE 80

# The command to start the Nginx server
CMD ["nginx", "-g", "daemon off;"]