# Matcha Realtime Backend Guide

This document explains the websocket, chat, conversation, and notification code in the backend. Keep it open next to these files:

- `models.go`
- `db.go`
- `main.go`
- `websocket.go`
- `realtime.go`

The short version: HTTP loads history, websockets deliver live events.

## Big Picture

The realtime backend has four jobs:

1. Keep track of connected websocket clients.
2. Receive live chat messages from a connected user.
3. Save chat messages and notifications in Postgres.
4. Deliver live events to the recipient if they are online.

There are two kinds of communication:

- **HTTP REST endpoints**: good for loading data, such as conversations, messages, and notifications.
- **Websocket endpoint**: good for instant events, such as receiving a message while the app is open.

The frontend should use both.

Example flow:

1. User logs in.
2. Browser opens `GET /api/accounts/ws`.
3. Backend authenticates the websocket using the existing `access_token` cookie.
4. User sends a websocket JSON message.
5. Backend saves the message in the database.
6. Backend sends the message to the recipient if they are online.
7. If the recipient was offline, they can still fetch the saved message later with HTTP.

## Database Models

File: `models.go`

### `Conversation`

```go
type Conversation struct {
	gorm.Model
	UserOneID uint          `gorm:"not null;uniqueIndex:idx_conversation_users"`
	UserTwoID uint          `gorm:"not null;uniqueIndex:idx_conversation_users"`
	Messages  []ChatMessage `gorm:"constraint:OnDelete:CASCADE;"`
}
```

This represents one chat between two users.

- `gorm.Model` adds `ID`, `CreatedAt`, `UpdatedAt`, and `DeletedAt`.
- `UserOneID` is one participant.
- `UserTwoID` is the other participant.
- The two user IDs are stored in sorted order. For users `9` and `2`, the conversation is stored as `UserOneID = 2`, `UserTwoID = 9`.
- `uniqueIndex:idx_conversation_users` prevents duplicate conversations between the same pair.
- `Messages []ChatMessage` tells GORM that a conversation can have many messages.
- `OnDelete:CASCADE` means deleting a conversation deletes its messages too.

Why sorted IDs matter:

Without sorting, these two rows would look different but mean the same thing:

```text
user_one_id = 1, user_two_id = 2
user_one_id = 2, user_two_id = 1
```

Sorting prevents that.

### `ChatMessage`

```go
type ChatMessage struct {
	gorm.Model
	ConversationID uint       `gorm:"not null;index"`
	SenderID       uint       `gorm:"not null;index"`
	RecipientID    uint       `gorm:"not null;index"`
	Content        string     `gorm:"not null"`
	ReadAt         *time.Time `gorm:"index"`
}
```

This represents one chat message.

- `ConversationID` links the message to a conversation.
- `SenderID` is the user who sent it.
- `RecipientID` is the user who should receive it.
- `Content` is the message text.
- `ReadAt` is `nil` until the recipient marks the message as read.
- `*time.Time` means “pointer to a time”. A pointer can be `nil`, which is useful for unread messages.

### `Notification`

```go
type Notification struct {
	gorm.Model
	UserID   uint       `gorm:"not null;index"`
	SenderID *uint      `gorm:"index"`
	Type     string     `gorm:"not null;default:'notification'"`
	Title    string     `gorm:"not null;default:''"`
	Message  string     `gorm:"not null"`
	Data     string     `gorm:"type:jsonb;not null;default:'{}'"`
	ReadAt   *time.Time `gorm:"index"`
}
```

This represents a notification for one user.

- `UserID` is the notification owner.
- `SenderID` is optional. For system notifications, it can be `nil`.
- `Type` lets you categorize notifications later, for example `match`, `like`, `message`.
- `Title` is short display text.
- `Message` is the body text.
- `Data` stores extra JSON. Example: `{"profile_id": 7}`.
- `ReadAt` is `nil` until the user reads it.

## Database Migration

File: `db.go`

```go
err = DB.AutoMigrate(&User{}, &Photo{}, &Conversation{}, &ChatMessage{}, &Notification{})
```

`AutoMigrate` tells GORM to create or update the database tables for those models.

This means when the backend starts, it makes sure these tables exist:

- `users`
- `photos`
- `conversations`
- `chat_messages`
- `notifications`

