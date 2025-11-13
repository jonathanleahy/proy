# POST /api/user/:id/report - V1 Endpoint Documentation

## Overview
Complex endpoint that generates a comprehensive user report by making 3 parallel API calls and aggregating data with optional filtering.

## Request

**Method:** POST
**Path:** `/api/user/:id/report`
**Content-Type:** application/json

### Path Parameters
- `id` (number, required): User ID

### Request Body (Optional)
```json
{
  "includeCompleted": true,  // optional, default: true
  "maxPosts": 5              // optional, no default (returns all if not specified)
}
```

## Response

**Status Code:** 200 OK
**Content-Type:** application/json

### Response Structure
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "stats": {
    "totalPosts": 10,
    "totalTodos": 20,
    "completedTodos": 11,
    "pendingTodos": 9,
    "completionRate": "55.0%"
  },
  "posts": [
    {
      "id": 1,
      "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
      "preview": "quia et suscipit\nsuscipit recusandae consequuntur expedita..."
    }
  ],
  "todos": {
    "pending": [
      "delectus aut autem",
      "quis ut nam facilis et officia qui"
    ],
    "completed": [
      "fugiat veniam minus",
      "et porro tempora"
    ]
  },
  "generatedAt": "2025-11-12T21:38:20.123Z"
}
```

## V1 Implementation Logic

### API Calls (Parallel using Promise.all)
Makes 3 parallel calls to JSONPlaceholder API:

1. **Get User:**
   ```
   GET https://jsonplaceholder.typicode.com/users/{userId}
   ```

2. **Get Posts:**
   ```
   GET https://jsonplaceholder.typicode.com/posts?userId={userId}
   ```

3. **Get Todos:**
   ```
   GET https://jsonplaceholder.typicode.com/todos?userId={userId}
   ```

### Data Processing

1. **Stats Calculation:**
   - `totalPosts`: Length of posts array
   - `totalTodos`: Length of todos array
   - `completedTodos`: Count of todos where `completed === true`
   - `pendingTodos`: Count of todos where `completed === false`
   - `completionRate`: `(completedTodos / totalTodos * 100).toFixed(1) + '%'` (or "0.0%" if no todos)

2. **Posts Formatting:**
   - If `maxPosts` specified: take first N posts (`posts.slice(0, maxPosts)`)
   - Format each post as: `{id, title, preview: body}`

3. **Todos Formatting:**
   - `pending`: Array of titles from todos where `completed === false`
   - `completed`: Array of titles from todos where `completed === true` (empty array if `includeCompleted === false`)

4. **Timestamp:**
   - `generatedAt`: ISO 8601 timestamp from `new Date().toISOString()`

## Test Requirements

### Required Test Cases

1. **Basic Functionality (User 1, No Options)**
   - Should return 200 status
   - Should include userId, userName, email
   - Should calculate stats correctly
   - Should include all posts
   - Should include both pending and completed todos
   - Should include generatedAt timestamp

2. **Optional Parameters (User 2)**
   - Test with `maxPosts: 3` - should limit posts array
   - Test with `includeCompleted: false` - should have empty completed todos array

3. **Data Validation (User 3)**
   - Verify completionRate calculation is correct
   - Verify post preview matches body field
   - Verify todos are correctly split into pending/completed

### Implementation Notes for V2

- Use parallel HTTP calls (goroutines + channels or similar)
- Handle content encoding (gzip, brotli)
- Parse optional request body
- Calculate completion rate to 1 decimal place
- Generate ISO 8601 timestamp
- Match exact field names and structure from v1
