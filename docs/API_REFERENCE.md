# CRM Services - Frontend API Reference

Base URL: `/v1`

All protected endpoints require the `auth_token` cookie (set automatically on login/register).
All request/response bodies are `application/json`.

---

## Table of Contents

- [Authentication](#authentication)
- [Members](#members)
- [Plans & Usage](#plans--usage)
- [Subscriptions & Billing](#subscriptions--billing)
- [Payments](#payments)
- [Roles & Permissions](#roles--permissions)
- [Contacts](#contacts)
- [Organizations](#organizations)
- [Chat](#chat)
- [Widget](#widget)
- [API Keys](#api-keys)
- [Webhooks (Outgoing)](#webhooks-outgoing)
- [Webhooks (Incoming Tokens)](#webhooks-incoming-tokens)
- [Error Handling](#error-handling)
- [Rate Limits](#rate-limits)

---

## Authentication

### POST `/v1/register`

Create a new account and organization.

**Rate limit:** Auth (5 req/min per IP)

**Request body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepass123",
  "organization_name": "My Company"
}
```

All fields are **required**.

**Response:** `201 Created`
```json
{
  "user": {
    "id": 1,
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "admin",
    "status": "active",
    "organization_id": "660e8400-e29b-41d4-a716-446655440001"
  }
}
```

Sets `auth_token` cookie (HttpOnly, Secure, SameSite).

The new organization is automatically assigned to the **Free plan** with system roles (Owner, Admin, Member) created. The registering user is assigned the **Owner** role.

---

### POST `/v1/login`

**Rate limit:** Auth (5 req/min per IP)

**Request body:**
```json
{
  "email": "john@example.com",
  "password": "securepass123"
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "admin",
    "status": "active",
    "organization_id": "660e8400-e29b-41d4-a716-446655440001"
  }
}
```

Sets `auth_token` cookie.

---

### POST `/v1/logout`

Clears the `auth_token` cookie.

**Response:** `200 OK`
```json
{
  "message": "logged out successfully"
}
```

---

### POST `/v1/invite/accept`

Accept a member invitation and activate the account.

**Rate limit:** Auth (5 req/min per IP)

**Request body:**
```json
{
  "token": "invite-token-from-email",
  "password": "securepass123"
}
```

- `token` **required** - The invite token received via email
- `password` **required** - Min 8 characters

**Response:** `200 OK`
```json
{
  "message": "Account activated successfully",
  "data": {
    "id": 2,
    "uuid": "...",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "role": "member",
    "status": "active",
    "organization_id": "..."
  }
}
```

---

## Members

All member endpoints require JWT authentication and specific permissions.

### POST `/v1/members/invite`

Invite a new member to the organization.

**Permission:** `members:invite`

**Request body:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com"
}
```

**Response:** `201 Created`
```json
{
  "message": "User invited successfully",
  "data": {
    "id": 2,
    "uuid": "...",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "role": "member",
    "status": "pending",
    "organization_id": "..."
  }
}
```

**Errors:**
- `403` - Member limit reached for current plan (`"Member limit reached for your current plan. Please upgrade to invite more members."`)

---

### GET `/v1/members`

List all members of the organization with their roles.

**Permission:** `members:list`

**Response:** `200 OK`
```json
{
  "members": [
    {
      "id": 1,
      "uuid": "...",
      "name": "John Doe",
      "email": "john@example.com",
      "status": "active",
      "role_name": "owner",
      "role_id": 1,
      "organization_id": "...",
      "joined_at": "2026-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "uuid": "...",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "status": "pending",
      "role_name": "member",
      "role_id": 3,
      "organization_id": "...",
      "joined_at": "2026-03-20T14:00:00Z"
    }
  ]
}
```

---

### DELETE `/v1/members/{userID}`

Remove a member from the organization. The **owner cannot be removed**.

**Permission:** `members:remove`

**Response:** `200 OK`
```json
{
  "message": "Member removed successfully"
}
```

---

### POST `/v1/members/{userID}/deactivate`

Deactivate a member (blocks login). The **owner cannot be deactivated**. Deactivated members still count toward the member limit.

**Permission:** `members:deactivate`

**Response:** `200 OK`
```json
{
  "message": "Member deactivated successfully"
}
```

---

### POST `/v1/members/{userID}/reactivate`

Reactivate a previously deactivated member.

**Permission:** `members:deactivate`

**Response:** `200 OK`
```json
{
  "message": "Member reactivated successfully"
}
```

---

### PATCH `/v1/members/{userID}/role`

Change a member's role.

**Permission:** `members:manage_role`

**Request body:**
```json
{
  "role_id": 2
}
```

**Response:** `200 OK`
```json
{
  "message": "Role assigned successfully"
}
```

---

## Plans & Usage

### GET `/v1/plans`

List all available plans.

**Permission:** `plans:read`

**Response:** `200 OK`
```json
[
  {
    "id": "uuid-free-plan",
    "name": "free",
    "display_name": "Free",
    "price_cents": 0,
    "currency": "BRL",
    "max_contacts": 100,
    "max_members": 3,
    "max_chat_responders": 1
  },
  {
    "id": "uuid-pro-plan",
    "name": "pro",
    "display_name": "Pro",
    "price_cents": 4990,
    "currency": "BRL",
    "max_contacts": 1000,
    "max_members": 10,
    "max_chat_responders": 5
  },
  {
    "id": "uuid-business-plan",
    "name": "business",
    "display_name": "Business",
    "price_cents": 14990,
    "currency": "BRL",
    "max_contacts": 10000,
    "max_members": 50,
    "max_chat_responders": 20
  }
]
```

---

### GET `/v1/organizations/{id}/usage`

Get the organization's current plan, subscription info, and resource usage.

**Permission:** `plans:read`

**Path param:** `id` - Organization UUID

**Response:** `200 OK`
```json
{
  "plan": {
    "id": "uuid-free-plan",
    "name": "free",
    "display_name": "Free",
    "price_cents": 0,
    "currency": "BRL",
    "max_contacts": 100,
    "max_members": 3,
    "max_chat_responders": 1
  },
  "subscription": {
    "status": "active",
    "current_period_start": "2026-01-01T00:00:00Z",
    "current_period_end": "2026-02-01T00:00:00Z"
  },
  "usage": {
    "contacts": {
      "current": 85,
      "limit": 100,
      "warning": "approaching_limit"
    },
    "members": {
      "current": 3,
      "limit": 3,
      "warning": "limit_reached"
    },
    "chat_responders": {
      "current": 0,
      "limit": 1
    }
  }
}
```

**Warning levels:**
| Warning | Threshold |
|---|---|
| _(empty)_ | Below 80% |
| `approaching_limit` | 80%-89% |
| `near_limit` | 90%-99% |
| `limit_reached` | 100% |

---

## Subscriptions & Billing

### POST `/v1/subscriptions`

Subscribe the organization to a paid plan via Mercado Pago transparent checkout.

**Permission:** `subscriptions:manage`

**Request body:**
```json
{
  "plan_id": "uuid-pro-plan",
  "payer_email": "billing@company.com",
  "card_token_id": "mp-card-token-from-sdk"
}
```

- `plan_id` **required** - UUID of the target plan (from `GET /plans`)
- `payer_email` **required** - Email for Mercado Pago billing
- `card_token_id` **required** - Token generated by the Mercado Pago JS SDK (transparent checkout)

**Response:** `201 Created`
```json
{
  "id": "uuid-subscription",
  "plan_id": "uuid-pro-plan",
  "status": "active",
  "current_period_start": "2026-03-30T12:00:00Z",
  "current_period_end": "2026-04-30T12:00:00Z"
}
```

**Errors:**
- `400` - `"organization already has an active paid subscription"`
- `400` - `"cannot subscribe to the free plan via payment"`
- `400` - `"plan not found"`

**Frontend integration:**
1. Use the [Mercado Pago JS SDK](https://www.mercadopago.com.br/developers/en/docs/checkout-api/landing) to collect card data
2. Generate a `card_token_id` via `mp.createCardToken()`
3. Send the token to this endpoint

---

### POST `/v1/subscriptions/upgrade`

Upgrade to a higher-tier plan. The current subscription is cancelled and a new one is created on the higher plan.

**Permission:** `subscriptions:manage`

**Request body:**
```json
{
  "plan_id": "uuid-business-plan",
  "payer_email": "billing@company.com",
  "card_token_id": "mp-card-token-from-sdk"
}
```

Same fields as Subscribe.

**Response:** `200 OK`
```json
{
  "id": "uuid-subscription",
  "plan_id": "uuid-business-plan",
  "status": "active",
  "current_period_start": "2026-03-30T14:00:00Z",
  "current_period_end": "2026-04-30T14:00:00Z"
}
```

**Errors:**
- `400` - `"organization is already on this plan"`
- `400` - `"downgrade must be done via cancellation"` (cannot upgrade to a cheaper plan)

---

### POST `/v1/subscriptions/cancel`

Cancel the current paid subscription. The organization **retains full access until the end of the current billing period**, then is automatically downgraded to the Free plan.

**Permission:** `subscriptions:manage`

**Request body:** _(none)_

**Response:** `200 OK`
```json
{
  "message": "subscription cancelled"
}
```

**Lifecycle after cancel:**
1. Status remains `active` with `cancel_at_period_end = true`
2. Organization keeps paid plan features until `current_period_end`
3. When Mercado Pago confirms cancellation via webhook, status transitions to `cancelled`
4. Once the billing period ends, the system automatically downgrades to the Free plan

---

## Payments

### GET `/v1/payments`

List all payment history for the organization.

**Permission:** `payments:read`

**Response:** `200 OK`
```json
[
  {
    "id": "uuid-payment",
    "status": "approved",
    "status_detail": "accredited",
    "amount_cents": 4990,
    "currency": "BRL",
    "payment_method": "credit_card",
    "description": "Pro - CRM Services",
    "paid_at": "2026-03-01T10:00:00Z",
    "created_at": "2026-03-01T10:00:00Z"
  }
]
```

**Payment statuses:**
| Status | Description |
|---|---|
| `approved` | Payment successful |
| `pending` | Waiting for processing |
| `rejected` | Payment declined |
| `refunded` | Payment refunded or charged back |

---

## Roles & Permissions

### GET `/v1/permissions`

List all available permissions in the system. **No authentication required** (but must be within the protected route group).

**Response:** `200 OK`
```json
[
  {
    "name": "contacts:create",
    "description": "Create contacts",
    "category": "contacts"
  },
  {
    "name": "contacts:read",
    "description": "View contacts",
    "category": "contacts"
  },
  {
    "name": "members:invite",
    "description": "Invite members",
    "category": "members"
  }
]
```

---

### GET `/v1/roles`

List all roles for the organization.

**Permission:** `roles:read`

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "uuid": "...",
    "name": "owner",
    "description": "Organization owner with full access",
    "is_system": true,
    "permissions": [
      { "id": 1, "name": "contacts:create", "description": "Create contacts", "category": "contacts" }
    ],
    "created_at": "2026-01-15T10:30:00Z"
  },
  {
    "id": 4,
    "uuid": "...",
    "name": "Sales Rep",
    "description": "Custom role for sales team",
    "is_system": false,
    "permissions": [],
    "created_at": "2026-03-20T14:00:00Z"
  }
]
```

---

### POST `/v1/roles`

Create a custom role.

**Permission:** `roles:manage`

**Request body:**
```json
{
  "name": "Sales Rep",
  "description": "Custom role for sales team",
  "permissions": ["contacts:create", "contacts:read", "contacts:update", "chats:read"]
}
```

- `name` **required**
- `permissions` **required** - At least one permission (use names from `GET /permissions`)

**Response:** `201 Created` - Role object

---

### PATCH `/v1/roles/{roleID}`

Update a custom role. System roles (Owner, Admin, Member) **cannot** be modified.

**Permission:** `roles:manage`

**Request body:**
```json
{
  "name": "Senior Sales Rep",
  "description": "Updated description",
  "permissions": ["contacts:create", "contacts:read", "contacts:update", "contacts:delete"]
}
```

- `name` **required**

**Response:** `200 OK` - Updated role object

---

### DELETE `/v1/roles/{roleID}`

Delete a custom role. System roles **cannot** be deleted.

**Permission:** `roles:manage`

**Response:** `200 OK`

---

## Contacts

### POST `/v1/contacts`

Create a new contact.

**Permission:** `contacts:create`

**Request body:**
```json
{
  "type": "person",
  "first_name": "Maria",
  "last_name": "Silva",
  "email": "maria@example.com",
  "phone": "+5511999990000",
  "mobile_phone": "+5511999990001",
  "company_name": "Acme Corp",
  "job_title": "CTO",
  "department": "Engineering",
  "street": "Rua Example",
  "number": "123",
  "complement": "Sala 4",
  "district": "Centro",
  "city": "Sao Paulo",
  "state": "SP",
  "zip_code": "01001-000",
  "country": "BR",
  "status": "lead",
  "source": "website",
  "tags": ["vip", "tech"],
  "notes": "Met at conference",
  "assigned_to_id": 1
}
```

- `type` **required** - `"person"` or `"company"`
- If `person`: `first_name` **required**
- If `company`: `company_name` **required**
- `status` defaults to `"lead"` if omitted

**Response:** `201 Created` - Contact object

**Errors:**
- `403` - Contact limit reached for current plan

---

### GET `/v1/contacts`

List contacts with pagination and filters.

**Permission:** `contacts:read`

**Query params:**
| Param | Type | Description |
|---|---|---|
| `page` | int | Page number (default: 1) |
| `page_size` | int | Items per page (default: 20) |
| `status` | string | Filter by status |
| `type` | string | `person` or `company` |
| `source` | string | Filter by source |
| `city` | string | Filter by city |
| `state` | string | Filter by state |
| `country` | string | Filter by country |
| `tags` | string | Comma-separated tags |
| `assigned_to_id` | uint | Filter by assigned user |
| `created_by_id` | uint | Filter by creator |
| `created_after` | date | ISO date |
| `created_before` | date | ISO date |

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid-contact",
      "type": "person",
      "full_name": "Maria Silva",
      "first_name": "Maria",
      "last_name": "Silva",
      "email": "maria@example.com",
      "phone": "+5511999990000",
      "status": "lead",
      "source": "website",
      "tags": ["vip"],
      "address": {
        "city": "Sao Paulo",
        "state": "SP",
        "country": "BR"
      },
      "created_at": "2026-03-20T14:00:00Z",
      "updated_at": "2026-03-20T14:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 85,
  "total_pages": 5
}
```

---

### GET `/v1/contacts/search?q={query}`

Search contacts by name or email.

**Permission:** `contacts:read`

---

### GET `/v1/contacts/{id}`

Get a single contact by UUID.

**Permission:** `contacts:read`

---

### GET `/v1/contacts/email/{email}`

Get a contact by email address.

**Permission:** `contacts:read`

---

### PATCH `/v1/contacts/{id}`

Update a contact (partial update - only send fields to change).

**Permission:** `contacts:update`

**Request body:**
```json
{
  "first_name": "Maria Updated",
  "status": "customer",
  "tags": ["vip", "returning"]
}
```

---

### DELETE `/v1/contacts/{id}`

Soft-delete a contact (can be restored).

**Permission:** `contacts:delete`

---

### DELETE `/v1/contacts/{id}/permanent`

Permanently delete a contact.

**Permission:** `contacts:delete`

---

## Organizations

### POST `/v1/organizations`

**Permission:** `organizations:manage`

**Request body:**
```json
{
  "name": "My Company",
  "slug": "my-company",
  "email": "admin@company.com",
  "phone": "+5511999990000",
  "website": "https://company.com",
  "document_id": "12345678000100",
  "industry": "Technology",
  "settings": {}
}
```

---

### GET `/v1/organizations/{id}`

**Permission:** `organizations:read`

**Response:** `200 OK`
```json
{
  "id": "uuid-org",
  "name": "My Company",
  "slug": "my-company",
  "email": "admin@company.com",
  "phone": "+5511999990000",
  "website": "https://company.com",
  "document_id": "12345678000100",
  "industry": "Technology",
  "plan": "free",
  "plan_id": 1,
  "settings": {},
  "is_active": true,
  "created_at": "2026-01-15T10:30:00Z",
  "updated_at": "2026-01-15T10:30:00Z"
}
```

---

### GET `/v1/organizations/slug/{slug}`

**Permission:** `organizations:read`

---

### PATCH `/v1/organizations/{id}`

Partial update. Send only fields to change.

**Permission:** `organizations:update`

---

### DELETE `/v1/organizations/{id}`

Soft-delete.

**Permission:** `organizations:delete`

---

### DELETE `/v1/organizations/{id}/permanent`

Permanent delete.

**Permission:** `organizations:delete`

---

### POST `/v1/organizations/{id}/restore`

Restore a soft-deleted organization.

**Permission:** `organizations:manage`

---

## Chat

### WebSocket `/v1/ws/chat/{chatID}`

Connect as a CRM agent. Requires JWT authentication.

### WebSocket `/v1/ws/widget/{chatID}`

Connect as a widget visitor. No authentication required.

**WebSocket message format:**
```json
{
  "type": "message",
  "content": "Hello!",
  "sender_id": 1,
  "visitor_id": "",
  "chat_id": 42
}
```

---

### GET `/v1/chats`

List all chats for the authenticated user.

**Permission:** `chats:read`

---

### GET `/v1/chats/{chatID}`

Get a single chat.

**Permission:** `chats:read`

**Response:** `200 OK`
```json
{
  "id": 1,
  "uuid": "...",
  "status": "open",
  "origin": "widget"
}
```

---

### GET `/v1/chats/{chatID}/messages`

Get messages for a chat.

**Permission:** `chats:read`

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "uuid": "...",
    "chat_id": 1,
    "sender_id": 1,
    "content": "Hello!",
    "type": "agent",
    "sender": {
      "id": 1,
      "uuid": "...",
      "name": "John Doe"
    },
    "created_at": "2026-03-20T14:00:00Z"
  }
]
```

---

## Widget

Widget endpoints require the `X-Widget-Key` header with a valid API key.

### POST `/v1/widget/init`

Initialize the chat widget.

**Headers:** `X-Widget-Key: your-public-api-key`

**Request body:**
```json
{
  "visitor_id": "visitor-fingerprint-123",
  "fingerprint": "browser-fp",
  "chat_id": null
}
```

**Response:** `200 OK`
```json
{
  "visitor_id": "visitor-fingerprint-123",
  "chat": {
    "id": 5,
    "uuid": "...",
    "status": "open",
    "origin": "widget"
  }
}
```

---

### POST `/v1/widget/chat`

Create a new widget chat.

**Headers:** `X-Widget-Key: your-public-api-key`

**Request body:**
```json
{
  "visitor_id": "visitor-fingerprint-123"
}
```

---

### GET `/v1/widget/chat/{chatID}/messages`

Get widget chat messages.

**Headers:** `X-Widget-Key: your-public-api-key`

---

## API Keys

### POST `/v1/api-keys`

Create a new API key for widget authentication.

**Permission:** `api_keys:create`

**Request body:**
```json
{
  "name": "Production Widget",
  "domain": "https://myapp.com"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "public_key": "pk_live_abc123...",
  "secret_key": "sk_live_xyz789...",
  "name": "Production Widget",
  "domain": "https://myapp.com",
  "is_active": true
}
```

> The `secret_key` is only returned on creation. Store it securely.

---

### GET `/v1/api-keys`

**Permission:** `api_keys:read`

---

### DELETE `/v1/api-keys/{keyID}`

**Permission:** `api_keys:delete`

---

## Webhooks (Outgoing)

### GET `/v1/webhooks/events`

List available webhook event types.

**Permission:** `webhooks:read`

**Response:** `200 OK`
```json
[
  "message.received",
  "message.sent",
  "chat.created",
  "chat.closed",
  "visitor.connected",
  "visitor.disconnected"
]
```

---

### POST `/v1/webhooks`

Create an outgoing webhook.

**Permission:** `webhooks:create`

**Request body:**
```json
{
  "name": "My Integration",
  "url": "https://myapp.com/webhook",
  "events": ["message.received", "chat.created"]
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "My Integration",
  "url": "https://myapp.com/webhook",
  "secret": "whsec_...",
  "events": ["message.received", "chat.created"],
  "is_active": true,
  "fail_count": 0,
  "created_at": "2026-03-20T14:00:00Z"
}
```

---

### GET `/v1/webhooks`

**Permission:** `webhooks:read`

---

### GET `/v1/webhooks/{webhookID}`

**Permission:** `webhooks:read`

---

### PUT `/v1/webhooks/{webhookID}`

**Permission:** `webhooks:update`

**Request body:**
```json
{
  "name": "Updated Name",
  "url": "https://myapp.com/webhook-v2",
  "events": ["message.received"],
  "is_active": true
}
```

---

### DELETE `/v1/webhooks/{webhookID}`

**Permission:** `webhooks:delete`

---

### GET `/v1/webhooks/{webhookID}/logs?limit=50`

Get delivery logs for a webhook.

**Permission:** `webhooks:read`

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "event_type": "message.received",
    "response_code": 200,
    "duration_ms": 150,
    "created_at": "2026-03-20T14:00:00Z"
  }
]
```

---

## Webhooks (Incoming Tokens)

### POST `/v1/webhooks/tokens`

Create a token for incoming webhooks.

**Permission:** `webhooks_tokens:create`

**Request body:**
```json
{
  "name": "External System"
}
```

---

### GET `/v1/webhooks/tokens`

**Permission:** `webhooks_tokens:read`

---

### DELETE `/v1/webhooks/tokens/{tokenID}`

**Permission:** `webhooks_tokens:delete`

---

### POST `/v1/webhook/incoming`

Process an incoming webhook. **Public endpoint** (rate-limited, requires token).

**Headers:** `X-Webhook-Token: your-token` (or query param `?token=your-token`)

**Request body:**
```json
{
  "action": "send_message",
  "chat_id": 1,
  "content": "Automated reply from external system",
  "data": {}
}
```

---

## Error Handling

All errors return a plain text body with an appropriate HTTP status code.

| Status | Meaning |
|---|---|
| `400` | Bad request / validation error |
| `401` | Not authenticated (missing or invalid `auth_token`) |
| `403` | Permission denied or plan limit reached |
| `404` | Resource not found |
| `429` | Rate limit exceeded |
| `500` | Internal server error |

**Plan limit errors (403):**
- `"Contact limit reached for your current plan. Please upgrade to add more contacts."`
- `"Member limit reached for your current plan. Please upgrade to invite more members."`
- `"Chat responder limit reached for your current plan. Please upgrade to add more responders."`

---

## Rate Limits

| Scope | Limit | Key |
|---|---|---|
| Auth endpoints | 5 req/min | Per IP |
| Widget endpoints | 30 req/min | Per API key |
| Webhook incoming | 60 req/min | Per token |
| API (all protected) | 100 req/min | Per user |

When rate-limited, the server returns `429 Too Many Requests`.

---

## Mercado Pago Webhook

### POST `/v1/webhook/mercadopago`

**Public endpoint** - Called by Mercado Pago to notify subscription and payment events. Do not call this from the frontend.

This endpoint processes:
- **Subscription events** (`subscription_preapproval`) - Updates subscription status (active, past_due, cancelled)
- **Payment events** (`payment`) - Creates/updates payment records and sends email notifications

The webhook is idempotent - duplicate notifications are handled safely.