## Routes

File: `main.go`

The hub is created and started here:

```go
hub := newWebsocketHub()
go hub.run()
```

- `newWebsocketHub()` creates the in-memory manager for live websocket clients.
- `go hub.run()` starts it in a goroutine.
- A goroutine is a lightweight concurrent task in Go.

The realtime routes are:

```go
privateMux.HandleFunc(pat.Get("/ws"), websocketHandler(hub))
privateMux.HandleFunc(pat.Post("/notifications/send"), notificationHandler(hub))
privateMux.HandleFunc(pat.Get("/conversations"), listConversations)
privateMux.HandleFunc(pat.Get("/conversations/:conversationID/messages"), listConversationMessages)
privateMux.HandleFunc(pat.Patch("/messages/:messageID/read"), markMessageRead)
privateMux.HandleFunc(pat.Get("/notifications"), listNotifications)
privateMux.HandleFunc(pat.Patch("/notifications/:notificationID/read"), markNotificationRead)
```

These are all under:

```go
mux.Handle(pat.New("/api/accounts/*"), privateMux)
```

So the real URLs are:

```text
GET   /api/accounts/ws
POST  /api/accounts/notifications/send
GET   /api/accounts/conversations
GET   /api/accounts/conversations/:conversationID/messages
PATCH /api/accounts/messages/:messageID/read
GET   /api/accounts/notifications
PATCH /api/accounts/notifications/:notificationID/read
```

Because they are in `privateMux`, they use your existing auth middleware.

## Websocket Concepts

File: `websocket.go`

### Why websockets?

Normal HTTP works like this:

1. Browser asks backend for something.
2. Backend replies.
3. Connection is done.

Websocket works like this:

1. Browser connects once.
2. Backend upgrades the connection.
3. Both sides can send messages anytime.
4. The connection stays open.

This is why websockets are useful for chat.

## Constants

```go
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)
```

- `writeWait`: maximum time allowed to write one event to the websocket.
- `pongWait`: maximum time allowed before the client must respond to a ping.
- `pingPeriod`: how often the server sends pings.
- `maxMessageSize`: biggest JSON websocket message accepted from the client.

The ping/pong system keeps dead connections from staying forever.

## `websocketEvent`

