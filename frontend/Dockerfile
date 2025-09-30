#DEV SETUP - UNCOMMENT THE VOLUMES IN DOCKER-COMPOSE IN CASE OF IT NOT WORKING

# Stage 1: Build the React app
# FROM node:18 AS build
# WORKDIR /app
# COPY package.json package-lock.json ./
# RUN npm install
# COPY . .
# RUN npm run build

# # Stage 2: Serve the app with Nginx
# FROM alpine AS output
# WORKDIR /dist
# COPY --from=build /app/dist .
# COPY --from=build /app/dist /usr/share/nginx/html

# # Uncomment these commands to run the Vite development server (hot reload)
# EXPOSE 8081
# CMD ["npm", "run", "dev"]

# Stage 1: Build
FROM node:20 AS builder
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install

COPY . .
RUN npm run build -- --output-path=./dist/frontend --configuration=production

# ----------------------------
# Stage 2: Production
FROM node:20-alpine
WORKDIR /usr/share/nginx/html

# Copy only what's needed
COPY --from=builder /app/dist/frontend .

ENV HOST=0.0.0.0
EXPOSE 8081

CMD ["npx", "serve", "-s", "./", "--listen", "8081", "--no-compression"]
