package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Add unsubscribe if input checks fail
// Add error handling

//var listenAddr = "192.168.178.23:4000" //localhost:4000
var listenAddr string

func init() {
	host, ok := os.LookupEnv("GAMEHOST")
	if !ok {
		host = getLocalIP()
	}
	listenAddr = fmt.Sprintf("%s:4000", host)
	log.Printf("Listening on %s", listenAddr)
}

func getLocalIP() string {
	// GetLocalIP returns the non loopback local IP of the host
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

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
			"actionId": document.getElementById("action-id").value,
			"playerName": document.getElementById("player-name").value,
			"playerToken": document.getElementById("player-token").value,
			"cardId": document.getElementById("card-id").value,
			"gameToken": document.getElementById("game-token").value
		}));
	}
	websocket.onmessage = onMessage;
	websocket.onclose = onClose;
</script>
Player Name <input id="player-name"/></br>
Action Id <input id="action-id"/></br>
Player Token <input id="player-token"/></br>
Game Token <input id="game-token"/></br>
Card Id <input id="card-id"/></br>
<button onclick="onSend(this)">Send</button>
<div id="chat"></div>
</html>
`))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("main", goid())
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

func getGameFromMap(gameToken string) (Game, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	var val, ok = gameMap[gameToken]
	return val, ok
}

func addGameToMap(gameToken string, game Game) {
	mutex.Lock()
	defer mutex.Unlock()
	gameMap[gameToken] = game
}

func convertGameStateToOutput(gameState *GameState) GameOutput {
	var gameOutput = GameOutput{
		GameState: gameState,
		ErrorMsg:  "",
	}
	return gameOutput
}
func isValidAction(actionId string) bool {
	actions := [10]string{
		"create", "join", "start", "leave",
		"concentrate", "ready",
		"propose-star", "agree-star", "reject-star",
		"card",
	}
	for _, a := range actions {
		if a == actionId {
			return true
		}
	}
	return false
}

func main() {
	log.Printf("hello server")
	log.Println("main", goid())
	http.HandleFunc("/", rootHandler)
	http.Handle("/socket", http.HandlerFunc(runGame))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func extractDetails(raw map[string]interface{}) InputDetails {
	var gameToken, _ = raw["gameToken"].(string)
	var playerToken, _ = raw["playerToken"].(string)
	var playerName, _ = raw["playerName"].(string)
	var actionId, _ = raw["actionId"].(string)
	var cardId, _ = raw["card"].(int)
	var details = InputDetails{
		GameToken:   gameToken,
		PlayerToken: playerToken,
		PlayerName:  playerName,
		ActionId:    actionId,
		CardId:      cardId,
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

	var myPlayerToken string
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
		log.Printf("Checking inputs...")
		log.Printf("playerToken: %v", myPlayerToken)
		log.Printf("playerName: %v", myPlayerName)
		log.Printf("gameToken: %v", myGame.token)
		log.Printf("actionId: %v", gameDetails.ActionId)
		log.Printf("cardId: %v", gameDetails.CardId)
		if gameDetails.PlayerName == "" {
			c.Close(websocket.StatusUnsupportedData, "playerName corrupted")
			return
		}
		if !isValidAction(gameDetails.ActionId) {
			c.Close(websocket.StatusUnsupportedData, "actionId corrupted")
			return
		}
		switch gameDetails.ActionId {
		case "create":
			log.Printf("Creating game...")
			myPlayerName = gameDetails.PlayerName
			myGame = *NewGame()
			addGameToMap(myGame.token, myGame)
			go myGame.Start()
			myPlayerToken, myPlayerChannel = myGame.Subscribe(myPlayerName)
			go listenPlayerChannel(c, ctx, myPlayerChannel)
		case "join":
			log.Printf("Joining game...")
			_myGame, ok := getGameFromMap(gameDetails.GameToken)
			myGame = _myGame
			if !ok {
				c.Close(websocket.StatusUnsupportedData, "GameToken corrupted")
				return
			}
			myPlayerName = gameDetails.PlayerName
			myPlayerToken, myPlayerChannel = myGame.Subscribe(myPlayerName)
			go listenPlayerChannel(c, ctx, myPlayerChannel)
		default:
			if gameDetails.PlayerToken != myPlayerToken {
				c.Close(websocket.StatusUnsupportedData, "playerToken corrupted")
				return
			}
			if gameDetails.GameToken != myGame.token {
				c.Close(websocket.StatusUnsupportedData, "gameToken corrupted")
				return
			}
			if myPlayerName != gameDetails.PlayerName {
				c.Close(websocket.StatusUnsupportedData, "PlayerName corrupted")
				return
			}
		}
		log.Printf("Passing inputs to game core...")
		myGame.inputCh <- gameDetails
	}
}
func listenPlayerChannel(c *websocket.Conn, ctx context.Context, myPlayerChannel chan GameState) {
	var err error
	log.Printf("Player channel opened...")
	for {
		gameState := <-myPlayerChannel
		if gameState.err != nil {
			c.Close(websocket.StatusUnsupportedData, err.Error())
			return
		}
		log.Printf("New game state received")
		output := convertGameStateToOutput(&gameState)
		err = wsjson.Write(ctx, c, output)
		if err != nil {
			c.Close(websocket.StatusInternalError, err.Error())
			log.Printf("Error in write")
			return
		}
	}
}