```go
type websocketEvent struct {
	Type           string          `json:"type"`
	ID             uint            `json:"id,omitempty"`
	ConversationID uint            `json:"conversation_id,omitempty"`
	RecipientID    uint            `json:"recipient_id,omitempty"`
	SenderID       uint            `json:"sender_id,omitempty"`
	Message        string          `json:"message,omitempty"`
	Title          string          `json:"title,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	CreatedAt      time.Time       `json:"created_at,omitempty"`
}
```

This is the JSON shape used for live websocket messages.

Example chat message from frontend:

```json
{
  "type": "chat_message",
  "recipient_id": 2,
  "message": "Hey!"
}
```

The backend fills in:

- `id`
- `conversation_id`
- `sender_id`
- `created_at`

Example event received by frontend:

```json
{
  "type": "chat_message",
  "id": 15,
  "conversation_id": 3,
  "recipient_id": 2,
  "sender_id": 1,
  "message": "Hey!",
  "created_at": "2026-05-02T15:00:00Z"
}
```

`json.RawMessage` means raw JSON bytes. It lets `Data` contain flexible JSON without needing a fixed struct.

`omitempty` means “do not include this field in JSON if it is empty”.

## `websocketClient`

```go
type websocketClient struct {
	hub    *websocketHub
	userID uint
	conn   *websocket.Conn
	send   chan websocketEvent
}
```

This represents one connected browser tab or device.

- `hub`: pointer to the central websocket hub.
- `userID`: authenticated user ID from the JWT cookie.
- `conn`: the actual websocket connection.
- `send`: channel used to queue events that should be written to this client.

A channel is a Go-safe queue for communication between goroutines.

## `websocketHub`

```go
type websocketHub struct {
	clients    map[uint]map[*websocketClient]bool
	register   chan *websocketClient
	unregister chan *websocketClient
	deliver    chan websocketEvent
}
```

This is the live connection manager.

- `clients`: maps user IDs to connected clients.
- `register`: channel for new clients.
- `unregister`: channel for disconnected clients.
- `deliver`: channel for events that should be sent to a user.

The nested map allows a user to be connected from multiple tabs/devices:

```text
user 2 -> tab A
user 2 -> tab B
user 2 -> phone
```

## Origin Check

```go
CheckOrigin: isAllowedWebsocketOrigin
```

This prevents random websites from opening authenticated websockets using your cookies.

The function:

```go
func isAllowedWebsocketOrigin(r *http.Request) bool
```

does this:

1. Reads the `Origin` header.
2. Allows requests with no origin. This helps non-browser clients and tests.
3. If `FRONTEND_URL` is set, only that exact origin is allowed.
4. If `FRONTEND_URL` is not set, localhost origins are allowed for development.

In production, set:

```env
FRONTEND_URL=https://matcha.42wolfsburg.de
```

## Creating The Hub

```go
func newWebsocketHub() *websocketHub {
	return &websocketHub{
		clients:    make(map[uint]map[*websocketClient]bool),
		register:   make(chan *websocketClient),
		unregister: make(chan *websocketClient),
		deliver:    make(chan websocketEvent),
	}
}
```

This initializes all maps and channels.

Important Go detail:

- A map must be made with `make` before use.
- A channel must be made with `make` before sending or receiving.

## Running The Hub

```go
func (hub *websocketHub) run()
```

This is an infinite loop:

```go
for {
	select {
	...
	}
}
```

`select` waits for one of several channel operations to happen.

It handles three cases.

### Register Client

```go
case client := <-hub.register:
```

This receives a new client.

If this is the first connection for the user, it creates the map:

```go
if hub.clients[client.userID] == nil {
	hub.clients[client.userID] = make(map[*websocketClient]bool)
}
```

Then it stores the client:

```go
hub.clients[client.userID][client] = true
```

### Unregister Client

```go
case client := <-hub.unregister:
```

This removes a disconnected client.

It deletes the client from the user's client map and closes the `send` channel.

Closing the `send` channel tells the write goroutine to stop.

If the user has no more clients, it removes the user from the hub.

### Deliver Event

```go
case event := <-hub.deliver:
```

This sends an event to every live connection for `event.RecipientID`.

If a client channel is blocked, the code closes and removes that client:

```go
default:
	delete(clients, client)
	close(client.send)
