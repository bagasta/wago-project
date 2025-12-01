# PRD.md - WhatsApp API Multi-Session with n8n Webhook Integration

## 1. Project Overview

### 1.1 Project Name
**WhatsApp Multi-Session API Gateway**

### 1.2 Description
Aplikasi berbasis Go (Golang) yang menyediakan API untuk mengelola multiple WhatsApp Web sessions dengan integrasi webhook n8n untuk AI backend processing. Sistem mendukung multi-user dengan PIN-based authentication dan real-time QR code updates.

### 1.3 Tech Stack
- **Backend**: Go (Golang) 1.21+
- **Frontend**: React 18 + Vite + Tailwind CSS
- **Database**: PostgreSQL 15+
- **WhatsApp Library**: whatsmeow (github.com/aldinokemal/go-whatsapp-web-multidevice reference)
- **WebSocket**: gorilla/websocket untuk real-time updates
- **Authentication**: PIN-based Basic Auth
- **UI Components**: Headless UI + Lucide Icons
- **Charts**: Recharts
- **Notifications**: Sonner (toast)

---

## 2. Core Features

### 2.1 User Management
- ✅ PIN Generation & Registration (6 karakter alfanumerik unik)
- ✅ PIN-based Login (Basic Auth)
- ✅ Multi-user support
- ✅ Unique PIN enforcement di database level
- ✅ User session management

### 2.2 WhatsApp Session Management
- ✅ Multiple sessions per user (no limit)
- ✅ QR Code generation untuk pairing
- ✅ Real-time QR code updates via WebSocket
- ✅ Session start/stop/reconnect
- ✅ Auto-reconnect on disconnect (configurable)
- ✅ Session status monitoring (connected/disconnected/qr)
- ✅ Webhook URL configuration per session
- ✅ Session device info storage

### 2.3 Message Handling
- ✅ Forward incoming messages ke n8n webhook
- ✅ Smart group message filtering (HANYA jika bot di-mention dengan @)
- ✅ Private message forwarding (semua pesan private di-forward)
- ✅ Message sending capability via API
- ✅ Media support (image, document, audio, video)
- ✅ Message delivery status tracking
- ✅ Quoted message support

### 2.4 Analytics & Monitoring
- ✅ Total messages processed (per session & global)
- ✅ Messages handled by AI (webhook success count)
- ✅ Average webhook response time (milliseconds)
- ✅ Private vs Group message ratio
- ✅ Session uptime tracking (percentage)
- ✅ Error rate monitoring
- ✅ Peak usage times (hourly breakdown)
- ✅ Daily/Weekly/Monthly aggregation

### 2.5 Webhook Integration
- ✅ Configurable webhook URL per session
- ✅ Automatic retry mechanism (3x dengan exponential backoff)
- ✅ Webhook response time tracking
- ✅ Webhook failure logging
- ✅ Timeout handling (30 seconds default)
- ✅ Payload customization (JSON format)

---

## 3. System Architecture

### 3.1 Folder Structure
```
whatsapp-api/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/           # Config loader & validation
│   │   │   └── config.go
│   │   ├── database/         # Database connection & migrations
│   │   │   ├── postgres.go
│   │   │   └── migrations.go
│   │   ├── handler/          # HTTP handlers
│   │   │   ├── auth.go
│   │   │   ├── session.go
│   │   │   ├── message.go
│   │   │   └── analytics.go
│   │   ├── middleware/       # Middlewares
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── ratelimit.go
│   │   ├── model/            # Data models
│   │   │   ├── user.go
│   │   │   ├── session.go
│   │   │   ├── analytics.go
│   │   │   └── message.go
│   │   ├── repository/       # Database queries
│   │   │   ├── user_repo.go
│   │   │   ├── session_repo.go
│   │   │   └── analytics_repo.go
│   │   ├── service/          # Business logic
│   │   │   ├── auth_service.go
│   │   │   ├── session_service.go
│   │   │   ├── message_service.go
│   │   │   ├── webhook_service.go
│   │   │   └── analytics_service.go
│   │   ├── whatsapp/         # WhatsApp integration
│   │   │   ├── client.go
│   │   │   ├── handler.go
│   │   │   ├── qr.go
│   │   │   └── message.go
│   │   ├── websocket/        # WebSocket manager
│   │   │   ├── hub.go
│   │   │   └── client.go
│   │   └── utils/            # Helper functions
│   │       ├── pin_generator.go
│   │       ├── response.go
│   │       └── validator.go
│   ├── migrations/           # SQL migrations
│   │   ├── 001_create_users.up.sql
│   │   ├── 001_create_users.down.sql
│   │   ├── 002_create_sessions.up.sql
│   │   ├── 002_create_sessions.down.sql
│   │   ├── 003_create_analytics.up.sql
│   │   └── 003_create_analytics.down.sql
│   ├── docs/
│   │   ├── API.md            # Complete API documentation
│   │   ├── CURL_COLLECTION.md # cURL examples
│   │   └── ARCHITECTURE.md   # System architecture
│   ├── .env.example
│   ├── .gitignore
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── Makefile
│   └── README.md
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/           # Reusable UI components
│   │   │   │   ├── Button.jsx
│   │   │   │   ├── Input.jsx
│   │   │   │   ├── Modal.jsx
│   │   │   │   ├── Badge.jsx
│   │   │   │   └── Card.jsx
│   │   │   ├── SessionCard.jsx
│   │   │   ├── QRCodeModal.jsx
│   │   │   ├── SessionForm.jsx
│   │   │   ├── AnalyticsChart.jsx
│   │   │   ├── StatsCard.jsx
│   │   │   └── Header.jsx
│   │   ├── pages/
│   │   │   ├── GeneratePin.jsx
│   │   │   ├── Login.jsx
│   │   │   ├── Dashboard.jsx
│   │   │   └── Analytics.jsx
│   │   ├── services/
│   │   │   ├── api.js        # Axios instance
│   │   │   ├── auth.js       # Auth API calls
│   │   │   ├── session.js    # Session API calls
│   │   │   ├── analytics.js  # Analytics API calls
│   │   │   └── websocket.js  # WebSocket client
│   │   ├── hooks/
│   │   │   ├── useAuth.js
│   │   │   ├── useSessions.js
│   │   │   ├── useWebSocket.js
│   │   │   └── useAnalytics.js
│   │   ├── utils/
│   │   │   ├── constants.js
│   │   │   ├── helpers.js
│   │   │   └── storage.js
│   │   ├── context/
│   │   │   └── AuthContext.jsx
│   │   ├── App.jsx
│   │   ├── main.jsx
│   │   └── index.css
│   ├── public/
│   │   └── favicon.ico
│   ├── .env.example
│   ├── .gitignore
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── Dockerfile
│   └── README.md
│
├── docker-compose.yml
├── .gitignore
├── HOW_TO_USE.md
└── README.md
```

### 3.2 Database Schema

#### Table: users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pin VARCHAR(6) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    CONSTRAINT pin_format CHECK (pin ~ '^[A-Z0-9]{6}$')
);

CREATE INDEX idx_users_pin ON users(pin);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

