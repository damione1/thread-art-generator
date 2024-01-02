FROM node:18 AS builder

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy frontend source files
COPY . .

# Build your frontend
RUN npm run build

# Use nginx to serve the static content
FROM nginx:alpine

# Copy the built app to the web server's root directory
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose the port nginx is listening on
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