```

That protects the hub from getting stuck behind a broken or slow client.

## Sending To A User

```go
func (hub *websocketHub) sendToUser(event websocketEvent) {
	hub.deliver <- event
}
```

This is a small helper.

Instead of directly touching `hub.clients`, other code sends an event into the hub's `deliver` channel.

That keeps all client-map changes inside `hub.run()`, which avoids data races.

## Websocket Handler

```go
func websocketHandler(hub *websocketHub) http.HandlerFunc
```

This returns an HTTP handler.

Why return a handler instead of being a handler directly?

Because it needs access to `hub`.

Inside the handler:

```go
userID, ok := r.Context().Value(userIDKey).(uint)
```

This gets the authenticated user ID from your auth middleware.

Then:

```go
conn, err := websocketUpgrader.Upgrade(w, r, nil)
```

This upgrades HTTP to websocket.

Then it creates a client:

```go
client := &websocketClient{
	hub:    hub,
	userID: userID,
	conn:   conn,
	send:   make(chan websocketEvent, 256),
}
```

The `256` means the client can buffer up to 256 outgoing events before being considered too slow.

Then:

```go
client.hub.register <- client
```

This tells the hub to track the client.

Finally:

```go
go client.writePump()
go client.readPump()
```

Two goroutines are started:

- `readPump`: reads messages from the browser.
- `writePump`: writes messages to the browser.

You need separate goroutines because reading and writing can happen independently.

## Read Pump

```go
func (client *websocketClient) readPump()
```

This reads incoming websocket JSON from the frontend.

The defer block:

```go
defer func() {
	client.hub.unregister <- client
	client.conn.Close()
}()
```

runs when `readPump` exits. It unregisters the client and closes the websocket.

Then:

```go
client.conn.SetReadLimit(maxMessageSize)
```

prevents giant websocket messages.

Then:

```go
client.conn.SetReadDeadline(time.Now().Add(pongWait))
```

sets a timeout for reading.

Then:

```go
client.conn.SetPongHandler(...)
```

refreshes the timeout whenever the browser responds to a ping.

The main loop:

```go
for {
	var event websocketEvent
	if err := client.conn.ReadJSON(&event); err != nil {
		...
		break
	}
	...
}
```

reads JSON from the socket into a `websocketEvent`.

The backend does not trust the client for sender identity:

```go
event.SenderID = client.userID
```

This is important. A malicious client cannot pretend to be another sender.

For chat messages:

```go
message, err := saveChatMessage(client.userID, event.RecipientID, event.Message)
```

This saves the message to the database.

Then the event is enriched with database values:

```go
event.ID = message.ID
event.ConversationID = message.ConversationID
event.Message = message.Content
event.CreatedAt = message.CreatedAt
```

Then the sender receives the saved event:

```go
client.send <- event
```

And the recipient receives it live:

```go
client.hub.sendToUser(event)
```

If the recipient is offline, nothing live happens, but the message is still in the database.

## Write Pump

```go
func (client *websocketClient) writePump()
```

This writes outgoing websocket events to the browser.

It creates a ticker:

```go
ticker := time.NewTicker(pingPeriod)
```

Every tick, it sends a websocket ping.

The main loop waits for either:

- an event from `client.send`
- a ping tick

When an event arrives:

```go
client.conn.WriteJSON(event)
```

serializes the Go struct to JSON and sends it.

When the channel is closed:

```go
client.conn.WriteMessage(websocket.CloseMessage, []byte{})
```

tells the browser the websocket is closing.

## Notification Handler

```go
func notificationHandler(hub *websocketHub) http.HandlerFunc
```

This handles:

```text
POST /api/accounts/notifications/send
```

This route is staff-only.

The code checks the current user:

```go
if err := requireStaffUser(currentUserID); err != nil
```

If the user is not staff, it returns `403 Forbidden`.

Why?

In Matcha, notifications should be created by backend events, not random users. For example:

- someone liked you
- you matched
- someone viewed your profile
- you received a message

The route can still be useful for admin/testing, but regular users cannot use it.

The request body shape is:

```json
{
  "user_id": 2,
  "title": "New match",
  "message": "You matched with Alex",
  "data": {
    "profile_id": 5
  }
}
```

The handler validates:

- `user_id` exists in the payload
- `message` is not empty
- `data` is valid JSON

Then:

```go
notification, err := saveNotification(...)
```

saves it in the database.

Then:

```go
hub.sendToUser(...)
```

sends it live if the user is online.

## Realtime Helpers

File: `realtime.go`

This file contains database and HTTP helpers for chat and notifications.

## Response Structs

These structs define what JSON is returned to the frontend.

### `chatMessageResponse`

Used when returning messages.

Fields include:

- message ID
- conversation ID
- sender ID
- recipient ID
- content
- read time
- created time

### `conversationResponse`

Used when returning the conversation list.

Fields include:

- conversation ID
- both user IDs
- last message
- updated time

### `notificationResponse`

Used when returning notifications.

Fields include:

- notification ID
- user ID
- optional sender ID
- type
- title
- message
- data
- read time
- created time

## `saveChatMessage`

```go
func saveChatMessage(senderID uint, recipientID uint, content string) (ChatMessage, error)
```

This is the core chat save function.

It:

1. Trims whitespace from the message.
2. Rejects empty messages.
3. Rejects messages to yourself.
4. Checks that the recipient user exists.
5. Opens a database transaction.
6. Finds or creates the conversation.
7. Creates the chat message.
8. Updates the conversation `updated_at` timestamp.
9. Returns the saved `ChatMessage`.

The transaction matters because these operations belong together.

If message creation succeeds but conversation timestamp update fails, the app could get inconsistent. A transaction prevents partial success.

## `findOrCreateConversation`

```go
func findOrCreateConversation(db *gorm.DB, userAID uint, userBID uint) (Conversation, error)
```

This finds the conversation between two users, or creates it.

It first sorts the user IDs:

```go
userOneID, userTwoID := orderUserIDs(userAID, userBID)
```

Then it tries to create the conversation:

```go
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "user_one_id"}, {Name: "user_two_id"}},
	DoNothing: true,
}).Create(&conversation)
```

This means:

> Try to insert this row. If another row already exists with the same two user IDs, do nothing instead of crashing.

This protects against a race where two users send the first message at almost the same time.

After that, it fetches the conversation from the database.

## `orderUserIDs`

```go
func orderUserIDs(userAID uint, userBID uint) (uint, uint)
```

This returns the smaller user ID first.

Example:

```go
orderUserIDs(10, 3)
```

returns:

```go
3, 10
```

## `saveNotification`

```go
func saveNotification(userID uint, senderID *uint, title string, message string, data json.RawMessage) (Notification, error)
```

This saves a notification to the database.

It:

1. Trims the message.
2. Rejects missing user ID.
3. Rejects empty message.
4. Defaults missing `data` to `{}`.
5. Validates that `data` is real JSON.
6. Checks that the recipient user exists.
7. Creates the notification.

`senderID *uint` is a pointer because it can be `nil`.

For system notifications, use `nil`.

For staff/admin-created notifications, use the staff user's ID.

## `createSystemNotification`

```go
func createSystemNotification(hub *websocketHub, userID uint, title string, message string, data any) (Notification, error)
```

This is the function you should use later inside backend actions.

Example: when user `3` likes user `7`, you might do:

```go
_, err := createSystemNotification(hub, 7, "New like", "Someone liked your profile", map[string]any{
	"from_user_id": 3,
})
```

It:

1. Converts `data` to JSON.
2. Saves the notification.
3. Sends it live through websocket if the user is online.
4. Returns the saved notification.

This is not an HTTP endpoint. It is a backend helper.

## `requireStaffUser`

```go
func requireStaffUser(userID uint) error
```

This checks if a user has `IsStaff = true`.

It is used by the manual notification endpoint.

## `listConversations`

```go
func listConversations(w http.ResponseWriter, r *http.Request)
```

Handles:

```text
GET /api/accounts/conversations
```

It:

1. Gets current user ID from context.
2. Reads optional `limit` query parameter.
3. Finds conversations where the user is either `user_one_id` or `user_two_id`.
4. Sorts by newest `updated_at`.
5. Looks up the latest message for each conversation.
6. Returns JSON.

Example response:

```json
{
  "conversations": [
    {
      "id": 3,
      "user_one_id": 1,
      "user_two_id": 2,
      "last_message": {
        "id": 15,
        "conversation_id": 3,
        "sender_id": 1,
        "recipient_id": 2,
        "content": "Hey!",
        "created_at": "2026-05-02T15:00:00Z"
      },
      "updated_at": "2026-05-02T15:00:00Z"
    }
  ]
}
```

## `listConversationMessages`

```go
func listConversationMessages(w http.ResponseWriter, r *http.Request)
```

Handles:

```text
GET /api/accounts/conversations/:conversationID/messages
```

It:

1. Gets current user ID.
2. Parses `conversationID` from the URL.
3. Verifies the user belongs to the conversation.
4. Loads the latest messages.
5. Reverses them so the frontend receives oldest to newest.
6. Returns JSON.

This ownership check is important:

```go
WHERE id = ? AND (user_one_id = ? OR user_two_id = ?)
```

It prevents users from reading conversations they do not belong to.

## `markMessageRead`

```go
func markMessageRead(w http.ResponseWriter, r *http.Request)
```

Handles:

```text
PATCH /api/accounts/messages/:messageID/read
```

It:

1. Gets current user ID.
2. Parses `messageID`.
3. Finds a message where `recipient_id` is the current user.
4. Sets `read_at` to now.
5. Returns the updated message.

Only the recipient can mark a message as read.

## `listNotifications`

```go
func listNotifications(w http.ResponseWriter, r *http.Request)
```

Handles:

```text
GET /api/accounts/notifications
GET /api/accounts/notifications?unread=true
```

It:

1. Gets current user ID.
2. Reads optional `limit`.
3. If `unread=true`, filters to unread notifications.
4. Returns newest notifications first.

## `markNotificationRead`

```go
func markNotificationRead(w http.ResponseWriter, r *http.Request)
```

Handles:

```text
PATCH /api/accounts/notifications/:notificationID/read
```

It:

1. Gets current user ID.
2. Parses `notificationID`.
3. Finds a notification owned by the current user.
4. Sets `read_at` to now.
5. Returns the updated notification.

## Builder Helpers

### `buildChatMessageResponse`

Converts a `ChatMessage` database model into JSON response shape.

This keeps API responses separate from database models.

### `buildNotificationResponse`

Converts a `Notification` model into JSON response shape.

It also parses `notification.Data` from a string into JSON.

## Parser Helpers

### `parseRealtimeLimit`

Reads a string like `"20"` and returns an integer.

If invalid or missing, it returns `50`.

If too big, it caps at `100`.

This prevents users from asking for too much data at once.

### `parseUintRouteParam`

Reads a route parameter like `:conversationID` or `:messageID`.

It returns:

```go
(uint, true)
```

if valid, or:

```go
(0, false)
```

if invalid.

### `writeJSON`

Sets:

```text
Content-Type: application/json
```

Then writes the HTTP status code and JSON body.

## Frontend Usage

### Connect Websocket

```js
const socket = new WebSocket("ws://localhost:8000/api/accounts/ws");
```

In production use `wss://`.