#### Table: sessions
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_name VARCHAR(100) NOT NULL,
    webhook_url TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'disconnected',
    qr_code TEXT,
    phone_number VARCHAR(20),
    device_info JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_connected TIMESTAMP,
    uptime_seconds BIGINT DEFAULT 0,
    CONSTRAINT unique_session_name UNIQUE(user_id, session_name),
    CONSTRAINT valid_status CHECK (status IN ('qr', 'connected', 'disconnected'))
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_created_at ON sessions(created_at DESC);
```

#### Table: analytics
```sql
CREATE TABLE analytics (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    message_id VARCHAR(100),
    from_number VARCHAR(50),
    message_type VARCHAR(20),
    is_group BOOLEAN DEFAULT false,
    is_mention BOOLEAN DEFAULT false,
    webhook_sent BOOLEAN DEFAULT false,
    webhook_success BOOLEAN DEFAULT false,
    webhook_response_time_ms INTEGER,
    webhook_status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_message_type CHECK (message_type IN ('text', 'image', 'document', 'audio', 'video', 'sticker', 'location', 'contact'))
);

CREATE INDEX idx_analytics_session_id ON analytics(session_id);
CREATE INDEX idx_analytics_created_at ON analytics(created_at DESC);
CREATE INDEX idx_analytics_session_created ON analytics(session_id, created_at DESC);
CREATE INDEX idx_analytics_webhook_sent ON analytics(webhook_sent, created_at);
CREATE INDEX idx_analytics_is_group ON analytics(is_group, created_at);
```

#### Table: messages_log (Optional - untuk debugging)
```sql
CREATE TABLE messages_log (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    direction VARCHAR(10) NOT NULL,
    from_number VARCHAR(50),
    to_number VARCHAR(50),
    message_type VARCHAR(20),
    content TEXT,
    media_url TEXT,
    group_id VARCHAR(100),
    group_name VARCHAR(200),
    is_group BOOLEAN DEFAULT false,
    quoted_message_id VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_direction CHECK (direction IN ('incoming', 'outgoing'))
);

CREATE INDEX idx_messages_session_timestamp ON messages_log(session_id, timestamp DESC);
CREATE INDEX idx_messages_from_number ON messages_log(from_number);
CREATE INDEX idx_messages_group_id ON messages_log(group_id) WHERE is_group = true;
```

---

## 4. API Endpoints

### 4.1 Authentication

#### POST /api/v1/auth/generate-pin
Generate new PIN untuk user baru

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/generate-pin
```

**Response:**
```json
{
    "success": true,
    "data": {
        "user_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "pin": "A1B2C3"
    },
    "message": "PIN generated successfully. Please save this PIN for login."
}
```

---

#### POST /api/v1/auth/login
Login dengan PIN (Basic Auth)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -u "A1B2C3:"
```

**Response:**
```json
{
    "success": true,
    "data": {
        "user_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "pin": "A1B2C3"
    },
    "message": "Login successful"
}
```

---

#### POST /api/v1/auth/logout
Logout user

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "message": "Logout successful"
}
```

---

### 4.2 Session Management

#### POST /api/v1/sessions
Create new WhatsApp session

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_name": "Agent 1",
    "webhook_url": "https://n8n.example.com/webhook/whatsapp"
  }'
```

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "session_name": "Agent 1",
        "webhook_url": "https://n8n.example.com/webhook/whatsapp",
        "status": "disconnected",
        "qr_code": null,
        "phone_number": null,
        "created_at": "2024-12-01T10:00:00Z"
    },
    "message": "Session created successfully. Please start the session to get QR code."
}
```

---

#### GET /api/v1/sessions
List all sessions untuk user yang login

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "data": [
        {
            "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
            "session_name": "Agent 1",
            "webhook_url": "https://n8n.example.com/webhook/whatsapp",
            "status": "connected",
            "phone_number": "6281234567890",
            "last_connected": "2024-12-01T10:00:00Z",
            "uptime_percentage": 98.5,
            "created_at": "2024-11-20T08:00:00Z"
        },
        {
            "session_id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
            "session_name": "Agent 2",
            "webhook_url": "https://n8n.example.com/webhook/whatsapp2",
            "status": "qr",
            "phone_number": null,
            "last_connected": null,
            "uptime_percentage": 0,
            "created_at": "2024-12-01T09:30:00Z"
        }
    ],
    "message": "Sessions retrieved successfully"
}
```

---

#### GET /api/v1/sessions/:id
Get session detail

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "session_name": "Agent 1",
        "webhook_url": "https://n8n.example.com/webhook/whatsapp",
        "status": "connected",
        "qr_code": null,
        "phone_number": "6281234567890",
        "device_info": {
            "platform": "android",
            "device_manufacturer": "Samsung",
            "device_model": "Galaxy S21"
        },
        "created_at": "2024-11-20T08:00:00Z",
        "updated_at": "2024-12-01T10:00:00Z",
        "last_connected": "2024-12-01T10:00:00Z"
    }
}
```

---

#### PUT /api/v1/sessions/:id
Update session (nama atau webhook URL)

**Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_name": "Agent 1 Updated",
    "webhook_url": "https://new-webhook.example.com"
  }'
```

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "session_name": "Agent 1 Updated",
        "webhook_url": "https://new-webhook.example.com",
        "status": "connected",
        "updated_at": "2024-12-01T10:15:00Z"
    },
    "message": "Session updated successfully"
}
```

---

#### POST /api/v1/sessions/:id/start
Start/Connect session (akan generate QR jika belum connected)

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890/start \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "status": "qr",
        "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
        "message": "Please scan the QR code to connect WhatsApp"
    }
}
```

---

#### POST /api/v1/sessions/:id/reconnect
Force reconnect session

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890/reconnect \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "status": "connected"
    },
    "message": "Session reconnected successfully"
}
```

---

#### DELETE /api/v1/sessions/:id
Delete session dan logout dari WhatsApp

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "message": "Session deleted successfully"
}
```

---

#### GET /api/v1/sessions/:id/qr
Get current QR code (jika status = qr)

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890/qr \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "data": {
        "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
        "status": "qr",
        "expires_in": 45
    }
}
```

---

### 4.3 Messaging

#### POST /api/v1/sessions/:id/send
Send message via WhatsApp

**Request (Text):**
```bash
curl -X POST http://localhost:8080/api/v1/sessions/a1b2c3d4/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "6281234567890",
    "type": "text",
    "content": "Hello from API"
  }'
```

**Request (Image):**
```bash
curl -X POST http://localhost:8080/api/v1/sessions/a1b2c3d4/send \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "6281234567890",
    "type": "image",
    "content": "Check this out!",
    "media_url": "https://example.com/image.jpg"
  }'
```

**Response:**
```json
{
    "success": true,
    "data": {
        "message_id": "3EB0C767D0B1234567890ABCDEF",
        "status": "sent",
        "timestamp": "2024-12-01T10:30:00Z"
    },
    "message": "Message sent successfully"
}
```

---

### 4.4 Analytics

#### GET /api/v1/analytics/sessions/:id
Get analytics untuk specific session

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/analytics/sessions/a1b2c3d4?period=daily&start_date=2024-11-01&end_date=2024-12-01" \
  -H "Authorization: Bearer <token>"
