# Wago Project API Collection

Base URL: `http://localhost:8080/api/v1`

## Authentication

### Generate PIN
```bash
curl -X POST http://localhost:8080/api/v1/auth/generate-pin \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "6281234567890"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "6281234567890",
    "pin": "123456"
  }'
```

> **Note:** For subsequent requests, include the `Authorization` header with the token received from login:
> `Authorization: Bearer <YOUR_TOKEN>`

### Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

## Sessions

### Create Session
```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_name": "My Bot Session",
    "webhook_url": "https://webhook.site/..."
  }'
```

### Get All Sessions
```bash
curl -X GET http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Start Session
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/start \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Stop Session
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/stop \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Update Session
```bash
curl -X PUT http://localhost:8080/api/v1/sessions/{session_id} \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_name": "Updated Session Name",
    "webhook_url": "https://new-webhook.url",
    "is_group_response_enabled": true
  }'
```

### Delete Session
```bash
curl -X DELETE http://localhost:8080/api/v1/sessions/{session_id} \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Get Session Analytics
```bash
curl -X GET http://localhost:8080/api/v1/sessions/{session_id}/analytics \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Get Session Contacts
```bash
curl -X GET http://localhost:8080/api/v1/sessions/{session_id}/contacts \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Send Message (Direct)
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/send-message \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "628123456789",
    "message": "Hello from Wago API!"
  }'
```

#### Send Message with PIN (alternative)
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/send-message \
  -H "Authorization: Pin <YOUR_PIN>" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "628123456789",
    "message": "Hello from Wago API!"
  }'
```
> You can also use header `X-Pin: <YOUR_PIN>` if you prefer keeping `Authorization` for other auth schemes.
