# Configuration Guide

## Available Test Configurations

### config.comprehensive.json ⭐ (Complete Test Suite)

**Purpose**: Complete test suite covering all rest-v1 endpoints

**What it tests** (40 endpoints total):
- `/api/user/:id` - 3 simple user fetches (JSONPlaceholder)
- `/api/user/:id/summary` - 3 user summaries (JSONPlaceholder)
- `/api/user/:id/report` - 3 complex reports (JSONPlaceholder)
- `/api/person?surname=X&dob=Y` - 25 person lookups (rest-external-user)
- `/api/people?surname=X` - 3 surname searches (rest-external-user)
- `/api/people?dob=Y` - 3 dob searches (rest-external-user)

**External calls**:
- ✅ Makes external calls to jsonplaceholder.typicode.com
- ✅ Makes external calls to rest-external-user (port 3006)
- ✅ Tests proxy recording in record mode
- ✅ Tests proxy playback in playback mode

**When to use**:
- Complete API validation
- Comprehensive testing of all endpoints
- End-to-end system testing

**Example**:
```bash
./test-record.sh config.comprehensive.json
```

---

### config.user-endpoints.json (JSONPlaceholder Testing)

**Purpose**: Test user endpoints with JSONPlaceholder API

**What it tests**:
- `/api/user/1` through `/api/user/5` - Get user data
- `/api/user/1-3/summary` - Get user summaries with posts
- `/api/user/1-2/report` - Get comprehensive user reports

**External calls**:
- ✅ Makes 10 external API calls to jsonplaceholder.typicode.com
- ✅ Tests proxy recording in record mode
- ✅ Tests proxy playback in playback mode

**When to use**:
- Testing proxy functionality with JSONPlaceholder
- Testing user endpoint implementations

**Example**:
```bash
./test-record.sh config.user-endpoints.json
```

---

### config.person-lookup.json (Person Lookup Testing)

**Purpose**: Test person lookup endpoints with full search criteria

**What it tests**:
- `/api/person?surname=X&dob=Y` - 25 person lookups
- Returns single person object for exact match

**External calls**:
- ✅ Makes 25 external calls to rest-external-user (port 3006)
- ✅ Tests proxy recording
- ✅ Tests API v1 vs v2 comparison

**When to use**:
- Testing person lookup with exact criteria
- Testing full search functionality

**Example**:
```bash
./test-record.sh config.person-lookup.json
```

---

### config.person-by-surname.json (Surname Search)

**Purpose**: Test partial search by surname only

**What it tests**:
- `/api/people?surname=X` - 5 surname searches
- Returns array of matching people

**External calls**:
- ✅ Makes 5 external calls to rest-external-user
- ✅ Tests partial search functionality

**When to use**:
- Testing surname-only search
- Testing array responses

**Example**:
```bash
./test-record.sh config.person-by-surname.json
```

---

### config.person-by-dob.json (DOB Search)

**Purpose**: Test partial search by date of birth only

**What it tests**:
- `/api/people?dob=Y` - 5 dob searches
- Returns array of matching people

**External calls**:
- ✅ Makes 5 external calls to rest-external-user
- ✅ Tests partial search functionality

**When to use**:
- Testing dob-only search
- Testing array responses

**Example**:
```bash
./test-record.sh config.person-by-dob.json
```

---

## Configuration Structure

```json
{
  "base_url_v1": "http://0.0.0.0:3002",
  "base_url_v2": "http://0.0.0.0:3004",
  "iterations": 1,
  "ignore_fields": ["generatedAt", "timestamp"],
  "endpoints": [
    {
      "path": "/api/endpoint",
      "method": "GET"
    },
    {
      "path": "/api/endpoint",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "key": "value"
      }
    }
  ]
}
```

## Creating Custom Configs

### For External API Testing (with proxy)

Your endpoints should make external HTTP calls:

```typescript
// Good - makes external call through proxy
const response = await axios.get(`${PROXY_URL}/external-api.com/data`);
```

### For Internal Testing (no proxy needed)

Your endpoints use embedded data or local processing:

```typescript
// No external call - uses embedded data
return this.localData.find(item => item.id === id);
```

## Summary

| Config | Endpoints | External Calls | Proxy Records | Use Case |
|--------|-----------|---------------|---------------|----------|
| comprehensive.json | 40 | ✅ Yes (both) | ✅ Yes | Complete test suite |
| user-endpoints.json | 10 | ✅ Yes (JSONPlaceholder) | ✅ Yes | User API testing |
| person-lookup.json | 25 | ✅ Yes (rest-external-user) | ✅ Yes | Person lookup testing |
| person-by-surname.json | 5 | ✅ Yes (rest-external-user) | ✅ Yes | Surname search |
| person-by-dob.json | 5 | ✅ Yes (rest-external-user) | ✅ Yes | DOB search |

**All configs make external calls and use the proxy for recording/playback!**