```

**Query Parameters:**
- `period`: daily | weekly | monthly (default: daily)
- `start_date`: YYYY-MM-DD (optional)
- `end_date`: YYYY-MM-DD (optional)

**Response:**
```json
{
    "success": true,
    "data": {
        "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "session_name": "Agent 1",
        "period": "daily",
        "start_date": "2024-11-01",
        "end_date": "2024-12-01",
        "summary": {
            "total_messages": 1250,
            "messages_handled_by_ai": 1180,
            "success_rate": 94.4,
            "average_response_time_ms": 450,
            "private_messages": 800,
            "group_messages": 450,
            "group_mentions": 380,
            "uptime_percentage": 99.2,
            "total_errors": 70
        },
        "peak_hours": [
            {"hour": 9, "count": 150},
            {"hour": 10, "count": 180},
            {"hour": 14, "count": 200},
            {"hour": 15, "count": 170}
        ],
        "daily_breakdown": [
            {
                "date": "2024-11-01",
                "total_messages": 45,
                "ai_handled": 42,
                "avg_response_time_ms": 420,
                "errors": 3
            },
            {
                "date": "2024-11-02",
                "total_messages": 52,
                "ai_handled": 50,
                "avg_response_time_ms": 380,
                "errors": 2
            }
        ],
        "message_types": {
            "text": 1100,
            "image": 80,
            "document": 40,
            "audio": 20,
            "video": 10
        }
    }
}
```

---

#### GET /api/v1/analytics/overview
Get overview analytics untuk semua sessions user

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/analytics/overview \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
    "success": true,
    "data": {
        "total_sessions": 5,
        "active_sessions": 4,
        "total_messages_today": 350,
        "total_messages_this_week": 2100,
        "total_messages_this_month": 8500,
        "avg_response_time_ms": 420,
        "overall_success_rate": 95.2,
        "top_performing_sessions": [
            {
                "session_id": "a1b2c3d4",
                "session_name": "Agent 1",
                "total_messages": 1250,
                "success_rate": 96.5,
                "avg_response_time_ms": 380
            },
            {
                "session_id": "b2c3d4e5",
                "session_name": "Agent 2",
                "total_messages": 980,
                "success_rate": 94.8,
                "avg_response_time_ms": 410
            }
        ]
    }
}
```

---

### 4.5 WebSocket

#### WS /ws/sessions/:id
Real-time updates untuk QR code dan status session

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/sessions/a1b2c3d4?token=<bearer_token>');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log(data);
};
```

**Message Types:**

**QR Code Update:**
```json
{
    "type": "qr_update",
    "data": {
        "qr_code": "data:image/png;base64,iVBORw0KGgo...",
        "expires_in": 60
    },
    "timestamp": "2024-12-01T10:00:00Z"
}
```

**Status Update:**
```json
{
    "type": "status_update",
    "data": {
        "status": "connected",
        "phone_number": "6281234567890",
        "device_info": {
            "platform": "android",
            "device_manufacturer": "Samsung"
        }
    },
    "timestamp": "2024-12-01T10:01:30Z"
}
```

**Message Received (Real-time notification):**
```json
{
    "type": "message_received",
    "data": {
        "message_id": "3EB0C767D0B1234567890",
        "from": "6281234567890",
        "from_name": "John Doe",
        "type": "text",
        "content": "Hello!",
        "is_mention": true,
        "mentioned_numbers": ["6289876543210"],
        "quoted_message": null
    }
}
```

---

### 5.3 Group Message WITHOUT Mention (DIABAIKAN - tidak dikirim ke webhook)
```json
// Pesan ini TIDAK akan dikirim ke webhook n8n
// Karena bot tidak di-mention dalam grup
{
    "from": "6281234567890",
    "content": "Hello everyone!",
    "is_group": true,
    "is_mention": false
}
// ❌ Webhook n8n tidak akan menerima payload ini
```

---

### 5.4 Media Message (Image)
```json
{
    "event": "message.received",
    "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "session_name": "Agent 1",
    "timestamp": "2024-12-01T10:10:00Z",
    "message": {
        "id": "3EB0C767D0B1234567890MEDIA",
        "from": "6281234567890",
        "from_name": "John Doe",
        "to": "6289876543210",
        "type": "image",
        "content": "Check this screenshot",
        "media_url": "https://storage.example.com/media/image_123.jpg",
        "is_group": false,
        "group_id": null,
        "group_name": null,
        "is_mention": false,
        "quoted_message": null
    }
}
```

---

### 5.5 Quoted/Reply Message
```json
{
    "event": "message.received",
    "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "session_name": "Agent 1",
    "timestamp": "2024-12-01T10:15:00Z",
    "message": {
        "id": "3EB0C767D0B1234567890REPLY",
        "from": "6281234567890",
        "from_name": "John Doe",
        "to": "6289876543210",
        "type": "text",
        "content": "Yes, exactly!",
        "media_url": null,
        "is_group": false,
        "group_id": null,
        "group_name": null,
        "is_mention": false,
        "quoted_message": {
            "id": "3EB0C767D0B1234567890ORIG",
            "from": "6289876543210",
            "content": "Do you mean the blue one?",
            "type": "text"
        }
    }
}
```

---

### 5.6 Webhook Retry Logic
- **Timeout**: 30 seconds per request
- **Retry Count**: 3 attempts
- **Backoff Strategy**: Exponential (1s, 2s, 4s)
- **Failure Handling**: Log ke analytics table dengan error_message
- **Success Criteria**: HTTP 200-299 status code

**Retry Flow:**
```
Attempt 1 → Failed (timeout/error) → Wait 1s
Attempt 2 → Failed (timeout/error) → Wait 2s
Attempt 3 → Failed (timeout/error) → Wait 4s
Final Failure → Log to analytics (webhook_success = false)
```

---

## 6. Frontend Implementation

### 6.1 Tech Stack Details

#### Core Libraries
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "axios": "^1.6.0",
    "recharts": "^2.10.0",
    "react-qr-code": "^2.0.12",
    "sonner": "^1.2.0",
    "lucide-react": "^0.294.0",
    "clsx": "^2.0.0",
    "date-fns": "^2.30.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.2.0",
    "vite": "^5.0.0",
    "tailwindcss": "^3.3.0",
    "postcss": "^8.4.32",
    "autoprefixer": "^10.4.16",
    "eslint": "^8.55.0",
    "eslint-plugin-react": "^7.33.2"
  }
}
```

---

### 6.2 Pages

#### 6.2.1 Generate PIN Page (`/`)
**Features:**
- Simple landing page dengan tombol "Generate PIN"
- Display generated PIN dengan copy button
- Warning message: "Save this PIN! You'll need it to login."
- Link ke Login page setelah generate

**UI Flow:**
1. User klik "Generate PIN"
2. API call ke `/api/v1/auth/generate-pin`
3. Display PIN dalam card besar dengan copy button
4. Show warning toast
5. Button "Go to Login"

---

#### 6.2.2 Login Page (`/login`)
**Features:**
- Input field untuk PIN (6 karakter)
- Login button
- Link ke "Generate PIN" page
- Error handling untuk wrong PIN

**UI Flow:**
1. User input PIN
2. Click "Login"
3. API call dengan Basic Auth
4. Save token di localStorage
5. Redirect ke Dashboard

---

#### 6.2.3 Dashboard Page (`/dashboard`)
**Features:**
- Header dengan user info + Logout button
- Button "+ New Session" (modal)
- Table/Grid sessions dengan:
  - Session Name
  - Webhook URL (truncated)
  - Status badge (green=connected, yellow=qr, gray=disconnected)
  - Phone Number (jika connected)
  - Action buttons: Start, Edit, Reconnect, View Analytics, Delete
- Real-time status updates via polling/websocket
- Empty state jika belum ada session

