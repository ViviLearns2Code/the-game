package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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
			"actionId": document.getElementById("action-id").value,
			"playerId": document.getElementById("player-id").value,
			"cardId": document.getElementById("card-id").value,
			"gameId": document.getElementById("game-id").value
		}));
	}
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;
</script>
Player Name <input id="player-name"/></br>
Action Id <input id="action-id"/></br>
Player Id <input id="player-id"/></br>
Game Id <input id="game-id"/></br>
Card Id <input id="card-id"/></br>
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

/*
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
*/
func extractDetails(raw map[string]interface{}) inputDetails {
	var gameId, _ = raw["gameId"].(string)
	var playerId, _ = raw["playerId"].(string)
	var playerName, _ = raw["playerName"].(string)
	var actionId, _ = raw["actionId"].(string)
	var cardId, _ = raw["card"].(int)
	var details = inputDetails{
		gameId:     gameId,
		playerId:   playerId,
		playerName: playerName,
		actionId:   actionId,
		cardId:     cardId,
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
	log.Printf("Entering loop...")
	for {
		var data = make(map[string]interface{})
		var err = wsjson.Read(ctx, c, &data)
		if err != nil {
			log.Printf("Error reading json %e", err)
		}
		gameDetails := extractDetails(data)
		log.Printf("Checking inputs")
		if gameDetails.playerName == "" {
			c.Close(websocket.StatusUnsupportedData, "playerName corrupted")
			return
		}
		if !isValidAction(gameDetails.actionId) {
			c.Close(websocket.StatusUnsupportedData, "actionId corrupted")
			return
		}
		switch gameDetails.actionId {
		case "create":
			log.Printf("Creating game...")
			myPlayerName = gameDetails.playerName
			myGame = *NewGame()
			addGameToMap(myGame.id, myGame)
			go myGame.Start()
			myPlayerId, myPlayerChannel = myGame.Subscribe(myPlayerName)
			go listenPlayerChannel(c, ctx, myPlayerId, myPlayerName, myPlayerChannel)
		case "join":
			log.Printf("Joining game...")
			_myGame, ok := getGameFromMap(gameDetails.gameId)
			myGame = _myGame
			if !ok {
				c.Close(websocket.StatusUnsupportedData, "GameID corrupted")
				return
			}
			myPlayerName = gameDetails.playerName
			myPlayerId, myPlayerChannel = myGame.Subscribe(myPlayerName)
			go listenPlayerChannel(c, ctx, myPlayerId, myPlayerName, myPlayerChannel)
		default:
			if gameDetails.playerId != myPlayerId {
				c.Close(websocket.StatusUnsupportedData, "playerId corrupted")
				return
			}
			if gameDetails.gameId != myGame.id {
				c.Close(websocket.StatusUnsupportedData, "gameId corrupted")
				return
			}
			if myPlayerName != gameDetails.playerName {
				c.Close(websocket.StatusUnsupportedData, "PlayerName corrupted")
				return
			}
		}
		log.Printf("playerId: %v", myPlayerId)
		log.Printf("playerName: %v", myPlayerName)
		log.Printf("gameId: %v", myGame.id)
		log.Printf("actionId: %v", gameDetails.actionId)
		log.Printf("Passing inputs to game core...")
		myGame.inputCh <- gameDetails
	}
}
func listenPlayerChannel(c *websocket.Conn, ctx context.Context, playerId string, playerName string, myPlayerChannel chan GameState) {
	var err error
	log.Printf("Player channel opened...")
	for {
		gameState := <-myPlayerChannel
		if gameState.err != nil {
			c.Close(websocket.StatusUnsupportedData, err.Error())
			return
		}
		log.Printf("New game state received")
		output := convertGameStateToOutput(gameState, playerId, playerName)
		err = wsjson.Write(ctx, c, output)
		if err != nil {
			c.Close(websocket.StatusInternalError, err.Error())
			log.Printf("Error in write")
			return
		}
	}
}
