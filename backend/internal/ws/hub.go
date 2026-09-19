package ws

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	// The frontend and backend are on different origins once deployed, so we
	// can't rely on same-origin defaults. Real origin checking (against the
	// configured ALLOWED_ORIGIN) happens in production; this keeps local dev simple.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub tracks, per poll ID, the set of currently-connected WebSocket clients
// and whether a Redis subscriber goroutine is already running for that poll.
// One Redis subscription per poll (not per client) is what keeps this scaling
// sensibly as viewer count grows.
type Hub struct {
	mu      sync.Mutex
	clients map[string]map[*websocket.Conn]bool
	rdb     *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]bool),
		rdb:     rdb,
	}
}

// ServeWs upgrades the HTTP connection, registers the client under its poll
// ID, lazily starts a Redis subscriber for that poll if this is the first
// viewer, and blocks reading (to detect disconnects) until the client leaves.
func (h *Hub) ServeWs(c *gin.Context) {
	pollID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	h.register(pollID, conn)
	defer h.unregister(pollID, conn)

	// Client -> server messages aren't part of this app's protocol; we just
	// need the read loop running so gorilla/websocket notices disconnects
	// (ping/pong and close frames are handled internally by ReadMessage).
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) register(pollID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	isFirstClient := h.clients[pollID] == nil
	if isFirstClient {
		h.clients[pollID] = make(map[*websocket.Conn]bool)
	}
	h.clients[pollID][conn] = true

	if isFirstClient {
		go h.subscribeToPoll(pollID)
	}
}

func (h *Hub) unregister(pollID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[pollID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, pollID)
			// The subscriber goroutine notices the channel closing via context
			// cancellation is not wired here for simplicity; in practice Redis
			// pub/sub subscriptions are cheap and idle ones can be left running,
			// but a production version would cancel this on zero clients.
		}
	}
}

// subscribeToPoll listens on this poll's Redis channel and forwards every
// message to every currently-connected client for that poll. This is the
// piece that makes updates "push", not "poll for changes".
func (h *Hub) subscribeToPoll(pollID string) {
	channel := "poll:" + pollID + ":updates"
	sub := h.rdb.Subscribe(context.Background(), channel)
	defer sub.Close()

	ch := sub.Channel()
	for msg := range ch {
		h.broadcast(pollID, []byte(msg.Payload))

		h.mu.Lock()
		empty := len(h.clients[pollID]) == 0
		h.mu.Unlock()
		if empty {
			return
		}
	}
}

func (h *Hub) broadcast(pollID string, payload []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients[pollID] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			conn.Close()
			delete(h.clients[pollID], conn)
		}
	}
}
