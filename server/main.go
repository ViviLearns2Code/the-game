package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
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

var listenAddr string
var gameOrigin string

// Global map, protected by lock
var mutex = &sync.RWMutex{}
var gameMap = make(map[string]Game)

func init() {
	var host, port string
	var ok bool
	host, ok = os.LookupEnv("GAMEHOST")
	if !ok {
		host = getLocalIP()
	}
	port, ok = os.LookupEnv("GAMEPORT")
	if !ok {
		port = "4000"
	}
	listenAddr = fmt.Sprintf("%s:%s", host, port)
	log.Printf("Listening on %s", listenAddr)
	gameOrigin, ok = os.LookupEnv("GAMEORIGIN")
	if !ok {
		gameOrigin = "localhost:8000"
	}
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

func removeGameFromMap(gameToken string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(gameMap, gameToken)
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

func convertGameStateToOutput(gameState *GameState) GameOutput {
	var err string
	if gameState.err != nil {
		err = gameState.err.Error()
	}
	var gameOutput = GameOutput{
		GameState: gameState,
		ErrorMsg:  err,
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
	var playerIconId, _ = raw["playerIconId"].(float64)
	var actionId, _ = raw["actionId"].(string)
	var cardId, _ = raw["cardId"].(float64)
	var details = InputDetails{
		GameToken:    gameToken,
		PlayerToken:  playerToken,
		PlayerName:   playerName,
		ActionId:     actionId,
		PlayerIconId: int(playerIconId),
		CardId:       int(cardId),
	}
	return details
}

func throwOut(myGame Game, myPlayerToken string) {
	myGame.Unsubscribe(myPlayerToken)
}

func isValidGame(gameToken string, gameTokenPrev string) bool {
	_, ok := getGameFromMap(gameToken)
	return ok && (gameToken == gameTokenPrev)
}
func isValidPlayer(playerToken string, playerTokenPrev string) bool {
	return (playerToken == playerTokenPrev)
}
func validateInput(gameDetails InputDetails, myGame Game, myPlayerToken string, myPlayerName string, myPlayerIconId int) bool {
	log.Printf("Checking inputs...")
	var ok = true
	// universal checks
	if gameDetails.PlayerName == "" {
		ok = false
		return ok
	}
	if !isValidAction(gameDetails.ActionId) {
		ok = false
		return ok
	}
	if (gameDetails.PlayerIconId < 0) && (gameDetails.PlayerIconId > 4) {
		ok = false
		return ok
	}
	// checks for create: fully covered
	if gameDetails.ActionId == "create" {
		return ok
	}
	// checks for join: fully covered
	if gameDetails.ActionId == "join" {
		if _, found := getGameFromMap(gameDetails.GameToken); !found {
			ok = false
		}
		return ok
	}
	// common checks for remaining actions
	if !isValidGame(gameDetails.GameToken, myGame.token) {
		ok = false
		return ok
	}
	if !isValidPlayer(gameDetails.PlayerToken, myPlayerToken) {
		ok = false
		return ok
	}
	// specific check for playing cards
	if gameDetails.ActionId == "card" {
		if gameDetails.CardId < 1 || gameDetails.CardId > 100 {
			ok = false
			return ok
		}
	}
	return ok
}

func rollDice() (isBorg bool) {
	rand.Seed(time.Now().UnixNano())
	random := rand.Float64()
	if random < 0.5 {
		isBorg = false
	} else {
		isBorg = true
	}
	return isBorg
}

func runGame(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection established...")
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: false,
		OriginPatterns:     []string{gameOrigin},
	})
	if err != nil {
		log.Printf(err.Error())
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour*120000)
	defer cancel()

	var myPlayerToken string
	var myPlayerIconId int
	var myPlayerName string
	var myGame Game
	var myPlayerChannel chan GameState
	log.Printf("Entering loop...")
	for {
		var data = make(map[string]interface{})

		var err = wsjson.Read(ctx, c, &data)
		if err != nil {
			log.Println(err.Error())
			// could happen if client closed connection
			throwOut(myGame, myPlayerToken)
			c.Close(websocket.StatusAbnormalClosure, err.Error())
			return
		}
		gameDetails := extractDetails(data)
		inputOk := validateInput(gameDetails, myGame, myPlayerToken, myPlayerName, myPlayerIconId)
		if !inputOk {
			output := convertGameStateToOutput(newGameState(myGame.token))
			output.ErrorMsg = "Wrong input"
			err = wsjson.Write(ctx, c, output)
			if err != nil {
				// when write fails, it is too broken
				log.Printf(err.Error())
				g, ok := getGameFromMap(myGame.token)
				if ok {
					g.Unsubscribe(myPlayerToken)
				}
				g, ok = getGameFromMap(gameDetails.GameToken)
				if ok {
					g.Unsubscribe(gameDetails.PlayerToken)
				}
				log.Println("Input check hahahaha", err.Error())
				c.Close(websocket.StatusInternalError, err.Error())
				return
			}
			continue
		}
		switch gameDetails.ActionId {
		case "create":
			log.Printf("Creating game...")
			isBorg := rollDice()
			myPlayerName = gameDetails.PlayerName
			myPlayerIconId = gameDetails.PlayerIconId
			myGame = *NewGame()
			addGameToMap(myGame.token, myGame)
			go myGame.Start(isBorg)
			myPlayerToken, myPlayerChannel = myGame.Subscribe(myPlayerName, myPlayerIconId)
			gameDetails.PlayerToken = myPlayerToken
			gameDetails.GameToken = myGame.token
			go listenPlayerChannel(c, ctx, myPlayerChannel)
		case "join":
			log.Printf("Joining game...")
			myGame, _ = getGameFromMap(gameDetails.GameToken)
			myPlayerName = gameDetails.PlayerName
			myPlayerIconId = gameDetails.PlayerIconId
			myPlayerToken, myPlayerChannel = myGame.Subscribe(myPlayerName, myPlayerIconId)
			gameDetails.PlayerToken = myPlayerToken
			go listenPlayerChannel(c, ctx, myPlayerChannel)
		case "leave":
			log.Printf("Leaving game...")
			myGame, _ := getGameFromMap(gameDetails.GameToken)
			myGame.Unsubscribe(myPlayerToken)
		default:
			log.Printf("Action %s", gameDetails.ActionId)
			myGame.inputCh <- gameDetails
		}
	}
}
func listenPlayerChannel(c *websocket.Conn, ctx context.Context, myPlayerChannel chan GameState) {
	var err error
	log.Printf("%v: Player channel opened...", goid())
	for {
		gameState, ok := <-myPlayerChannel
		log.Printf("%v: Player channel received something...", goid())
		if !ok {
			log.Println("closed!")
			return
		}
		log.Printf("%v: event name %s", goid(), gameState.GameStateEvent.Name)
		log.Printf("%v: game token %s", goid(), gameState.GameToken)
		if gameState.GameStateEvent.Name == "gameOver" {
			log.Printf("Game over! Removing game...")
			removeGameFromMap(gameState.GameToken)
			output := convertGameStateToOutput(&gameState)
			err = wsjson.Write(ctx, c, output)
			if err != nil {
				log.Printf("Error when writing to player channel %s for game %s", gameState.PlayerToken, gameState.GameToken)
				c.Close(websocket.StatusInternalError, err.Error())
				return
			}
			c.Close(websocket.StatusNormalClosure, "Game Over")
			return
		}
		g, ok := getGameFromMap(gameState.GameToken)
		if !ok {
			// should never happen
			log.Printf("%v: token incorrect", goid())
			panic("Game returned invalid gameToken - shutting down...")
		}
		if gameState.err != nil {
			if gameState.err.severity == "fatal" {
				log.Println("Fatal game error", gameState.err.Error())
				g.Unsubscribe(gameState.PlayerToken)
				c.Close(websocket.StatusUnsupportedData, err.Error())
				return
			}
		}
		log.Printf("New game state received")
		output := convertGameStateToOutput(&gameState)
		err = wsjson.Write(ctx, c, output)
		if err != nil {
			log.Printf("Error when writing to player channel %s for game %s", gameState.PlayerToken, gameState.GameToken)
			g.Unsubscribe(gameState.PlayerToken)
			c.Close(websocket.StatusInternalError, err.Error())
			return
		}
	}
}
