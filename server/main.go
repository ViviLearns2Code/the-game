package main

import (
	"context"
	"fmt"
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
	log.Println("starting server")
	mux := http.NewServeMux()
	mux.Handle("/socket", http.HandlerFunc(runGame))
	s := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err := s.ListenAndServe()
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
	log.Println("checking inputs...")
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
	log.Printf("number of goroutines %d", runtime.NumGoroutine())
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: false,
		OriginPatterns:     []string{gameOrigin},
	})
	log.Println("connection established...")
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour*1)
	defer cancel()

	var myPlayerToken string
	var myPlayerIconId int
	var myPlayerName string
	var myGame Game
	var myPlayerChannel chan GameState
	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusInternalError, ctx.Err().Error())
			return
		default:
		}
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
			output.ErrorMsg = "wrong input"
			err = wsjson.Write(ctx, c, output)
			if err != nil {
				// when write fails, it is too broken
				log.Println(err.Error())
				g, ok := getGameFromMap(myGame.token)
				if ok {
					g.Unsubscribe(myPlayerToken)
				}
				g, ok = getGameFromMap(gameDetails.GameToken)
				if ok {
					g.Unsubscribe(gameDetails.PlayerToken)
				}
				c.Close(websocket.StatusInternalError, err.Error())
				return
			}
			continue
		}
		switch gameDetails.ActionId {
		case "create":
			isBorg := rollDice()
			myPlayerName = gameDetails.PlayerName
			myPlayerIconId = gameDetails.PlayerIconId
			myGame = *NewGame()
			addGameToMap(myGame.token, myGame)
			go myGame.Start(isBorg)
			log.Printf("game %s is started", myGame.token)
			myPlayerToken, myPlayerChannel = myGame.Subscribe(myPlayerName, myPlayerIconId)
			gameDetails.PlayerToken = myPlayerToken
			gameDetails.GameToken = myGame.token
			go listenPlayerChannel(c, ctx, myPlayerChannel, myPlayerToken)
			log.Printf("player %s registered to game %s", myPlayerToken, myGame.token)
		case "join":
			myGame, _ = getGameFromMap(gameDetails.GameToken)
			myPlayerName = gameDetails.PlayerName
			myPlayerIconId = gameDetails.PlayerIconId
			myPlayerToken, myPlayerChannel = myGame.Subscribe(myPlayerName, myPlayerIconId)
			gameDetails.PlayerToken = myPlayerToken
			go listenPlayerChannel(c, ctx, myPlayerChannel, myPlayerToken)
			log.Printf("player %s registered to game %s", myPlayerToken, myGame.token)
		case "leave":
			log.Println("leaving game...")
			myGame, _ := getGameFromMap(gameDetails.GameToken)
			myGame.Unsubscribe(myPlayerToken)
			log.Printf("player %s unregistered from game %s", myPlayerToken, myGame.token)
		default:
			myGame.inputCh <- gameDetails
			log.Printf("player %s submitted action %s", myPlayerToken, gameDetails.ActionId)
		}
	}
}
func listenPlayerChannel(c *websocket.Conn, ctx context.Context, myPlayerChannel chan GameState, myPlayerToken string) {
	var err error
	log.Printf("%d: channel opened for player %s", goid(), myPlayerToken)
	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusInternalError, ctx.Err().Error())
			return
		default:
		}
		gameState, ok := <-myPlayerChannel
		if !ok {
			log.Printf("%d: channel for player %s closed!", goid(), myPlayerToken)
			return
		}
		log.Printf("%d: player %s received game event %s for game %s", goid(), myPlayerToken, gameState.GameStateEvent.Name, gameState.GameToken)
		if gameState.GameStateEvent.Name == "gameOver" {
			removeGameFromMap(gameState.GameToken)
			log.Printf("game over: removed game %s", gameState.GameToken)
			output := convertGameStateToOutput(&gameState)
			err = wsjson.Write(ctx, c, output)
			if err != nil {
				log.Printf("error when writing to player channel %s for game %s", gameState.PlayerToken, gameState.GameToken)
				c.Close(websocket.StatusInternalError, err.Error())
				return
			}
			c.Close(websocket.StatusNormalClosure, "Game Over")
			return
		}
		g, ok := getGameFromMap(gameState.GameToken)
		if !ok {
			// should never happen
			log.Printf("%d: game %s does not exist: panic", goid(), gameState.GameToken)
			panic("game returned invalid game token: shutting down")
		}
		if gameState.err != nil {
			if gameState.err.severity == "fatal" {
				log.Printf("%d: fatal error in game %s: %s", goid(), gameState.GameToken, gameState.err.Error())
				g.Unsubscribe(myPlayerToken)
				c.Close(websocket.StatusUnsupportedData, err.Error())
				return
			}
		}
		output := convertGameStateToOutput(&gameState)
		err = wsjson.Write(ctx, c, output)
		if err != nil {
			log.Printf("error when writing response for player %s in game %s", myPlayerToken, gameState.GameToken)
			g.Unsubscribe(myPlayerToken)
			c.Close(websocket.StatusInternalError, err.Error())
			return
		}
	}
}