**UI Flow:**
1. Load semua sessions dari API
2. Poll status setiap 30 detik (atau WebSocket)
3. Click "+ New Session" → Open modal
4. Click "Start" → Open QR Modal (jika status disconnected/qr)
5. Click "Edit" → Open edit modal
6. Click "Delete" → Confirmation → Delete session

---

#### 6.2.4 QR Code Modal
**Features:**
- Display QR code (real-time update via WebSocket)
- Status indicator: "Waiting for scan..." / "Connected!"
- Timer countdown (60 seconds)
- Auto-close saat status = "connected"
- Close button

**UI Flow:**
1. Open modal saat click "Start Session"
2. Establish WebSocket connection ke `/ws/sessions/:id`
3. Listen untuk `qr_update` messages
4. Update QR code real-time
5. Listen untuk `status_update` → jika "connected", show success + auto close
6. Jika QR expire, show "Refresh QR" button

---

#### 6.2.5 Analytics Page (`/analytics/:sessionId`)
**Features:**
- Session selector dropdown (jika mau lihat session lain)
- Date range picker (Today, Last 7 days, Last 30 days, Custom)
- Stats Cards:
  - Total Messages
  - AI Handled Messages (+ success rate %)
  - Avg Response Time (ms)
  - Uptime %
- Charts:
  - Line Chart: Messages over time
  - Pie Chart: Private vs Group messages
  - Bar Chart: Peak hours
  - Bar Chart: Message types (text, image, document, etc)
- Table: Daily breakdown

**UI Flow:**
1. Load analytics data dari API
2. User select date range → Reload data
3. Interactive charts dengan tooltips
4. Export button (future: CSV/PDF)

---

### 6.3 Component Structure

#### Button.jsx (Reusable)
```jsx
export default function Button({ 
  children, 
  variant = 'primary', 
  size = 'md', 
  onClick, 
  disabled,
  className 
}) {
  const baseStyles = 'rounded font-medium transition focus:outline-none focus:ring-2';
  const variants = {
    primary: 'bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500',
    danger: 'bg-red-600 hover:bg-red-700 text-white focus:ring-red-500',
    secondary: 'bg-gray-200 hover:bg-gray-300 text-gray-800 focus:ring-gray-400',
    ghost: 'hover:bg-gray-100 text-gray-700'
  };
  const sizes = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2',
    lg: 'px-6 py-3 text-lg'
  };
  
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`${baseStyles} ${variants[variant]} ${sizes[size]} ${className} ${
        disabled ? 'opacity-50 cursor-not-allowed' : ''
      }`}
    >
      {children}
    </button>
  );
}
```

---

#### SessionCard.jsx
```jsx
import { Phone, Wifi, WifiOff, QrCode } from 'lucide-react';
import Badge from './ui/Badge';
import Button from './ui/Button';

export default function SessionCard({ session, onStart, onEdit, onDelete, onAnalytics }) {
  const statusConfig = {
    connected: { color: 'green', icon: Wifi, text: 'Connected' },
    qr: { color: 'yellow', icon: QrCode, text: 'Waiting QR' },
    disconnected: { color: 'gray', icon: WifiOff, text: 'Disconnected' }
  };
  
  const { color, icon: Icon, text } = statusConfig[session.status] || statusConfig.disconnected;
  
  return (
    <div className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="font-bold text-lg">{session.session_name}</h3>
          <p className="text-sm text-gray-500 truncate max-w-xs">
            {session.webhook_url}
          </p>
          {session.phone_number && (
            <div className="flex items-center gap-1 mt-1 text-sm text-gray-600">
              <Phone size={14} />
              <span>{session.phone_number}</span>
            </div>
          )}
        </div>
        <Badge color={color}>
          <Icon size={14} className="mr-1" />
          {text}
        </Badge>
      </div>
      
      <div className="flex gap-2 flex-wrap">
        {session.status !== 'connected' && (
          <Button size="sm" onClick={() => onStart(session)}>
            Start
          </Button>
        )}
        <Button size="sm" variant="secondary" onClick={() => onEdit(session)}>
          Edit
        </Button>
        {session.status === 'connected' && (
          <Button size="sm" variant="secondary">
            Reconnect
          </Button>
        )}
        <Button size="sm" variant="secondary" onClick={() => onAnalytics(session.session_id)}>
          Analytics
        </Button>
        <Button size="sm" variant="danger" onClick={() => onDelete(session)}>
          Delete
        </Button>
      </div>
    </div>
  );
}
```

---

#### QRCodeModal.jsx
```jsx
import { useEffect, useState } from 'react';
import { QRCodeSVG } from 'react-qr-code';
import { X, Loader2, CheckCircle } from 'lucide-react';
import { connectWebSocket } from '../services/websocket';

export default function QRCodeModal({ session, onClose }) {
  const [qrCode, setQrCode] = useState(null);
  const [status, setStatus] = useState('loading');
  const [countdown, setCountdown] = useState(60);
  
  useEffect(() => {
    // Connect to WebSocket
    const ws = connectWebSocket(session.session_id);
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'qr_update') {
        setQrCode(data.data.qr_code);
        setStatus('waiting');
        setCountdown(data.data.expires_in || 60);
      }
      
      if (data.type === 'status_update' && data.data.status === 'connected') {
        setStatus('connected');
        setTimeout(() => onClose(), 2000); // Auto close after 2s
      }
    };
    
    return () => ws.close();
  }, [session.session_id]);
  
  useEffect(() => {
    if (status === 'waiting' && countdown > 0) {
      const timer = setInterval(() => setCountdown(c => c - 1), 1000);
      return () => clearInterval(timer);
    }
  }, [status, countdown]);
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold">Connect WhatsApp</h2>
          <button onClick={onClose} className="hover:bg-gray-100 p-1 rounded">
            <X size={20} />
          </button>
        </div>
        
        <div className="flex flex-col items-center">
          {status === 'loading' && (
            <div className="py-12">
              <Loader2 className="animate-spin" size={48} />
              <p className="mt-4 text-gray-600">Generating QR Code...</p>
            </div>
          )}
          
          {status === 'waiting' && qrCode && (
            <>
              <div className="p-4 bg-white border-2 border-gray-200 rounded-lg">
                <QRCodeSVG value={qrCode} size={256} />
              </div>
              <p className="mt-4 text-sm text-gray-600 text-center">
                Scan this QR code with WhatsApp
              </p>
              <p className="text-xs text-gray-500 mt-2">
                Expires in {countdown}s
              </p>
            </>
          )}
          
          {status === 'connected' && (
            <div className="py-12 text-center">
              <CheckCircle className="text-green-500 mx-auto" size={64} />
              <p className="mt-4 text-lg font-semibold text-green-600">
                Connected Successfully!
              </p>
            </div>
          )}
          
          {countdown === 0 && status === 'waiting' && (
            <Button className="mt-4" onClick={() => window.location.reload()}>
              Refresh QR Code
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
```

---

#### AnalyticsChart.jsx (using Recharts)
```jsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export default function MessageTrendChart({ data }) {
  return (
    <div className="bg-white p-6 rounded-lg border border-gray-200">
      <h3 className="text-lg font-semibold mb-4">Message Trend</h3>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="date" />
          <YAxis />
          <Tooltip />
          <Legend />
          <Line 
            type="monotone" 
            dataKey="total_messages" 
            stroke="#3b82f6" 
            name="Total Messages"
            strokeWidth={2}
          />
          <Line 
            type="monotone" 
            dataKey="ai_handled" 
            stroke="#10b981" 
            name="AI Handled"
            strokeWidth={2}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
```

