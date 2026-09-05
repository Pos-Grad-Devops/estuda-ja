package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func waitPayload(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(time.Second):
		t.Fatal("timeout esperando broadcast")
		return nil
	}
}

func TestHubIsolatesAulas(t *testing.T) {
	hub := NewHub()
	a := NewClient(1, 10, "Ana")
	b := NewClient(2, 20, "Bruno")
	hub.Register(a)
	hub.Register(b)
	require.Equal(t, 1, hub.ClientCount(1))
	require.Equal(t, 1, hub.ClientCount(2))

	hub.Broadcast(1, []byte("msg-aula-1"))
	got := waitPayload(t, a.Send)
	require.Equal(t, "msg-aula-1", string(got))

	select {
	case msg := <-b.Send:
		t.Fatalf("aula B recebeu mensagem da aula A: %s", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHubBroadcastIncludesSender(t *testing.T) {
	hub := NewHub()
	sender := NewClient(7, 1, "Maria")
	peer := NewClient(7, 2, "João")
	hub.Register(sender)
	hub.Register(peer)

	hub.Broadcast(7, []byte("eco"))
	require.Equal(t, "eco", string(waitPayload(t, sender.Send)))
	require.Equal(t, "eco", string(waitPayload(t, peer.Send)))
}

func TestHubRegisterUnregisterAndEmptyRoomGC(t *testing.T) {
	hub := NewHub()
	c1 := NewClient(3, 1, "A")
	c2 := NewClient(3, 2, "B")
	hub.Register(c1)
	hub.Register(c2)
	require.Equal(t, 1, hub.RoomCount())
	require.Equal(t, 2, hub.ClientCount(3))

	hub.Unregister(c1)
	require.Equal(t, 1, hub.ClientCount(3))
	select {
	case _, ok := <-c1.Send:
		require.False(t, ok, "Send do cliente removido deve ser fechado")
	case <-time.After(time.Second):
		t.Fatal("timeout esperando close do Send")
	}

	hub.Broadcast(3, []byte("ainda-na-sala"))
	require.Equal(t, "ainda-na-sala", string(waitPayload(t, c2.Send)))

	hub.Unregister(c2)
	require.Equal(t, 0, hub.ClientCount(3))
	require.Equal(t, 0, hub.RoomCount())
}

func TestHubRestartClearsRooms(t *testing.T) {
	hub := NewHub()
	hub.Register(NewClient(9, 1, "X"))
	require.Equal(t, 1, hub.RoomCount())

	hub2 := NewHub()
	require.Equal(t, 0, hub2.RoomCount())
	require.Equal(t, 0, hub2.ClientCount(9))
}
