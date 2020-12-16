package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func TestServer(t *testing.T) {
	var wsAddress string
	host, ok := os.LookupEnv("GAMEHOST")
	if !ok {
		host = getLocalIP()
	}
	wsAddress = fmt.Sprintf("ws://%s:443/socket", host)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsAddress, nil)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// correct input
	var payloadRequest = &InputDetails{
		PlayerName: "mary",
		ActionId:   "create",
	}
	err = wsjson.Write(ctx, c, payloadRequest)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var payloadResponse = &GameOutput{}
	err = wsjson.Read(ctx, c, &payloadResponse)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	// wrong input
	payloadRequest = &InputDetails{
		PlayerName: "mary",
		ActionId:   "",
	}
	err = wsjson.Write(ctx, c, payloadRequest)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	payloadResponse = &GameOutput{}
	err = wsjson.Read(ctx, c, &payloadResponse)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	assert.NotEqual(t, payloadResponse.ErrorMsg, "")
	c.Close(websocket.StatusNormalClosure, "")
}