---

### 6.4 Services Layer

#### api.js (Axios Instance)
```javascript
import axios from 'axios';
import { toast } from 'sonner';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
});

// Request interceptor - add token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
      toast.error('Session expired. Please login again.');
    } else {
      toast.error(error.response?.data?.message || 'An error occurred');
    }
    return Promise.reject(error);
  }
);

export default api;
```

---

#### session.js (Session API)
```javascript
import api from './api';

export const sessionService = {
  // Get all sessions
  getSessions: () => api.get('/sessions'),
  
  // Get session by ID
  getSession: (id) => api.get(`/sessions/${id}`),
  
  // Create new session
  createSession: (data) => api.post('/sessions', data),
  
  // Update session
  updateSession: (id, data) => api.put(`/sessions/${id}`, data),
  
  // Delete session
  deleteSession: (id) => api.delete(`/sessions/${id}`),
  
  // Start session
  startSession: (id) => api.post(`/sessions/${id}/start`),
  
  // Reconnect session
  reconnectSession: (id) => api.post(`/sessions/${id}/reconnect`),
  
  // Get QR code
  getQR: (id) => api.get(`/sessions/${id}/qr`),
  
  // Send message
  sendMessage: (id, data) => api.post(`/sessions/${id}/send`, data),
};
```

---

#### websocket.js (WebSocket Client)
```javascript
export function connectWebSocket(sessionId) {
  const token = localStorage.getItem('token');
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
  const ws = new WebSocket(`${wsUrl}/ws/sessions/${sessionId}?token=${token}`);
  
  ws.onopen = () => {
    console.log('WebSocket connected for session:', sessionId);
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
  
  ws.onclose = () => {
    console.log('WebSocket closed for session:', sessionId);
  };
  
  return ws;
}
```

---

### 6.5 Custom Hooks

#### useAuth.js
```javascript
import { useState, useEffect } from 'react';
import { authService } from '../services/auth';

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const token = localStorage.getItem('token');
    const pin = localStorage.getItem('pin');
    
    if (token && pin) {
      setUser({ pin });
    }
    setLoading(false);
  }, []);
  
  const login = async (pin) => {
    const response = await authService.login(pin);
    localStorage.setItem('token', response.data.token);
    localStorage.setItem('pin', response.data.pin);
    setUser({ pin: response.data.pin });
    return response;
  };
  
  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('pin');
    setUser(null);
  };
  
  return { user, loading, login, logout };
}
```

---

#### useSessions.js
```javascript
import { useState, useEffect } from 'react';
import { sessionService } from '../services/session';
import { toast } from 'sonner';

export function useSessions() {
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  
  const fetchSessions = async () => {
    try {
      setLoading(true);
      const response = await sessionService.getSessions();
      setSessions(response.data);
    } catch (error) {
      toast.error('Failed to load sessions');
    } finally {
      setLoading(false);
    }
  };
  
  useEffect(() => {
    fetchSessions();
    
    // Poll every 30 seconds
    const interval = setInterval(fetchSessions, 30000);
    return () => clearInterval(interval);
  }, []);
  
  const createSession = async (data) => {
    try {
      await sessionService.createSession(data);
      toast.success('Session created successfully');
      fetchSessions();
    } catch (error) {
      toast.error('Failed to create session');
    }
  };
  
  const deleteSession = async (id) => {
    try {
      await sessionService.deleteSession(id);
      toast.success('Session deleted successfully');
      fetchSessions();
    } catch (error) {
      toast.error('Failed to delete session');
    }
  };
  
  return { sessions, loading, createSession, deleteSession, refetch: fetchSessions };
}
```

---

## 7. Non-Functional Requirements

### 7.1 Performance
- ⚡ **API Response Time**: < 200ms (excluding webhook calls)
- ⚡ **Concurrent Sessions**: Minimum 100 sessions simultaneously
- ⚡ **Memory Usage**: < 50MB per WhatsApp session
- ⚡ **Database**: Connection pooling (max 20 connections, idle 5)
- ⚡ **Frontend Bundle**: < 500KB gzipped (production build)
- ⚡ **First Contentful Paint**: < 1.5s
- ⚡ **Time to Interactive**: < 3s

### 7.2 Security
- 🔒 **PIN Format**: 6 karakter alfanumerik (A-Z, 0-9), uppercase only
- 🔒 **PIN Uniqueness**: Enforced di database level dengan UNIQUE constraint
- 🔒 **Rate Limiting**: 100 requests/minute per user/IP
- 🔒 **CORS**: Whitelist specific origins only
- 🔒 **Input Validation**: Semua input di-validate & sanitize
- 🔒 **SQL Injection**: Prepared statements only
- 🔒 **XSS Protection**: Content Security Policy headers
- 🔒 **Token Expiry**: JWT token expire setelah 24 jam

### 7.3 Reliability
- ♻️ **Auto-Reconnect**: WhatsApp sessions auto-reconnect on disconnect
- ♻️ **Graceful Shutdown**: Handle SIGTERM dengan proper cleanup
- ♻️ **Error Logging**: Structured logging dengan levels (INFO, WARN, ERROR)
- ♻️ **Database Transactions**: ACID compliance
- ♻️ **Webhook Retry**: 3 attempts dengan exponential backoff
- ♻️ **Health Check**: `/health` endpoint untuk monitoring

### 7.4 Scalability
- 📈 **Horizontal Scaling**: Stateless API design
- 📈 **Database Indexing**: Proper indexes untuk query optimization
- 📈 **Environment Config**: Semua config via environment variables
- 📈 **Docker Support**: Multi-stage builds untuk optimal image size

### 7.5 Code Quality
- ✨ **Go Standards**: Idiomatic Go code, following effective Go guidelines
- ✨ **Error Handling**: Proper error wrapping dan contextual errors
- ✨ **Comments**: Godoc-style comments untuk exported functions
- ✨ **Linting**: golangci-lint dengan standard ruleset
- ✨ **Testing**: Unit tests untuk critical business logic (minimum 60% coverage)
- ✨ **Logging**: Structured logging dengan contextual fields

---

## 8. Deployment

### 8.1 Docker Compose
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: whatsapp_db
    environment:
      POSTGRES_DB: whatsapp_api
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "${DB_PORT:-5432}:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U admin -d whatsapp_api"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: whatsapp_backend
    ports:
      - "${API_PORT:-8080}:8080"
    environment:
      DATABASE_URL: postgresql://admin:${DB_PASSWORD:-changeme}@postgres:5432/whatsapp_api?sslmode=disable
      PORT: 8080
      ENV: production
      WHATSAPP_SESSION_PATH: /app/sessions
      WHATSAPP_AUTO_RECONNECT: "true"
      WEBHOOK_TIMEOUT: 30
      WEBHOOK_MAX_RETRIES: 3
      RATE_LIMIT_PER_MINUTE: 100
      CORS_ALLOWED_ORIGINS: http://localhost:3000
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - whatsapp_sessions:/app/sessions
    restart: unless-stopped

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: whatsapp_frontend
    ports:
      - "${FRONTEND_PORT:-3000}:80"
    environment:
      VITE_API_URL: http://localhost:8080/api/v1
      VITE_WS_URL: ws://localhost:8080
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  whatsapp_sessions:

