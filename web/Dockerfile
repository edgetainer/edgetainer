FROM node:20-alpine AS build

WORKDIR /app

# Copy package.json and package-lock.json to leverage Docker caching
COPY package*.json ./
RUN npm ci

# Copy the rest of the application
COPY . .

# Build the application
RUN npm run build

# Create production image
FROM nginx:alpine

# Copy the built app to the Nginx server
COPY --from=build /app/dist /usr/share/nginx/html

# Copy custom Nginx config if needed
# COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start Nginx server
CMD ["nginx", "-g", "daemon off;"]
