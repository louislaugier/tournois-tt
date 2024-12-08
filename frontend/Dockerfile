FROM node:20.18.1-alpine as builder

WORKDIR /tournois-tt/frontend

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Use a lightweight server to serve static files
FROM node:20.18.1-alpine
WORKDIR /app
COPY --from=builder /tournois-tt/frontend/dist .
RUN npm install -g serve

CMD ["serve", "-s", ".", "-l", "3000"]