networks:
  default:
    name: whatsapp_network
```

---

### 8.2 Backend Dockerfile
```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Create sessions directory
RUN mkdir -p /app/sessions

EXPOSE 8080

CMD ["./main"]
```

---

### 8.3 Frontend Dockerfile
```dockerfile
# Stage 1: Build
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

# Build for production
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:alpine

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

---

### 8.4 Environment Variables

#### Backend (.env)
```bash
# Server
PORT=8080
ENV=production

# Database
DATABASE_URL=postgresql://admin:password@localhost:5432/whatsapp_api?sslmode=disable
DB_MAX_CONNECTIONS=20
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_MAX_LIFETIME=300s

# Security
RATE_LIMIT_PER_MINUTE=100
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourapp.com
JWT_SECRET=your-secret-key-change-this-in-production

# WhatsApp
WHATSAPP_SESSION_PATH=/app/sessions
WHATSAPP_AUTO_RECONNECT=true
WHATSAPP_QR_TIMEOUT_SECONDS=60
WHATSAPP_RECONNECT_INTERVAL_SECONDS=5

# Webhook
WEBHOOK_TIMEOUT_SECONDS=30
WEBHOOK_MAX_RETRIES=3
WEBHOOK_RETRY_BACKOFF_SECONDS=1,2,4

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

#### Frontend (.env)
```bash
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080
```

---

## 9. Documentation Requirements

### 9.1 README.md (Root)
```markdown
# WhatsApp Multi-Session API Gateway

Enterprise-grade WhatsApp Web API with multi-session support and n8n webhook integration.

## Features
- 🚀 Multi-user & multi-session support
- 📱 Real-time QR code updates
- 🤖 AI backend integration via webhooks
- 📊 Comprehensive analytics
- 🔒 Secure PIN-based authentication
- ⚡ High performance & low memory footprint

## Quick Start
1. Clone repository
2. `docker-compose up -d`
3. Open http://localhost:3000
4. Generate PIN & start using!

## Documentation
- [API Documentation](./backend/docs/API.md)
- [cURL Collection](./backend/docs/CURL_COLLECTION.md)
- [How to Use](./HOW_TO_USE.md)
- [Architecture](./backend/docs/ARCHITECTURE.md)
```

---

### 9.2 backend/README.md
```markdown
# WhatsApp API Backend

Go-based WhatsApp Web API server.

## Setup

### Local Development
\`\`\`bash
# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Run migrations
make migrate-up

# Run server
make run
\`\`\`

### Docker
\`\`\`bash
docker build -t whatsapp-backend .
docker run -p 8080:8080 whatsapp-backend
\`\`\`

## Testing
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint
```

## Project Structure
See [ARCHITECTURE.md](./docs/ARCHITECTURE.md) for details.
```

---

### 9.3 frontend/README.md
```markdown
# WhatsApp API Frontend

React-based dashboard for managing WhatsApp sessions.

## Setup

### Local Development
```bash
# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Run dev server
npm run dev
```

### Production Build
```bash
npm run build
npm run preview
```

### Docker
```bash
docker build -t whatsapp-frontend .
docker run -p 3000:80 whatsapp-frontend
```

## Available Scripts
- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
```

---

### 9.4 HOW_TO_USE.md
```markdown
# How to Use WhatsApp Multi-Session API

## Step-by-Step Guide

### 1. Generate Your PIN
1. Open the application at http://localhost:3000
2. Click **"Generate PIN"** button
3. **IMPORTANT**: Save the generated PIN (e.g., "A1B2C3")
4. You'll need this PIN to login

### 2. Login
1. Go to the Login page
2. Enter your 6-character PIN
3. Click **"Login"**

### 3. Create Your First Session
1. On the Dashboard, click **"+ New Session"**
2. Fill in:
   - **Session Name**: e.g., "Customer Support Bot"
   - **Webhook URL**: Your n8n webhook URL (e.g., "https://n8n.yourapp.com/webhook/whatsapp")
3. Click **"Create"**

### 4. Connect WhatsApp
1. Find your newly created session in the table
2. Click **"Start"** button
3. A QR code will appear in a modal
4. Open WhatsApp on your phone
5. Go to **Settings** > **Linked Devices** > **Link a Device**
6. Scan the QR code
7. Wait for "Connected!" message
8. Modal will close automatically

### 5. Configure Group Messages (Important!)
- **Private messages**: All messages will be forwarded to your webhook
- **Group messages**: Only messages where your bot is @mentioned will be forwarded
- Example: In a group, someone types "@628123456789 help me" → forwarded
- Example: In a group, someone types "Hello everyone" → NOT forwarded

### 6. Send Messages via API
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "6281234567890",
    "type": "text",
    "content": "Hello from API!"
  }'
```

### 7. View Analytics
1. Click **"View Analytics"** on any session
2. Select date range
3. See:
   - Total messages processed
   - AI success rate
   - Response times
   - Private vs Group breakdown
   - Peak usage hours

### 8. Managing Sessions
- **Edit**: Change session name or webhook URL
- **Reconnect**: If connection is lost, force reconnect
- **Delete**: Remove session and logout from WhatsApp
- **Logout**: Use logout button in header to logout from dashboard

## Troubleshooting

### QR Code Not Showing
- Check if session status is "disconnected" or "qr"
- Try clicking "Start" again
- Check backend logs for errors

### Messages Not Being Forwarded
- Verify webhook URL is correct
- Check n8n webhook is running
- For groups: Make sure bot is @mentioned
- Check analytics for webhook errors

### Session Disconnected
- Click "Reconnect" button
- If that fails, delete session and create new one
- WhatsApp may disconnect if phone is offline for long time

### Can't Login
- Verify PIN is exactly 6 characters
- PIN is case-sensitive
- Generate new PIN if forgotten (old one will be invalid)

## API Usage Examples

See [CURL_COLLECTION.md](./backend/docs/CURL_COLLECTION.md) for complete API examples.

## Support

For issues, please check:
1. Backend logs: `docker logs whatsapp_backend`
2. Frontend console: Browser DevTools
3. Database: Check PostgreSQL logs

## Best Practices

1. **Save your PIN immediately** - There's no password recovery
2. **Use HTTPS** for webhook URLs in production
3. **Monitor analytics** regularly for performance issues
4. **Set rate limits** on n8n to prevent spam
5. **Keep sessions connected** - Disconnect/reconnect too often may trigger WhatsApp limits
6. **One phone number per session** - Don't use same WhatsApp account on multiple sessions

---

**Need help?** Contact support or check the documentation.
```

---

## 10. Testing Strategy

### 10.1 Backend Testing

#### Unit Tests
```go
// internal/service/auth_service_test.go
func TestGeneratePIN(t *testing.T) {
    service := NewAuthService(mockRepo)
    
    pin, err := service.GeneratePIN()
    
    assert.NoError(t, err)
    assert.Len(t, pin, 6)
    assert.Regexp(t, "^[A-Z0-9]{6}$", pin)
}

func TestLoginWithValidPIN(t *testing.T) {
    service := NewAuthService(mockRepo)
    
    token, err := service.Login("ABC123")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
}
```

#### Integration Tests
```go
// internal/handler/session_handler_test.go
func TestCreateSession(t *testing.T) {
    // Setup test server
    router := setupTestRouter()
    
    // Create request
    body := `{"session_name":"Test","webhook_url":"https://test.com"}`
    req := httptest.NewRequest("POST", "/api/v1/sessions", strings.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+testToken)
    
    // Execute
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, resp.Code)
}
```

