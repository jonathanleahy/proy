# Test Data for REST API v1

This file contains mock data used for testing the REST API endpoints.

## User Data

```json
{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "address": {
    "street": "Kulas Light",
    "suite": "Apt. 556",
    "city": "Gwenborough",
    "zipcode": "92998-3874",
    "geo": {
      "lat": "-37.3159",
      "lng": "81.1496"
    }
  },
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org",
  "company": {
    "name": "Romaguera-Crona",
    "catchPhrase": "Multi-layered client-server neural-net",
    "bs": "harness real-time e-markets"
  }
}
```

## Posts Data

```json
[
  {
    "userId": 1,
    "id": 1,
    "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
    "body": "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto"
  },
  {
    "userId": 1,
    "id": 2,
    "title": "qui est esse",
    "body": "est rerum tempore vitae\nsequi sint nihil reprehenderit dolor beatae ea dolores neque\nfugiat blanditiis voluptate porro vel nihil molestiae ut reiciendis\nqui aperiam non debitis possimus qui neque nisi nulla"
  },
  {
    "userId": 1,
    "id": 3,
    "title": "ea molestias quasi exercitationem repellat qui ipsa sit aut",
    "body": "et iusto sed quo iure\nvoluptatem occaecati omnis eligendi aut ad\nvoluptatem doloribus vel accusantium quis pariatur\nmolestiae porro eius odio et labore et velit aut"
  }
]
```

## Todos Data

```json
[
  {
    "userId": 1,
    "id": 1,
    "title": "delectus aut autem",
    "completed": false
  },
  {
    "userId": 1,
    "id": 2,
    "title": "quis ut nam facilis et officia qui",
    "completed": false
  },
  {
    "userId": 1,
    "id": 3,
    "title": "fugiat veniam minus",
    "completed": false
  },
  {
    "userId": 1,
    "id": 4,
    "title": "et porro tempora",
    "completed": true
  },
  {
    "userId": 1,
    "id": 5,
    "title": "laboriosam mollitia et enim quasi adipisci quia provident illum",
    "completed": false
  }
]
```

## Expected Endpoint Responses

### Endpoint 1: GET /api/user/:id

Simple user fetch - returns user data as-is from external API.

**Expected Response:**
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

### Endpoint 2: GET /api/user/:id/summary

Fetches user + posts, returns summary with post count and recent post titles.

**Expected Response:**
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "postCount": 3,
  "recentPosts": [
    "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
    "qui est esse",
    "ea molestias quasi exercitationem repellat qui ipsa sit aut"
  ],
  "summary": "User Leanne Graham has written 3 posts"
}
```

### Endpoint 3: POST /api/user/:id/report

Fetches user + posts + todos in parallel, accepts body params for filtering.

**Request Body:**
```json
{
  "includeCompleted": true,
  "maxPosts": 2
}
```

**Expected Response:**
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "stats": {
    "totalPosts": 3,
    "totalTodos": 5,
    "completedTodos": 1,
    "pendingTodos": 4,
    "completionRate": "20.0%"
  },
  "posts": [
    {
      "id": 1,
      "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
      "preview": "quia et suscipit\nsuscipit recusandae..."
    },
    {
      "id": 2,
      "title": "qui est esse",
      "preview": "est rerum tempore vitae..."
    }
  ],
  "todos": {
    "pending": [
      "delectus aut autem",
      "quis ut nam facilis et officia qui",
      "fugiat veniam minus",
      "laboriosam mollitia et enim quasi adipisci quia provident illum"
    ],
    "completed": [
      "et porro tempora"
    ]
  },
  "generatedAt": "2025-11-01T15:00:00.000Z"
}
```

## External API Endpoints Used

- **JSONPlaceholder API** (https://jsonplaceholder.typicode.com)
  - GET /users/:id - Fetch user data
  - GET /posts?userId=:id - Fetch user's posts
  - GET /todos?userId=:id - Fetch user's todos
