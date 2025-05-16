# Chat Package

The Chat package provides real-time messaging capabilities for the GreenTrade application, including both REST API endpoints for message management and WebSocket functionality for real-time updates.

## Core Components

### WebSocket Management

The WebSocket implementation handles:

1. **Client Connections**: Manages WebSocket clients with their connection state
2. **Real-time Messaging**: Enables instant message delivery between users
3. **Connection Tracking**: Maintains a registry of active connections
4. **Rate Limiting**: Prevents abuse by limiting message frequency
5. **Ping/Pong**: Ensures connection health through regular heartbeats

### Conversation Management

The conversation functionality provides:

1. **Conversation Creation**: Establishes chat sessions between users
2. **Conversation Retrieval**: Gets a user's active conversations
3. **Conversation Storage**: Persists conversations in the database

### Message Handling

Message functionality includes:

1. **Message Posting**: Sends messages to conversations
2. **Message Retrieval**: Gets message history for a conversation
3. **Message Storage**: Persistently stores messages

## Security Features

The Chat package implements several security measures:

1. **Rate Limiting**: Prevents flooding by limiting message frequency
2. **CORS Configuration**: Restricts WebSocket connections to trusted origins
3. **Authentication**: Ensures only authenticated users can access chats
4. **Connection Timeouts**: Automatically closes inactive connections

## Technical Implementation

The WebSocket server uses:

1. Fiber's WebSocket module for connection handling
2. Concurrent-safe maps with mutex locks for connection tracking
3. JSON for message serialization
4. Timeouts and pings to maintain connection health