#### Test Coverage Target
- **Repository Layer**: 80%+
- **Service Layer**: 70%+
- **Handler Layer**: 60%+
- **Overall**: 60%+

---

### 10.2 Frontend Testing (Optional for v1.0)

```javascript
// src/components/__tests__/SessionCard.test.jsx
import { render, screen, fireEvent } from '@testing-library/react';
import SessionCard from '../SessionCard';

test('renders session card with correct data', () => {
  const session = {
    session_name: 'Test Session',
    status: 'connected',
    phone_number: '628123456789'
  };
  
  render(<SessionCard session={session} />);
  
  expect(screen.getByText('Test Session')).toBeInTheDocument();
  expect(screen.getByText('Connected')).toBeInTheDocument();
});
```

---

### 10.3 Manual Testing Checklist

#### Authentication Flow
- [ ] Generate PIN successfully
- [ ] PIN is unique (try generating multiple)
- [ ] Login with correct PIN works
- [ ] Login with wrong PIN fails
- [ ] Logout works
- [ ] Token persists in localStorage

#### Session Management
- [ ] Create new session
- [ ] Session name validation (empty, too long)
- [ ] Webhook URL validation (valid URL format)
- [ ] List all sessions
- [ ] Edit session (name and webhook)
- [ ] Delete session with confirmation
- [ ] Multiple sessions per user

#### WhatsApp Connection
- [ ] Start session generates QR code
- [ ] QR code displays in modal
- [ ] Real-time QR updates via WebSocket
- [ ] Scan QR code connects successfully
- [ ] Modal auto-closes on connection
- [ ] Status badge updates (disconnected → qr → connected)
- [ ] Phone number displays after connection
- [ ] Reconnect button works
- [ ] Auto-reconnect on disconnect (backend)

#### Message Handling
- [ ] Send text message via API
- [ ] Send image message via API
- [ ] Receive private message → forwarded to webhook
- [ ] Receive group message WITHOUT mention → NOT forwarded
- [ ] Receive group message WITH mention → forwarded to webhook
- [ ] Webhook receives correct payload format
- [ ] Webhook retry on failure (check logs)
- [ ] Analytics records all events

#### Analytics
- [ ] View session analytics
- [ ] Date range filter works
- [ ] Charts display correctly
- [ ] Stats cards show accurate numbers
- [ ] Private vs Group ratio correct
- [ ] Peak hours chart accurate
- [ ] Daily breakdown table
- [ ] Overview analytics (all sessions)

#### Performance
- [ ] Dashboard loads < 2s
- [ ] API responses < 200ms
- [ ] WebSocket connection stable
- [ ] Multiple sessions don't slow down
- [ ] Frontend bundle size < 500KB gzipped
- [ ] No memory leaks (check DevTools)

#### Error Handling
- [ ] Invalid webhook URL shows error
- [ ] Network error shows toast
- [ ] Token expiry redirects to login
- [ ] QR timeout shows refresh button
- [ ] Delete confirmation modal
- [ ] Form validation errors

#### Edge Cases
- [ ] Session name with special characters
- [ ] Very long webhook URLs
- [ ] Rapid create/delete sessions
- [ ] Multiple browser tabs open
- [ ] Disconnect phone while connected
- [ ] Backend restart while session active
- [ ] Database connection loss

---

## 11. Success Metrics

### 11.1 Technical Metrics
| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| API Uptime | 99.9% | Health check monitoring |
| Avg Response Time | < 200ms | Application logs |
| Memory per Session | < 50MB | Docker stats |
| Webhook Delivery | > 95% | Analytics table |
| Database Query Time | < 50ms | PostgreSQL slow query log |
| Frontend Bundle Size | < 500KB | Webpack/Vite build output |
| Test Coverage | > 60% | go test -cover |

### 11.2 User Experience Metrics
| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| QR Scan to Connected | < 5s | WebSocket timing logs |
| Dashboard Load Time | < 2s | Lighthouse/DevTools |
| Analytics Load Time | < 3s | Frontend performance monitoring |
| Session Creation | < 3s | End-to-end timing |

---

## 12. Future Enhancements (Out of Scope v1.0)

### Phase 2 (Next 3 months)
- [ ] Message templates (quick replies)
- [ ] Scheduled messages
- [ ] Broadcast messages to multiple recipients
- [ ] Advanced filtering in analytics
- [ ] Export analytics (CSV/PDF)
- [ ] Webhook signature verification (HMAC)
- [ ] API key authentication (alternative to PIN)

### Phase 3 (6-12 months)
- [ ] WhatsApp Business API support
- [ ] Multi-language support (i18n)
- [ ] Mobile app (React Native)
- [ ] Two-factor authentication (2FA)
- [ ] Role-based access control (RBAC)
- [ ] Team collaboration features
- [ ] Message queue (Redis/RabbitMQ)
- [ ] Horizontal scaling with load balancer
- [ ] Advanced monitoring (Prometheus/Grafana)
- [ ] Auto-scaling based on load

### Long-term Ideas
- [ ] AI-powered message suggestions
- [ ] Chatbot builder (no-code)
- [ ] CRM integration (Salesforce, HubSpot)
- [ ] Payment gateway integration
- [ ] Voice message support
- [ ] Video call integration
- [ ] Message encryption at rest

---

## 13. Timeline Estimation

### Week 1-2: Backend Core (16-20 hours)
- [x] Project setup & folder structure
- [x] Database schema & migrations
- [x] Authentication system (PIN generation, login)
- [x] Session CRUD operations
- [x] WhatsApp integration (whatsmeow)
- [x] QR code generation

### Week 2-3: Webhook & Message Handling (12-16 hours)
- [x] Webhook service (send to n8n)
- [x] Message handler (incoming messages)
- [x] Group mention filtering logic
- [x] Retry mechanism
- [x] Message sending API
- [x] WebSocket implementation

### Week 3: Analytics (8-10 hours)
- [x] Analytics data collection
- [x] Analytics aggregation queries
- [x] Analytics API endpoints
- [x] Stats calculation logic

### Week 4: Frontend Development (16-20 hours)
- [x] Project setup (React + Vite + Tailwind)
- [x] Authentication pages (Generate PIN, Login)
- [x] Dashboard page & Session cards
- [x] QR Code modal with WebSocket
- [x] Session form (create/edit)
- [x] Analytics page with charts

### Week 5: Testing & Documentation (12-16 hours)
- [x] Unit tests (backend critical functions)
- [x] Integration tests (API endpoints)
- [x] Manual testing checklist
- [x] API documentation (API.md)
- [x] cURL collection
- [x] HOW_TO_USE guide
- [x] README files

### Week 6: Polish & Deployment (8-12 hours)
- [x] Docker Compose setup
- [x] Environment variables setup
- [x] Production configuration
- [x] Performance optimization
- [x] Security hardening
- [x] Final bug fixes
- [x] Deployment documentation

**Total Estimated Time: 72-94 hours (5-6 weeks for 1 developer)**

---

## 14. Dependencies

