#!/bin/bash

# Start API in the background
cd api
go run ./cmd/main.go &
API_PID=$!

# Start frontend in the background
cd ../frontend
npm start &
FRONTEND_PID=$!

# Wait for both processes
wait $API_PID $FRONTEND_PID 