Because authentication uses cookies, the browser sends the `access_token` cookie automatically when the domain matches.

### Send Chat Message

```js
socket.send(JSON.stringify({
  type: "chat_message",
  recipient_id: 2,
  message: "Hey!"
}));
```

### Receive Events

```js
socket.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === "chat_message") {
    // Add message to chat UI
  }

  if (data.type === "notification") {
    // Add notification to notification UI
  }
};
```

### Load Conversations

```http
GET /api/accounts/conversations
```

### Load Messages

```http
GET /api/accounts/conversations/3/messages
```

### Mark Message Read

```http
PATCH /api/accounts/messages/15/read
```

### Load Notifications

```http
GET /api/accounts/notifications
```

Only unread:

```http
GET /api/accounts/notifications?unread=true
```

### Mark Notification Read

```http
PATCH /api/accounts/notifications/8/read
```

## How To Create Matcha Notifications Later

When you add a like or match endpoint, call:

```go
_, err := createSystemNotification(hub, targetUserID, "New like", "Someone liked your profile", map[string]any{
	"from_user_id": currentUserID,
})
```

For a match:

```go
_, err := createSystemNotification(hub, targetUserID, "It's a match", "You have a new match", map[string]any{
	"matched_user_id": currentUserID,
})
```

Important: to call `createSystemNotification`, the handler needs access to `hub`. If your future `likeUser` handler needs notifications, define it like:

