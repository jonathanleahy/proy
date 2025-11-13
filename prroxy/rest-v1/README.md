# REST API v1 - Legacy System

A TypeScript REST API server built with TDD (Test-Driven Development) methodology. This serves as the legacy system (version 1) for migration testing with the proxy tool.

## Features

- ✅ **80%+ Test Coverage** - Built with TDD from the ground up
- ✅ **Three Endpoints** - Progressive complexity demonstration
- ✅ **External API Integration** - Calls to JSONPlaceholder API
- ✅ **TypeScript** - Full type safety
- ✅ **Express** - Lightweight and fast

## Endpoints

### 1. GET `/api/user/:id` - Simple

Fetches user data from external API and returns simplified response.

**Example:**
```bash
curl http://0.0.0.0:3000/api/user/1
```

**Response:**
```json
{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org"
}
```

### 2. GET `/api/user/:id/summary` - Medium Complexity

Fetches user and their posts, then aggregates into a summary.

**Example:**
```bash
curl http://0.0.0.0:3000/api/user/1/summary
```

**Response:**
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "postCount": 10,
  "recentPosts": ["Post title 1", "Post title 2", "..."],
  "summary": "User Leanne Graham has written 10 posts"
}
```

### 3. POST `/api/user/:id/report` - Complex

Makes 3 parallel API calls (user, posts, todos) and generates comprehensive report.

**Example:**
```bash
curl -X POST http://0.0.0.0:3000/api/user/1/report \
  -H "Content-Type: application/json" \
  -d '{"includeCompleted": true, "maxPosts": 2}'
```

**Request Body:**
```json
{
  "includeCompleted": true,
  "maxPosts": 2
}
```

**Response:**
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "stats": {
    "totalPosts": 10,
    "totalTodos": 20,
    "completedTodos": 10,
    "pendingTodos": 10,
    "completionRate": "50.0%"
  },
  "posts": [
    {
      "id": 1,
      "title": "Post title",
      "preview": "Post body preview..."
    }
  ],
  "todos": {
    "pending": ["Todo 1", "Todo 2"],
    "completed": ["Todo 3", "Todo 4"]
  },
  "generatedAt": "2025-11-01T15:00:00.000Z"
}
```

## Installation

```bash
# Install dependencies
npm install

# Run tests with coverage
npm test

# Start development server
npm run dev

# Build for production
npm run build

# Run production server
npm start
```

## Development

### Running Tests

```bash
# Run all tests with coverage
npm test

# Watch mode for TDD
npm run test:watch

# Check coverage threshold (80%)
npm test -- --coverage
```

### Project Structure

```
rest-v1/
├── src/
│   ├── routes/         # API route handlers
│   ├── services/       # Business logic layer
│   ├── types/          # TypeScript interfaces
│   ├── app.ts          # Express app configuration
│   └── server.ts       # Server entry point
├── tests/
│   ├── userService.test.ts
│   ├── userRoutes.test.ts
│   └── app.test.ts
├── test-data.md        # Mock data reference
├── package.json
├── tsconfig.json
└── jest.config.js
```

## Test Coverage

This project maintains 80%+ test coverage across:
- Service layer (business logic)
- Route handlers (API endpoints)
- Application setup
- Error handling

```bash
# View coverage report
npm test
open coverage/lcov-report/index.html
```

## External Dependencies

- **JSONPlaceholder API** - https://jsonplaceholder.typicode.com
  - Used for demonstration purposes
  - Provides fake user, post, and todo data

## Use Case

This API serves as the legacy system (CRMv1) in migration testing scenarios. It demonstrates:
1. **Simple external API calls** (Endpoint 1)
2. **Sequential API calls with data manipulation** (Endpoint 2)
3. **Parallel API calls with complex business logic** (Endpoint 3)

The proxy tool can record interactions with this API to facilitate migration to a new system while ensuring 100% compatibility.

## Environment Variables

```bash
PORT=3000  # Server port (default: 3000)
```

## Health Check

```bash
curl http://0.0.0.0:3000/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-11-01T15:00:00.000Z"
}
```

## Error Handling

All endpoints include proper error handling:
- **400** - Invalid input (e.g., invalid user ID)
- **404** - Resource not found
- **500** - Server error

Example error response:
```json
{
  "error": "User not found"
}
```

## Notes

- Built with TDD methodology
- All tests written before implementation
- Comprehensive test coverage ensures reliability
- Ideal for demonstrating API migration strategies
