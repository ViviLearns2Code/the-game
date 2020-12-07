package main

import (
	//"encoding/json"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var rootTemplate = template.Must(template.New("root").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<script>
	websocket = new WebSocket("ws://{{.}}/socket");
	var onMessage = function(m){
		var node = document.createElement("p");
		var textnode = document.createTextNode(m.data);
		node.appendChild(textnode);
		document.getElementById("chat").appendChild(node);
	}
	var onClose = function(m){
		var node = document.createElement("p");
		var textnode = document.createTextNode("Connection closed: "+m.reason);
		node.appendChild(textnode);
		document.getElementById("chat").appendChild(node);
	}
	var onSend = function(e){
		websocket.send(JSON.stringify({
			"playerName": document.getElementById("player-name").value,
			"actionId": document.getElementById("action-id").value
		}));
	}
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;
</script>
Player Name <input id="player-name"/>
Action Id <input id="action-id">
<button onclick="onSend(this)">Send</button>
<div id="chat"></div>
</html>
`))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("main", goid())
	rootTemplate.Execute(w, listenAddr)
}

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// Global map, protected by lock
var mutex = &sync.RWMutex{}
var gameMap = make(map[string]Game)

func getGameFromMap(gameId string) (Game, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	var val, ok = gameMap[gameId]
	return val, ok
}

func addGameToMap(gameId string, game Game) {
	mutex.Lock()
	defer mutex.Unlock()
	gameMap[gameId] = game
}

type gameDetails struct {
	gameId     string
	playerId   string
	playerName string
	actionId   string
	raw        map[string]string
}

type GameStateOutput struct {
	TopCard int   `json:"topCard, omitempty"`
	Level   int   `json:"level, omitempty"`
	Lives   int   `json:"lives, omitempty"`
	Stars   int   `json:"stars, omitempty"`
	Hand    []int `json:"hand, omitempty"`
}

type GameOutput struct {
	GameId     string           `json:"gameId"`
	PlayerId   string           `json:"playerId"`
	PlayerName string           `json:"playerName"`
	GameState  *GameStateOutput `json:"gameState, omitempty"`
}

func convertGameStateToOutput(gameState GameState, playerId string, playerName string) GameOutput {
	var gameStateOutput = GameStateOutput{
		TopCard: 12,
		Level:   2,
		Lives:   3,
		Stars:   3,
		Hand:    []int{20, 45, 88},
	}
	var gameOutput = GameOutput{
		GameId:     gameState.gameId,
		PlayerId:   playerId,
		PlayerName: playerName,
		GameState:  &gameStateOutput,
	}
	return gameOutput
}
func isValidAction(actionId string) bool {
	actions := [8]string{"create", "join", "concentrate", "ready", "propose-star", "agree-star", "reject-star", "card"}
	for _, a := range actions {
		if a == actionId {
			return true
		}
	}
	return false
}

const listenAddr = "192.168.178.23:4000" //localhost:4000

func main() {
	log.Printf("hello server")
	fmt.Println("main", goid())
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", http.HandlerFunc(runGame))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("hello handler")
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"http://0.0.0.0"},
	})
	if err != nil {
		log.Printf("error in accept")
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour*120000)
	defer cancel()

	for {
		fmt.Println("main", goid())
		var v = make(map[string]string)
		log.Printf("hello for")
		err = wsjson.Read(ctx, c, &v)
		log.Printf("received: %v", v)
		if err != nil {
			log.Printf("error in read")
			break
		}
		err = wsjson.Write(ctx, c, struct{ Message string }{
			"hello, browser",
		})
		if err != nil {
			log.Printf("error in write")
			break
		}
	}
	if websocket.CloseStatus(err) == websocket.StatusGoingAway {
		err = nil
	}
	log.Printf("bye handler")
	c.Close(websocket.StatusNormalClosure, "")
	return
}
func extractDetails(raw map[string]string) gameDetails {
	var gameId, _ = raw["gameId"]
	var playerId, _ = raw["playerId"]
	var playerName, _ = raw["playerName"]
	var actionId, _ = raw["actionId"]
	var details = gameDetails{
		gameId:     gameId,
		playerId:   playerId,
		playerName: playerName,
		actionId:   actionId,
		raw:        raw,
	}
	return details
}
func runGame(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection established...")
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: false,
		OriginPatterns:     []string{"0.0.0.0:8000"},
	})
	if err != nil {
		log.Printf("error in accept %e", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour*120000)
	defer cancel()

	var myPlayerId string
	var myPlayerName string
	var myGame Game
	var myPlayerChannel chan GameState
	var failed = false
	log.Printf("Entering loop...")
	for {
		var data = make(map[string]string)
		var err = wsjson.Read(ctx, c, &data)
		if err != nil {
			log.Printf("Error reading json %e", err)
		}
		gameDetails := extractDetails(data)
		log.Printf("Checking inputs")
		if gameDetails.playerName == "" {
			c.Close(websocket.StatusUnsupportedData, "playerName corrupted")
			break
		}
		if gameDetails.playerId != myPlayerId {
			c.Close(websocket.StatusUnsupportedData, "playerId corrupted")
			break
		}
		if !isValidAction(gameDetails.actionId) {
			c.Close(websocket.StatusUnsupportedData, "actionId corrupted")
			break
		}
		if gameDetails.gameId != myGame.id {
			c.Close(websocket.StatusUnsupportedData, "gameId corrupted")
			break
		}
		switch gameDetails.actionId {
		case "create":
			log.Printf("Creating game...")
			myPlayerName = gameDetails.playerName
			myGame = *NewGame()
			addGameToMap(myGame.id, myGame)
			go myGame.Start()
			myPlayerId, myPlayerChannel, err = myGame.Subscribe(myPlayerName)
			if err != nil {
				c.Close(websocket.StatusUnsupportedData, err.Error())
				failed = true
				break
			}
			go listenPlayerChannel(c, ctx, myPlayerId, myPlayerName, myPlayerChannel)
		case "join":
			log.Printf("Joining game...")
			_myGame, ok := getGameFromMap(gameDetails.gameId)
			myGame = _myGame
			if !ok {
				c.Close(websocket.StatusUnsupportedData, "GameID corrupted")
				failed = true
				break
			}
			myPlayerName = gameDetails.playerName
			myPlayerId, myPlayerChannel, err = myGame.Subscribe(myPlayerName)
			if err != nil {
				c.Close(websocket.StatusUnsupportedData, err.Error())
				failed = true
				break
			}
			go listenPlayerChannel(c, ctx, myPlayerId, myPlayerName, myPlayerChannel)
		default:
			if myPlayerName != gameDetails.playerName {
				c.Close(websocket.StatusUnsupportedData, "PlayerName corrupted")
				failed = true
				break
			}
		}
		if failed {
			break
		}
		log.Printf("playerId: %v", myPlayerId)
		log.Printf("playerName: %v", myPlayerName)
		log.Printf("gameId: %v", myGame.id)
		log.Printf("actionId: %v", gameDetails.actionId)
		log.Printf("Passing inputs to game core...")
		myGame.inputCh <- data
	}
}
func listenPlayerChannel(c *websocket.Conn, ctx context.Context, playerId string, playerName string, myPlayerChannel chan GameState) {
	var err error
	log.Printf("Player channel opened...")
	for {
		gameState := <-myPlayerChannel
		log.Printf("New game state received")
		output := convertGameStateToOutput(gameState, playerId, playerName)
		err = wsjson.Write(ctx, c, output)
		if err != nil {
			log.Printf("Error in write")
			break
		}
	}
	if websocket.CloseStatus(err) == websocket.StatusGoingAway {
		err = nil
	}
	c.Close(websocket.StatusNormalClosure, "")
}