```go
func likeUser(hub *websocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// like logic
		createSystemNotification(hub, targetUserID, "New like", "Someone liked your profile", nil)
	}
}
```

Then register it like:

```go
privateMux.HandleFunc(pat.Post("/likes/:userID"), likeUser(hub))
```

## Important Design Choices

### Why Save Before Sending?

The backend saves chat messages before websocket delivery.

This means:

- if recipient is online, they get it live
- if recipient is offline, they can fetch it later
- refresh does not lose messages

### Why One Hub?

One hub centralizes websocket client state.

Without the hub, every handler might try to read/write the client map directly, causing data races.

### Why Channels?

Channels let goroutines communicate safely.

The hub owns the `clients` map. Other goroutines send requests into the hub through channels.

### Why Transactions?

Saving a chat message touches multiple things:

- conversation
- message
- conversation timestamp

A transaction makes those changes succeed or fail together.

### Why Staff-Only Notification Endpoint?

Matcha notifications should come from app events, not arbitrary users.

Normal users should not be able to send:

```text
"You have a new match"
```

to anyone they want.

The public-looking route is now limited to staff. Real product notifications should use `createSystemNotification`.

## Current Limitations

- Conversation list uses one extra query per conversation to fetch the last message. This is okay for now, but later you can optimize it with `LastMessageID` on `Conversation`.
- Websocket delivery is in-memory. If you run multiple backend instances, a user connected to instance A will not receive events sent by instance B unless you add Redis Pub/Sub or another broker.
- There is no message deletion endpoint.
- There is no typing indicator yet.
- There is no presence tracking yet.

## Quick Mental Model

Think of it like this:

```text
websocket.go = live wire
realtime.go  = database and REST API
models.go    = database shape
main.go      = route wiring
db.go        = migration setup
```

The live wire sends events now.

The database keeps them for later.
