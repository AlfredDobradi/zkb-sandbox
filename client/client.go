package main

import (
	"flag"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"time"
	"ws-sandbox/pkg/model"

	"github.com/gorilla/websocket"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "zkillboard.com", Path: "/websocket/"}
	slog.Info("connecting to websocket", "url", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	if err := c.WriteJSON(map[string]string{
		"action":  "sub",
		"channel": "killstream",
	}); err != nil {
		slog.Warn("failed to subscribe to killstream", "error", err)
		os.Exit(1)
	}

	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	go func() {
		for range t.C {
			slog.Info("sending ping")
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Warn("failed to send ping", "error", err)
				continue
			}
		}
	}()

	// retries := 0
	// retry := 2

	var killmail model.Killmail

	go func() {
		defer close(done)
		for {
			err := c.ReadJSON(&killmail)
			if err != nil {
				slog.Warn("error reading message", "error", err)
				continue
			}
			// for err != nil && retries < 5 {
			// 	if c != nil {
			// 		c.Close()
			// 	}

			// 	cc, _, connErr := websocket.DefaultDialer.Dial(u.String(), nil)
			// 	if connErr != nil {
			// 		retryInterval := time.Duration(retry) * time.Second
			// 		slog.Error("error reading message",
			// 			"error", connErr,
			// 			"retries", retries,
			// 			"sleeping", retryInterval.String(),
			// 		)
			// 		time.Sleep(retryInterval)
			// 		retries++
			// 		retry = retry << 1
			// 		continue
			// 	}

			// 	c = cc
			// 	err = nil
			// }
			// if err != nil {
			// 	slog.Error("couldn't reconnect")
			// 	return
			// }
			// retries = 0
			if killmail.Zkill.NPC {
				continue
			}

			slog.Info("received new killmail",
				"original_timestamp", killmail.OriginalTimestamp,
				"id", killmail.KillmailID,
				"hash", killmail.Zkill.Hash,
				"url", killmail.Zkill.URL,
			)
		}
	}()

	<-interrupt
	slog.Info("interrupt received, closing connection")
}
