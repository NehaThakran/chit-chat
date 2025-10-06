# GoChat – Real-Time Chat App in Go

GoChat is a blazing-fast real-time chat application built with Go and WebSockets. Designed for scalability and responsiveness, it supports private messaging, typing indicators, and message persistence, etc. The frontend is built with React, making it easy to customize and extend.

## Features

- Real-time communication using WebSockets
- Private messaging between users
- Typing indicators for active conversations
- Message persistence with a database (MongoDB)
- Offline Message Delivery
- Chat History Loading
- Group Chat Rooms (Basic)

## Tech Stack

- **Backend**: Go (Golang), Gorilla WebSocket, net/http
- **Frontend**: React
- **Database**: MongoDB
- **Deployment**: Docker

## Getting Started

1. Clone the repo
2. Install dependencies
3. Run the Go server - go run main.go db.go
4. Launch the frontend - npm start

or directly run using docker:
docker-compose up --build

## Future Enhancements

- Read receipts
- Emoji support
- File sharing
- Push notifications
- User authentication and session management


