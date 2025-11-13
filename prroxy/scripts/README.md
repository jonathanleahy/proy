# Scripts Directory

Utility scripts for the Prroxy project.

## Available Scripts

### `record-test-data.sh`

Systematically records test data from JSONPlaceholder API for use with REST v1 endpoints.

#### What It Records

- **10 Users** (IDs 1-10)
- **Posts by User** (filtered by userId for users 1-10)
- **Todos by User** (filtered by userId for users 1-10)
- **100 Individual Posts** (IDs 1-100)
- **200 Individual Todos** (IDs 1-200)

**Total**: ~330 HTTP interactions recorded

#### Prerequisites

1. Proxy must be running:
   ```bash
   cd proxy && make run
   # OR
   ./build/proxy
   ```

2. Proxy must be accessible at `http://0.0.0.0:8080` (default)

#### Usage

```bash
# From project root
./scripts/record-test-data.sh

# Or with custom proxy URL
PROXY_URL=http://0.0.0.0:9000 ./scripts/record-test-data.sh
```

#### What It Does

1. ✓ Checks if proxy is running
2. ✓ Sets proxy to record mode
3. ✓ Records all users (1-10)
4. ✓ Records posts for each user
5. ✓ Records todos for each user
6. ✓ Records individual posts (1-100)
7. ✓ Records individual todos (1-200)
8. ✓ Shows summary and next steps

#### Output

```
================================================
  JSONPlaceholder Test Data Recording Script
================================================

Checking if proxy is running...
✓ Proxy is running at http://0.0.0.0:8080

Setting proxy to record mode...
✓ Proxy set to record mode

Recording Users (1-10)...
  Recording: User 1... ✓
  Recording: User 2... ✓
  ...

================================================
  Recording Summary
================================================

Total requests:      330
Successful:          330
Failed:              0

Recordings saved to:
  /home/user/prroxy/recordings/

✓ All recordings completed successfully!
```

#### Environment Variables

- `PROXY_URL` - Proxy server URL (default: `http://0.0.0.0:8080`)

#### After Recording

Once recordings are complete:

1. **View in dashboard**:
   ```bash
   open http://0.0.0.0:8080/admin/ui
   ```

2. **Switch to playback mode**:
   ```bash
   curl -X POST http://0.0.0.0:8080/admin/mode \
     -H "Content-Type: application/json" \
     -d '{"mode":"playback"}'
   ```

3. **Test REST v1 with recorded data**:
   ```bash
   cd rest-v1
   npm start
   curl http://0.0.0.0:3000/api/user/1
   ```

#### Why These Recordings?

The script records data needed by all three REST v1 endpoints:

| Endpoint | Needs | Recordings |
|----------|-------|------------|
| `GET /api/user/:id` | User data | Users 1-10 |
| `GET /api/user/:id/summary` | User + Posts | Users 1-10, Posts by userId |
| `POST /api/user/:id/report` | User + Posts + Todos | Users 1-10, Posts by userId, Todos by userId |

Additional individual posts/todos provide comprehensive coverage for edge cases.

#### Troubleshooting

**Proxy not running**:
```bash
cd proxy
make build
make run
```

**Connection refused**:
- Check proxy is on port 8080: `curl http://0.0.0.0:8080/health`
- Or set custom port: `PROXY_URL=http://0.0.0.0:9000 ./scripts/record-test-data.sh`

**Recordings not saved**:
- Check `recordings/` directory exists and is writable
- Check proxy logs for errors

## Future Scripts

- `switch-proxy-mode.sh` - Toggle between record/playback modes
- `clear-recordings.sh` - Delete all recordings
- `validate-recordings.sh` - Verify recording completeness
