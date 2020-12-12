package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func TestServer(t *testing.T) {
	var wsAddress string
	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = getLocalIP()
	}
	if !ok {
		wsAddress = fmt.Sprintf("ws://%s:4000/socket", host)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsAddress, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	var payload = &InputDetails{
		PlayerName: "mary",
		ActionId:   "create",
	}
	err = wsjson.Write(ctx, c, payload)
	if err != nil {
		t.Error(err.Error())
		return
	}

	c.Close(websocket.StatusNormalClosure, "")
}
