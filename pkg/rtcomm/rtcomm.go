package rtcomm

import (
	"github.com/google/uuid"
	"sync"
)

// TODO associate lock with individual clients to prevent global locking
type Clients struct {
	sync.RWMutex
	clients map[uuid.UUID][]chan StateUpdate
}

func NewClients() *Clients {
	return &Clients{
		clients: make(map[uuid.UUID][]chan StateUpdate),
	}
}

func (c *Clients) AddClient(id uuid.UUID, client chan StateUpdate) {
	c.Lock()
	defer c.Unlock()
	c.clients[id] = append(c.clients[id], client)
}

//TODO remove debug
func (c *Clients) Count() (sessions, clients uint) {
	c.RLock()
	defer c.RUnlock()
	for _, s := range c.clients {
		sessions++
		clients += uint(len(s))
	}
	return
}

// TODO optimize
func (c *Clients) RemoveClient(id uuid.UUID, client chan StateUpdate) {
	c.Lock()
	defer c.Unlock()
	var newClients = make([]chan StateUpdate, 0, len(c.clients[id]) - 1)
	for i := 0; i < len(c.clients[id]); i++ {
		if c.clients[id][i] != client {
			newClients = append(newClients, c.clients[id][i])
		}
	}
	c.clients[id] = newClients
}

func (c *Clients) SendToAll(id uuid.UUID, su StateUpdate) (sent, dropped uint) {
	c.RLock()
	defer c.RUnlock()
	for i := 0; i < len(c.clients[id]); i++ {
		select {
		case c.clients[id][i] <- su:
			sent++
		default:
			dropped++
		}
	}
	return
}

