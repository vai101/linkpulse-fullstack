# --- Stage 1: The "builder" ---
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files from the root of the build context
COPY package*.json ./
RUN npm install

# Copy the rest of the source code from the root of the build context
COPY . .
RUN npm run build

# --- Stage 2: The "server" ---
FROM nginx:stable-alpine

# Copy the built files from the correct dist folder
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]