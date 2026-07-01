# Twitch & Kick Mock Server

## Overview

The Twitch & Kick Mock Server is a specialized utility for the Twir project. It enables developers to build, test, and debug features without requiring real Twitch or Kick accounts or dealing with rate limits. This service mimics several core Twitch and Kick components, providing a stable and predictable environment for local development.

By using this mock server, you avoid the need for complex OAuth flows with real servers and can trigger various events on demand. It is particularly useful for CI/CD pipelines and developers who prefer to work offline or in isolated environments.

## Credits

This project is based on the original Twitch mock server from the [Twir project](https://github.com/twirapp/twir):

- **Original Author:** [satont](https://github.com/satont)
- **Original Repository:** [github.com/twirapp/twir/apps/twitch-mock](https://github.com/twirapp/twir/tree/main/apps/twitch-mock)

Kick API support was added to extend the mock server's capabilities.

## Quick Start

Follow these steps to enable the mock server in your local Twir environment:

1. **Configure Environment:**
   Copy the mock configuration example to your local `.env` file:

   ```bash
   cp .env.mock.example .env
   ```

2. **Start Infrastructure:**
   Spin up the required background services, including the mock server, using Docker:

   ```bash
   docker compose -f docker-compose.dev.yml up -d
   ```

3. **Launch Twir:**
   Start the application in development mode:

   ```bash
   bun dev
   ```

4. **Login and Test:**
   Visit `http://localhost:3010` in your browser. You can now log in using the mock credentials.

## Running Locally (without Docker)

```bash
go run ./cmd/main.go
```

## Fake Users

The mock server comes pre-configured with two primary user accounts. Use these IDs and logins when testing features that require specific roles.

> **Note:** The `broadcaster_user_id` used in notification payloads and EventSub subscription conditions is the same value returned by the `GET /helix/users` endpoint (`"id"` field) and the token validation endpoint (`"user_id"` field). All derive from the same config constants, so the values are always consistent.

### Twitch

| Role        | ID    | Login        | Display Name |
| ----------- | ----- | ------------ | ------------ |
| Broadcaster | 12345 | mockstreamer | MockStreamer |
| Bot         | 67890 | mockbot      | MockBot      |

### Kick

| Role        | ID    | Name         |
| ----------- | ----- | ------------ |
| Broadcaster | 12345 | MockStreamer |
| Bot         | 67890 | MockBot      |

## Ports

The Mock Server exposes several ports for different functionalities:

| Port | Service                  | Description                                                           |
| ---- | ------------------------ | --------------------------------------------------------------------- |
| 7777 | HTTP: OAuth2 + APIs      | Handles authentication, Twitch Helix API, and Kick API requests.      |
| 8081 | WebSocket: EventSub      | Provides a WebSocket interface for receiving real-time Twitch events. |
| 3333 | Admin UI                 | A web interface for managing the mock server and triggering events.   |

## Mocked Endpoints

### Twitch

- **Authentication:** OAuth2 authorization and token exchange.
- **Users:** Fetching user profiles and following status.
- **Streams:** Getting stream information and metadata.
- **Channels:** Updating channel details and information.
- **Moderation:** Managing bans, timeouts, and moderators.
- **Subscriptions:** Checking subscriber lists and status.
- **EventSub:** Managing and receiving EventSub subscriptions.

### Kick

- **Authentication:** OAuth2 token exchange and authorization.
- **Token Introspection:** Validating access tokens.
- **Users:** Fetching user profiles.
- **Livestreams:** Getting livestream information and metadata.
- **Channels:** Fetching channel details.
- **Chat:** Sending chat messages.
- **Moderation:** Managing bans and unbans.
- **Categories:** Fetching category information.
- **Event Subscriptions:** Managing webhook subscriptions.

## Triggering Events

You can manually trigger events to test how the application responds to different scenarios.

### Admin UI

Access the graphical interface at:
`http://localhost:3333/admin`

The admin UI has tabs for both **Twitch** and **Kick** with separate controls for each platform.

### CLI / Curl Examples

#### Twitch Events

**Trigger a Follow Event (`channel.follow`):**

```bash
curl -X POST http://localhost:3333/admin/trigger/channel.follow \
  -H "Content-Type: application/json" \
  -d '{"user_id": "11111", "user_login": "testuser", "broadcaster_user_id": "12345"}'
```

**Trigger a Subscribe Event (`channel.subscribe`):**

```bash
curl -X POST http://localhost:3333/admin/trigger/channel.subscribe \
  -H "Content-Type: application/json" \
  -d '{"user_id": "11111", "user_login": "testuser", "broadcaster_user_id": "12345", "tier": "1000"}'
```

**Trigger a Cheer Event (`channel.cheer`):**

```bash
curl -X POST http://localhost:3333/admin/trigger/channel.cheer \
  -H "Content-Type: application/json" \
  -d '{"user_id": "11111", "user_login": "testuser", "broadcaster_user_id": "12345", "bits": 100}'
```

#### Kick Events

**Trigger a Follow Event (`follow`):**

```bash
curl -X POST http://localhost:3333/admin/kick/trigger/follow \
  -H "Content-Type: application/json" \
  -d '{"user_id": 99999, "username": "testuser", "broadcaster_user_id": 12345}'
```

**Trigger a Livestream Started Event (`livestream.started`):**

```bash
curl -X POST http://localhost:3333/admin/kick/trigger/livestream.started \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Add a Webhook Subscription:**

```bash
curl -X POST http://localhost:3333/admin/kick/webhooks \
  -H "Content-Type: application/json" \
  -d '{"url": "https://your-webhook-url.com", "events": ["follow", "subscription"]}'
```

## What Is NOT Mocked

Please be aware of the following limitations:

- **Twitch IRC/Chat Protocol:** This service does not mock the Twitch IRC servers. Twir connects to a real or local IRC instance separately.
- **Twitch Player Iframe:** Visual components like the Twitch video player are not included.
- **Advanced Helix Features:** Some niche or rarely used Helix endpoints might return empty results or 404s.
- **Kick Channel Rewards:** Channel rewards endpoints are not implemented.

## Troubleshooting

### Port already in use

If you see an error stating that port 7777, 8081, or 3333 is already in use, you must identify and kill the process occupying that port.

```bash
lsof -i :7777
kill -9 <PID>
```

### Login fails

Ensure that `TWITCH_MOCK_ENABLED=true` is set in your `.env` file. If this is missing or set to `false`, Twir will attempt to connect to the production Twitch API.

### EventSub not connecting

Verify that the mock server is healthy and running by checking the Docker logs:

```bash
docker compose -f docker-compose.dev.yml logs twitch-mock
```

Also, confirm that `TWITCH_MOCK_WS_URL` is correctly set to `ws://localhost:8081/ws`.

### Kick webhooks not receiving events

Ensure that you have added a webhook subscription via the admin UI or API:

```bash
curl -X POST http://localhost:3333/admin/kick/webhooks \
  -H "Content-Type: application/json" \
  -d '{"url": "https://your-webhook-url.com", "events": ["*"]}'
```

Then trigger events via the admin UI or:

```bash
curl -X POST http://localhost:3333/admin/kick/trigger/follow \
  -H "Content-Type: application/json" \
  -d '{"user_id": 99999, "username": "testuser", "broadcaster_user_id": 12345}'
```
