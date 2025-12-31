# REST API Quick Reference

Quick curl commands for testing the REST API.

## Base URL
```
http://localhost:8080
```

## Health Check
```bash
curl http://localhost:8080/health
```

## Tenant Endpoints

### Create Tenant
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "My Tenant"}'
```

### Get Tenant
```bash
curl http://localhost:8080/api/v1/tenants/{TENANT_ID}
```

## Story Endpoints

### Create Story
```bash
curl -X POST http://localhost:8080/api/v1/stories \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "{TENANT_ID}", "title": "My Story"}'
```

### Get Story
```bash
curl http://localhost:8080/api/v1/stories/{STORY_ID}
```

### List Stories
```bash
curl "http://localhost:8080/api/v1/stories?tenant_id={TENANT_ID}&limit=10&offset=0"
```

### Clone Story
```bash
curl -X POST http://localhost:8080/api/v1/stories/{STORY_ID}/clone \
  -H "Content-Type: application/json"
```

## Quick Test Script

```bash
#!/bin/bash
BASE_URL="http://localhost:8080"

# Create tenant
TENANT=$(curl -s -X POST $BASE_URL/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Tenant"}')
TENANT_ID=$(echo $TENANT | jq -r '.tenant.id')
echo "Tenant ID: $TENANT_ID"

# Create story
STORY=$(curl -s -X POST $BASE_URL/api/v1/stories \
  -H "Content-Type: application/json" \
  -d "{\"tenant_id\": \"$TENANT_ID\", \"title\": \"Test Story\"}")
STORY_ID=$(echo $STORY | jq -r '.story.id')
echo "Story ID: $STORY_ID"

# Get story
curl -s $BASE_URL/api/v1/stories/$STORY_ID | jq

# List stories
curl -s "$BASE_URL/api/v1/stories?tenant_id=$TENANT_ID" | jq

# Clone story
curl -s -X POST $BASE_URL/api/v1/stories/$STORY_ID/clone \
  -H "Content-Type: application/json" | jq
```