### 14.1 Backend (Go)
```go
// go.mod
module github.com/yourusername/whatsapp-api

go 1.21

require (
    github.com/gorilla/mux v1.8.1
    github.com/gorilla/websocket v1.5.1
    github.com/lib/pq v1.10.9
    github.com/joho/godotenv v1.5.1
    go.mau.fi/whatsmeow v0.0.0-20240101000000-abcdef123456
    google.golang.org/protobuf v1.31.0
    github.com/golang-migrate/migrate/v4 v4.17.0
    github.com/rs/cors v1.10.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/google/uuid v1.5.0
    go.uber.org/zap v1.26.0
    golang.org/x/time v0.5.0 // rate limiting
)
```

### 14.2 Frontend (React)
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.1",
    "axios": "^1.6.2",
    "recharts": "^2.10.3",
    "react-qr-code": "^2.0.12",
    "sonner": "^1.2.3",
    "lucide-react": "^0.294.0",
    "clsx": "^2.0.0",
    "date-fns": "^2.30.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.2.1",
    "vite": "^5.0.8",
    "tailwindcss": "^3.3.6",
    "postcss": "^8.4.32",
    "autoprefixer": "^10.4.16",
    "eslint": "^8.56.0",
    "eslint-plugin-react": "^7.33.2",
    "eslint-plugin-react-hooks": "^4.6.0"
  }
}
```

---

## 15. Risk Assessment & Mitigation

### High Risk
| Risk | Impact | Mitigation |
|------|--------|------------|
| WhatsApp ban/block | Critical | Follow WhatsApp usage limits, implement rate limiting |
| Webhook endpoint down | High | Implement retry mechanism, log failures |
| Database connection loss | High | Connection pooling, auto-reconnect, health checks |

### Medium Risk
| Risk | Impact | Mitigation |
|------|--------|------------|
| Memory leak in long-running sessions | Medium | Proper cleanup, monitoring, periodic restarts |
| Concurrent session limit | Medium | Document limits, implement queue system if needed |
| QR code expiry issues | Medium | Implement refresh mechanism, clear user messaging |

### Low Risk
| Risk | Impact | Mitigation |
|------|--------|------------|
| Frontend browser compatibility | Low | Use modern evergreen browsers only |
| Timezone issues in analytics | Low | Store all timestamps in UTC |

---

## 16. Monitoring & Observability

### 16.1 Health Check Endpoint
```go
// GET /health
{
    "status": "healthy",
    "database": "connected",
    "active_sessions": 42,
    "memory_usage_mb": 512,
    "uptime_seconds": 86400
}
```

### 16.2 Logging Strategy
```go
// Structured logging with zap
logger.Info("Session created",
    zap.String("session_id", sessionID),
    zap.String("user_id", userID),
    zap.String("session_name", name),
)

logger.Error("Webhook failed",
    zap.String("session_id", sessionID),
    zap.String("webhook_url", url),
    zap.Int("status_code", statusCode),
    zap.Error(err),
)
```

### 16.3 Metrics to Track
- Active sessions count
- Messages processed per minute
- Webhook success/failure rate
- Average webhook response time
- Database connection pool usage
- Memory usage per session
- API endpoint latencies

---

## 17. Security Considerations

### 17.1 OWASP Top 10 Coverage
- ✅ **A01 - Broken Access Control**: PIN-based auth, session validation
- ✅ **A02 - Cryptographic Failures**: HTTPS in production, secure token storage
- ✅ **A03 - Injection**: Prepared statements, input validation
- ✅ **A04 - Insecure Design**: Rate limiting, proper error handling
- ✅ **A05 - Security Misconfiguration**: Environment-based config, minimal exposure
- ✅ **A06 - Vulnerable Components**: Regular dependency updates
- ✅ **A07 - Auth Failures**: Proper token expiry, logout functionality
- ✅ **A08 - Data Integrity**: Database constraints, ACID transactions
- ✅ **A09 - Logging Failures**: Structured logging, error tracking
- ✅ **A10 - SSRF**: Webhook URL validation (future enhancement)

### 17.2 Production Security Checklist
- [ ] Change default database credentials
- [ ] Use strong JWT secret (minimum 32 characters)
- [ ] Enable HTTPS (Let's Encrypt/CloudFlare)
- [ ] Set proper CORS origins (no wildcard *)
- [ ] Implement rate limiting per IP
- [ ] Regular database backups
- [ ] Monitor failed login attempts
- [ ] Use secrets management (Vault/AWS Secrets Manager)
- [ ] Enable PostgreSQL SSL connections
- [ ] Implement request logging for audit trail

---

## 18. License & Legal

### 18.1 License
- **Backend**: MIT License
- **Frontend**: MIT License
- **Dependencies**: Check individual library licenses

### 18.2 WhatsApp Terms of Service
⚠️ **Important**: This project uses WhatsApp Web protocol via whatsmeow library. Users must comply with:
- WhatsApp Terms of Service
- WhatsApp Business Policy (if using for business)
- No spamming or bulk messaging
- Respect rate limits to avoid bans

**Disclaimer**: This is an unofficial API. Use at your own risk. The developers are not responsible for any WhatsApp account bans or violations.

---

## 19. Support & Contribution

### 19.1 Getting Help
- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check HOW_TO_USE.md and API.md
- **Discord/Slack**: Community support (future)

### 19.2 Contributing
```markdown
# Contributing Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Code Style
- Backend: Run `gofmt` and `golangci-lint`
- Frontend: Run `npm run lint`
- Write tests for new features
- Update documentation

## Commit Messages
- Use conventional commits (feat:, fix:, docs:, etc.)
- Be descriptive but concise
```

---

## 20. Appendix

### 20.1 Glossary
- **PIN**: 6-character alphanumeric authentication code
- **Session**: Single WhatsApp Web connection instance
- **Webhook**: HTTP endpoint that receives WhatsApp events
- **QR Code**: Quick Response code for WhatsApp pairing
- **Mention**: @-tagging in WhatsApp groups
- **n8n**: Workflow automation tool (like Zapier)

### 20.2 Reference Links
- WhatsApp Web Protocol: https://github.com/tulir/whatsmeow
- Go WhatsApp Reference: https://github.com/aldinokemal/go-whatsapp-web-multidevice
- React Documentation: https://react.dev
- Vite Documentation: https://vitejs.dev
- Tailwind CSS: https://tailwindcss.com
- Recharts: https://recharts.org
- PostgreSQL: https://www.postgresql.org/docs/

### 20.3 FAQ

**Q: Can I use multiple WhatsApp accounts?**
A: Yes, create multiple sessions with different QR codes.

**Q: Will my WhatsApp account be banned?**
A: Follow WhatsApp ToS, don't spam, respect rate limits. No guarantees.

**Q: Can I self-host this?**
A: Yes, fully self-hostable with Docker.

**Q: Does this support WhatsApp Business API?**
A: v1.0 uses WhatsApp Web. Business API planned for v2.0.

**Q: What happens if I lose my PIN?**
A: Generate a new one. Old sessions will need to be recreated.

**Q: Can I run this on shared hosting?**
A: Need VPS/dedicated server with Docker support.

**Q: How many messages can I send per day?**
A: Depends on WhatsApp limits (~1000-5000/day). Don't spam.

---

## 21. Contact Information

**Project Maintainer**: [Your Name]
**Email**: [your.email@example.com]
**GitHub**: [https://github.com/yourusername/whatsapp-api]
**Documentation**: [https://docs.yourapp.com]

---

**Version**: 1.0.0  
**Last Updated**: December 1, 2024  
**Status**: Ready for Development 🚀

---