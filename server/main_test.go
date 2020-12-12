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
	host, ok := os.LookupEnv("GAMEHOST")
	if !ok {
		host = getLocalIP()
	}
	wsAddress = fmt.Sprintf("ws://%s:4000/socket", host)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsAddress, nil)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// correct input
	var payload_request = &InputDetails{
		PlayerName: "mary",
		ActionId:   "create",
	}
	err = wsjson.Write(ctx, c, payload_request)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	var payload_response = &GameOutput{}
	err = wsjson.Read(ctx, c, &payload_response)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	// wrong input
	payload_request = &InputDetails{
		PlayerName: "mary",
		ActionId:   "",
	}
	err = wsjson.Write(ctx, c, payload_request)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	payload_response = &GameOutput{}
	err = wsjson.Read(ctx, c, &payload_response)
	if err == nil {
		t.Fatal(err.Error())
		return
	}
	c.Close(websocket.StatusNormalClosure, "")
}
