# Build stage
FROM node:24.9.0-alpine3.21 AS builder
WORKDIR /app
# Copy files needed for package manager first for more layers
COPY . ./
RUN npm install 
RUN npm run build

# Runtime stage
FROM nginx:1.29-alpine 
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*
COPY --from=builder /app/dist ./
ENTRYPOINT ["nginx", "-g", "daemon off;"]
