// Package ws, ilan bazlı WebSocket odalarını yönetir ve olayları bağlı istemcilere iletir.
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     sameOrigin,
}

// sameOrigin, yalnızca aynı siteden açılan WebSocket bağlantılarını kabul eder
// (siteler arası WebSocket ele geçirme / CSWSH saldırılarını önler).
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // Tarayıcı dışı istemciler Origin göndermez
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Host == r.Host
}

// Client, aktif bir WebSocket istemci bağlantısını temsil eder.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	listingID int64
}

type broadcastMessage struct {
	listingID int64
	payload   []byte
}

// Hub, ilan odalarını ve bağlı istemcileri yönetir.
// Oda haritasına yalnızca Run goroutine'i erişir; bu yüzden kilit gerekmez.
type Hub struct {
	rooms      map[int64]map[*Client]struct{}
	broadcast  chan broadcastMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]struct{}),
		broadcast:  make(chan broadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run, hub döngüsünü çalıştırır; ctx iptal edilince tüm bağlantıları kapatıp döner.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for _, clients := range h.rooms {
				for c := range clients {
					close(c.send)
				}
			}
			h.rooms = make(map[int64]map[*Client]struct{})
			return

		case c := <-h.register:
			if h.rooms[c.listingID] == nil {
				h.rooms[c.listingID] = make(map[*Client]struct{})
			}
			h.rooms[c.listingID][c] = struct{}{}

		case c := <-h.unregister:
			h.remove(c)

		case msg := <-h.broadcast:
			for c := range h.rooms[msg.listingID] {
				select {
				case c.send <- msg.payload:
				default:
					h.remove(c) // Mesajları tüketemeyen yavaş istemciyi düşür
				}
			}
		}
	}
}

func (h *Hub) remove(c *Client) {
	clients, ok := h.rooms[c.listingID]
	if !ok {
		return
	}
	if _, exists := clients[c]; !exists {
		return
	}
	delete(clients, c)
	close(c.send)
	if len(clients) == 0 {
		delete(h.rooms, c.listingID)
	}
}

// Broadcast, belirli bir ilanın odasındaki tüm istemcilere veri gönderir.
// Çağıranı asla bloklamaz; kuyruk doluysa mesaj düşürülür (istemciler ayrıca yedek polling yapar).
func (h *Hub) Broadcast(listingID int64, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("WS JSON hatası: %v", err)
		return
	}
	select {
	case h.broadcast <- broadcastMessage{listingID: listingID, payload: payload}:
	default:
		log.Printf("⚠️  WS yayın kuyruğu dolu, ilan %d mesajı düşürüldü", listingID)
	}
}

// ServeWS, HTTP isteğini WebSocket protokolüne yükseltir ve istemciyi ilgili ilanın odasına kaydeder.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, listingID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade hatası: %v", err)
		return
	}

	c := &Client{hub: h, conn: conn, send: make(chan []byte, 16), listingID: listingID}
	h.register <- c

	go c.writePump()
	go c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// Her mesaj ayrı bir frame olarak gönderilir (istemci her frame'i tek JSON olarak ayrıştırır)
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
