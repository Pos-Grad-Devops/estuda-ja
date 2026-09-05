package chat

import "sync"

const sendBuffer = 32

// Client é uma conexão WebSocket registrada numa sala (aula_id).
type Client struct {
	AulaID uint
	UserID uint
	Nome   string
	Send   chan []byte

	writeMu sync.Mutex
}

// Write serializa writes concorrentes na conexão (broadcast + ping/pong + erros).
func (c *Client) Write(fn func() error) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return fn()
}

type Room struct {
	clients map[*Client]struct{}
}

type envelope struct {
	aulaID  uint
	payload []byte
}

type clientOp struct {
	client *Client
	done   chan struct{}
}

// Hub mantém salas em memória por aula_id. Reinício do processo zera o mapa.
type Hub struct {
	mu         sync.RWMutex
	rooms      map[uint]*Room
	register   chan clientOp
	unregister chan clientOp
	broadcast  chan envelope
}

func NewHub() *Hub {
	h := &Hub{
		rooms:      make(map[uint]*Room),
		register:   make(chan clientOp),
		unregister: make(chan clientOp),
		broadcast:  make(chan envelope, 64),
	}
	go h.Run()
	return h
}

func NewClient(aulaID, userID uint, nome string) *Client {
	return &Client{
		AulaID: aulaID,
		UserID: userID,
		Nome:   nome,
		Send:   make(chan []byte, sendBuffer),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case op := <-h.register:
			c := op.client
			h.mu.Lock()
			room := h.rooms[c.AulaID]
			if room == nil {
				room = &Room{clients: make(map[*Client]struct{})}
				h.rooms[c.AulaID] = room
			}
			room.clients[c] = struct{}{}
			h.mu.Unlock()
			close(op.done)

		case op := <-h.unregister:
			c := op.client
			h.mu.Lock()
			if room, ok := h.rooms[c.AulaID]; ok {
				if _, exists := room.clients[c]; exists {
					delete(room.clients, c)
					close(c.Send)
				}
				if len(room.clients) == 0 {
					delete(h.rooms, c.AulaID)
				}
			}
			h.mu.Unlock()
			close(op.done)

		case msg := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[msg.aulaID]
			if room != nil {
				for client := range room.clients {
					select {
					case client.Send <- msg.payload:
					default:
						// cliente lento: descarta para não bloquear o hub
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(c *Client) {
	done := make(chan struct{})
	h.register <- clientOp{client: c, done: done}
	<-done
}

func (h *Hub) Unregister(c *Client) {
	done := make(chan struct{})
	h.unregister <- clientOp{client: c, done: done}
	<-done
}

func (h *Hub) Broadcast(aulaID uint, payload []byte) {
	h.broadcast <- envelope{aulaID: aulaID, payload: payload}
}

func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

func (h *Hub) ClientCount(aulaID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room := h.rooms[aulaID]
	if room == nil {
		return 0
	}
	return len(room.clients)
}
