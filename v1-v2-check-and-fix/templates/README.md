# Templates for REST API Services

This directory contains template files for setting up different types of REST API services.

## CRM API Start Script (Groovy/Spring)

If your REST v1 service is a Groovy/Spring Boot application (like the CRM API), it needs a start script that runs through Gradle.

### Setup

```bash
# Copy the template to your CRM API directory
cp templates/crm-api-start.sh.template ~/work/git-other/crm-api/start.sh

# Make it executable
chmod +x ~/work/git-other/crm-api/start.sh

# Test it
cd ~/work/git-other/crm-api
PORT=8002 ./start.sh
```

### What It Does

The script:
1. Reads the `PORT` environment variable (defaults to 8002)
2. Exports it as `SERVER_PORT` for Spring Boot
3. Runs the application with `./gradlew run --console=plain`

### Customization

If your Spring Boot app uses a different port variable, edit the script:

```bash
# For SERVER_PORT
export SERVER_PORT=$PORT

# For server.port
export server_port=$PORT

# Or pass as JVM argument
./gradlew run --console=plain -Dserver.port=$PORT
```

## Node.js/TypeScript Services

Node.js services typically have their own `start.sh` that runs:
```bash
npm start
# or
node dist/server.js
```

No template needed - these usually work out of the box.

## Go Services

Go services are built and run directly:
```bash
# Build
go build -o service ./cmd/server

# Run
PORT=8080 ./service
```

The main start.sh script handles these automatically.
