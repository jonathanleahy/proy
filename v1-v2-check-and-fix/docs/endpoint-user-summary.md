# GET /api/user/:id/summary

## Purpose
Returns aggregated user information including post count and list of post titles.

## V1 Implementation Details

### Request
```
GET /api/user/:id/summary
```

**Path Parameters:**
- `id` (number): User ID

**Example:**
```bash
curl http://localhost:3002/api/user/1/summary
```

### Response Format

**Success (200 OK):**
```json
{
  "userId": 1,
  "userName": "Leanne Graham",
  "email": "Sincere@april.biz",
  "postCount": 10,
  "recentPosts": [
    "sunt aut facere repellat provident",
    "qui est esse",
    "ea molestias quasi exercitationem",
    "... more post titles ..."
  ],
  "summary": "User Leanne Graham has written 10 posts"
}
```

**Fields:**
- `userId` (number): The user's ID
- `userName` (string): The user's full name
- `email` (string): The user's email address
- `postCount` (number): Total number of posts by this user
- `recentPosts` (string[]): Array of all post titles
- `summary` (string): Formatted summary string

## V1 Implementation Logic

1. **Sequential API Calls:**
   - Call 1: `GET /users/:userId` → Get user data
   - Call 2: `GET /posts?userId=:userId` → Get user's posts

2. **Data Transformation:**
   - Extract all post titles from posts array
   - Count total posts
   - Format summary string: `"User {name} has written {count} posts"`

3. **Error Handling:**
   - Returns 500 on any failure with generic error message

## External API Calls
- **JSONPlaceholder API** via proxy:
  - `https://jsonplaceholder.typicode.com/users/:id`
  - `https://jsonplaceholder.typicode.com/posts?userId=:id`

## V2 Implementation Requirements

To match V1 behavior exactly, V2 must:
1. Make same two sequential API calls through proxy
2. Return identical JSON structure
3. Use same field names (camelCase)
4. Format summary string identically
5. Handle errors appropriately

## Test Requirements

Test must verify:
- ✅ Endpoint returns 200 OK
- ✅ Response has correct JSON structure
- ✅ All fields present and correctly typed
- ✅ `postCount` matches array length
- ✅ `recentPosts` contains all post titles
- ✅ `summary` string formatted correctly
- ✅ Exact match with V1 response for same input
