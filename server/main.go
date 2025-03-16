package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	outboxes map[string]*websocket.Conn
}

func (h *Hub) AddConnection(conn *websocket.Conn) {
	slog.Info("adding connection to outboxes", "remote_addr", conn.RemoteAddr().String())
	h.outboxes[conn.RemoteAddr().String()] = conn
}

func (h *Hub) RemoveConnection(conn *websocket.Conn) {
	slog.Info("removing connection from outboxes", "remote_addr", conn.RemoteAddr().String())
	delete(h.outboxes, conn.RemoteAddr().String())
}

func (h *Hub) Broadcast(msg []byte) {
	if len(msg) == 0 {
		return
	}

	if len(h.outboxes) == 0 {
		slog.Info("no outboxes to broadcast to")
		return
	}

	for _, conn := range h.outboxes {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			slog.Error("error writing message to outbox: ", "error", err)
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	h := &Hub{
		outboxes: make(map[string]*websocket.Conn, 0),
	}

	t := time.NewTicker(5 * time.Second)
	stopTicker := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopTicker:
				return
			case <-t.C:
				slog.Info("tick tock")
				h.Broadcast([]byte(time.Now().Format(time.RFC3339)))
			}
		}
	}()

	http.HandleFunc("/echo", echo(h))

	s := http.Server{Addr: *addr}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			slog.Error("error starting server", "error", err)
		}
	}()

	<-sigChan
	slog.Info("shutting down")
	if err := s.Shutdown(context.Background()); err != nil {
		slog.Error("error shutting down server", "error", err)
	}
	t.Stop()
	close(stopTicker)
}

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(h *Hub) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer func() {
			h.RemoveConnection(c)
			c.Close()
		}()

		h.AddConnection(c)

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
	return fn
}
