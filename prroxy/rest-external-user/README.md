# REST External User Service

Mock external service that stores person data. Used to test proxy recording/playback functionality.

## Purpose

Simulates an external microservice that both rest-v1 and rest-v2 call through the proxy.

## API

### GET /health
Health check endpoint

**Response:**
```json
{
  "status": "healthy",
  "service": "rest-external-user",
  "port": "3006"
}
```

### GET /person?surname=X&dob=YYYY-MM-DD
Lookup person by surname and date of birth

**Parameters:**
- `surname` (required) - Person's surname
- `dob` (required) - Date of birth in format YYYY-MM-DD

**Success Response (200):**
```json
{
  "firstname": "Emma",
  "surname": "Thompson",
  "dob": "1985-03-15",
  "country": "United Kingdom"
}
```

**Error Responses:**
- 400 - Missing required parameters
- 404 - Person not found

## Data

Contains 25 people with various surnames and dates of birth for testing.

## Running

```bash
# Default port 3006
./start.sh

# Custom port
PORT=8000 ./start.sh
```

## Architecture

```
┌─────────────┐      ┌──────────┐      ┌────────────────────┐
│  rest-v1    │─────>│  Proxy   │─────>│ rest-external-user │
│  (client)   │      │ (8099)   │      │    (port 3006)     │
└─────────────┘      └──────────┘      └────────────────────┘
       │                   │
┌─────────────┐            │
│  rest-v2    │────────────┘
│  (client)   │
└─────────────┘
```

Both rest-v1 and rest-v2 act as thin clients calling this external service through the proxy.
