package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"go.uber.org/zap"

	"kudytudy-api/internal/database"
)

// WSHub manages WebSocket client connections and bridges them to Redis PubSub.
type WSHub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // channel -> set of connections

	redisClient *database.RedisClient
	logger      *zap.Logger
}

// NewWSHub creates a new WebSocket hub.
func NewWSHub(redisClient *database.RedisClient, logger *zap.Logger) *WSHub {
	hub := &WSHub{
		clients:     make(map[string]map[*websocket.Conn]bool),
		redisClient: redisClient,
		logger:      logger.Named("ws_hub"),
	}

	// Subscribe to Redis PubSub channels for cross-instance broadcast.
	if redisClient != nil {
		go hub.subscribeRedis()
	}

	return hub
}

// RegisterFastHTTP adds a fasthttp websocket connection to a channel.
func (h *WSHub) RegisterFastHTTP(channel string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[channel] == nil {
		h.clients[channel] = make(map[*websocket.Conn]bool)
	}
	h.clients[channel][conn] = true

	h.logger.Debug("ws client registered",
		zap.String("channel", channel),
		zap.Int("total", len(h.clients[channel])),
	)
}

// UnregisterFastHTTP removes a fasthttp websocket connection from a channel.
func (h *WSHub) UnregisterFastHTTP(channel string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[channel]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, channel)
		}
	}

	_ = conn.Close()
}

// Broadcast sends a message to all local clients on a channel.
func (h *WSHub) Broadcast(channel string, msg []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0)
	for conn := range h.clients[channel] {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			h.logger.Debug("ws write error, removing client", zap.Error(err))
			h.UnregisterFastHTTP(channel, conn)
		}
	}
}

// Publish sends a message to all instances via Redis PubSub (or local-only).
func (h *WSHub) Publish(channel string, msg []byte) {
	if h.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.redisClient.Client.Publish(ctx, "ws:"+channel, msg).Err(); err != nil {
			h.logger.Warn("redis publish error", zap.String("channel", channel), zap.Error(err))
		}
	} else {
		h.Broadcast(channel, msg)
	}
}

// subscribeRedis listens to Redis PubSub and broadcasts to local clients.
func (h *WSHub) subscribeRedis() {
	ctx := context.Background()
	pubsub := h.redisClient.Client.PSubscribe(ctx, "ws:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		// Strip "ws:" prefix to get the channel name.
		channel := msg.Channel[3:]
		h.Broadcast(channel, []byte(msg.Payload))
	}
}

// BroadcastJSON serializes v to JSON and publishes to the channel.
func (h *WSHub) BroadcastJSON(channel string, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		h.logger.Warn("json marshal error for ws broadcast", zap.Error(err))
		return
	}
	h.Publish(channel, data)
}

// ActiveConnections returns the total number of active connections across all channels.
func (h *WSHub) ActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, conns := range h.clients {
		total += len(conns)
	}
	return total
}

// Well-known WS channel constants.
const (
	ChannelWeather       = "weather"
	ChannelNotifications = "notifications"
)

// WeatherMsg wraps weather data for WS broadcast.
type WeatherMsg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	TS   int64       `json:"ts"`
}
