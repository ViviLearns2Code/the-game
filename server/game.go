package main

import (
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
)

func (g *Game) Start() {
	// loop
	var manager = NewGameManager()
	var gameState = NewGameState(g.token)
	nextPlayerId := 1
	for {
		select {
		case inputDetails := <-g.inputCh:
			if actionCheck(inputDetails, gameState) {
				gameLogicBasedOnAction(inputDetails, manager, gameState)
			}
			g.publishCh <- true
		case subscriber := <-g.subCh:
			if (len(gameState.PlayerNames) >= 4) || manager.started {
				var err = NewGameError("error", "cannot join game anymore")
				subscriber.playerChannel <- GameState{
					err: err,
				}
			} else {
				gameState.PlayerNames[nextPlayerId] = subscriber.playerName
				manager.playerId2Token[subscriber.playerToken] = nextPlayerId
				manager.subs[subscriber.playerToken] = subscriber.playerChannel
				nextPlayerId++
				g.publishCh <- true
			}
		case <-g.unsubCh:
			gameState.GameStateEvent.Name = "gameOver"
			g.publishCh <- true
		case <-g.publishCh:
			log.Printf("New game state published")
			convertFromGameManagerToChannelOutput(manager, gameState)
			gameState.updateEventsAfterProcessedEvent(manager.started)
		}
	}
}

func NewGame() *Game {
	return &Game{
		token:     uuid.New().String(), //concurrent reads only!
		inputCh:   make(chan InputDetails),
		publishCh: make(chan bool, 1),
		subCh:     make(chan subscription, 1),
		unsubCh:   make(chan string, 1),
	}
}

func describe(i interface{}) {
	log.Printf("(%v, %T)\n", i, i)
}

func (g *Game) Subscribe(playerName string) (string, chan GameState) {
	playerToken := uuid.New().String()
	playerChannel := make(chan GameState)
	g.subCh <- subscription{
		playerToken:   playerToken,
		playerName:    playerName,
		playerChannel: playerChannel,
	}
	return playerToken, playerChannel
}

func (g *Game) Unsubscribe(playerToken string) {
	g.unsubCh <- playerToken
}

func NewGameManager() *GameManager {
	return &GameManager{
		playerId2Token: nil,
		subs:           nil,
		started:        false,
		CardsManager:   nil,
	}
}

func NewGameState(gt string) *GameState {
	return &GameState{
		GameToken:        gt,
		PlayerToken:      nil,
		PlayerName:       nil,
		PlayerId:         nil,
		CardsOfPlayer:    nil,
		PlayerNames:      nil,
		CardsOnTable:     CardsOnTable{0, 0, 0, 0},
		GameStateEvent:   GameStateEvent{"", "", false, false, false, false},
		ReadyEvent:       ReadyEvent{"lobby", 0, nil},
		PlaceCardEvent:   nil,
		ProcessStarEvent: nil,
		err:              nil,
	}
}

func actionCheck(inputDetails InputDetails, gameState *GameState) bool {
	if gameState.ReadyEvent.Name != "" {
		if (inputDetails.ActionId != "start") || (inputDetails.ActionId != "ready") {
			gameState.err = NewGameError("warning", "wrong action:  game is not started or is in concentration")
			return false
		}
		if (gameState.ReadyEvent.Name == "lobby") || (inputDetails.ActionId != "start") {
			gameState.err = NewGameError("warning", "wrong action:  game is not started, start action is expected")
			return false
		}
		if (gameState.ReadyEvent.Name == "concentrate") || (inputDetails.ActionId != "ready") {
			gameState.err = NewGameError("warning", "wrong action:  game is in concentrating, ready action is expected")
			return false
		}
	}
	if gameState.ProcessStarEvent.Name == "agreeStar" {
		if (inputDetails.ActionId != "reject-Star") || (inputDetails.ActionId != "agree-star") {
			gameState.err = NewGameError("warning", "wrong action:  star is proposed, agree/reject star action is expected")
			return false
		}
	}
	return true
}

func gameLogicBasedOnAction(raw InputDetails, manager *GameManager, gameState *GameState) {
	currplayerIdx := manager.playerId2Token[raw.PlayerToken]
	switch raw.ActionId {
	case "start":
		gameState.ReadyEvent.Ready = append(gameState.ReadyEvent.Ready, currplayerIdx)
		actIfStartGame(manager, gameState)
	case "concentrate":
		gameState.ReadyEvent.triggerReadyEvent(currplayerIdx)
	case "ready":
		gameState.ReadyEvent.Ready = append(gameState.ReadyEvent.Ready, currplayerIdx)
	case "propose-star":
		gameState.ProcessStarEvent.triggerProcessStarEvent(currplayerIdx)
	case "agree-star":
		gameState.ProcessStarEvent.ProStar = append(gameState.ProcessStarEvent.ProStar, currplayerIdx)
		actIfUsingStar(manager, gameState)
	case "reject-star":
		gameState.ProcessStarEvent.Name = "reject-star"
		gameState.ProcessStarEvent.ConStar = append(gameState.ProcessStarEvent.ConStar, currplayerIdx)
		gameState.ReadyEvent.setReadyEventToCencentrate()
	default:
		if raw.CardId != 0 {
			if raw.CardId != manager.cardsInHands[currplayerIdx][0] {
				gameState.err = NewGameError("warning", "wrong action: you have a smaller card")
			} else {
				if wrongPlacedCard(raw.CardId, manager) {
					actDueToWrongPlacedCard(manager, gameState, currplayerIdx)
				} else {
					actDueToRightPlacedCard(manager, gameState, currplayerIdx, raw.CardId)
				}
			}
		}
	}

}

func actIfStartGame(communicator *GameManager, gameState *GameState) {
	nrOfPlayer := len(gameState.PlayerNames)
	if (nrOfPlayer != 1) && (nrOfPlayer == len(gameState.ReadyEvent.Ready)) {
		communicator.started = true
		communicator.CardsManager.createCardsManager(nrOfPlayer)
		communicator.CardsManager.handoutCards(nrOfPlayer)
		gameState.ReadyEvent.setReadyEventToCencentrate()
		gameState.GameStateEvent.Name = "levelup"
		gameState.GameStateEvent.LevelTitle = communicator.CardsManager.levelCards[1].levelTitle
	}
}
func (cards *CardsManager) createCardsManager(nrOfPlayer int) {
	maxLevel := 12
	switch nrOfPlayer {
	case 2:
		cards.CardsOnTable.Level = 1
		cards.CardsOnTable.Lives = 2
		cards.CardsOnTable.Stars = 1
	case 3:
		maxLevel = 10
		cards.CardsOnTable.Level = 1
		cards.CardsOnTable.Lives = 3
		cards.CardsOnTable.Stars = 1
	case 4:
		maxLevel = 8
		cards.CardsOnTable.Level = 1
		cards.CardsOnTable.Lives = 4
		cards.CardsOnTable.Stars = 1
	}
	for i := 1; i <= maxLevel; i++ {
		switch i {
		case 1:
			cards.levelCards[1] = LevelCard{"Erhöhte Sensibilitat", false, false}
		case 2:
			cards.levelCards[2] = LevelCard{"Verstärkte Empathie", false, true}
		case 3:
			cards.levelCards[3] = LevelCard{"Erweitertes Bewusstsein", true, false}
		case 4:
			cards.levelCards[4] = LevelCard{"Sub-kognitive Wahrnehmung", false, false}
		case 5:
			cards.levelCards[5] = LevelCard{"Gruppenbewusstsein", false, true}
		case 6:
			cards.levelCards[6] = LevelCard{"Gedankenwahrenehmung", true, false}
		case 7:
			cards.levelCards[7] = LevelCard{"Telepathische Kommnikation", false, false}
		case 8:
			cards.levelCards[8] = LevelCard{"Auserkörperliche Präsenz", false, true}
		case 9:
			cards.levelCards[9] = LevelCard{"Quantenbewusstsein", true, false}
		case 10:
			cards.levelCards[10] = LevelCard{"Abspaltung vom Raum-Zeit-Kontinuum", false, false}
		case 11:
			cards.levelCards[11] = LevelCard{"Metaphysische Harmonie", false, false}
		case 12:
			cards.levelCards[12] = LevelCard{"Verschmelzung von Geist und Materie", false, false}
		}
	}
}
func (cards *CardsManager) handoutCards(nrOfPlayer int) {
	cards.cardsInHands = make(map[int][]int)
	rand.Seed(time.Now().UTC().UnixNano())
	numberCards := rand.Perm(100)
	for i := 1; i <= cards.CardsOnTable.Level; i++ {
		for j := 1; j <= nrOfPlayer; j++ {
			cards.cardsInHands[j] = append(cards.cardsInHands[j], numberCards[len(numberCards)-1])
			numberCards = numberCards[:len(numberCards)-1]
		}
	}
	for _, cardsInHand := range cards.cardsInHands {
		sort.Ints(cardsInHand)
	}
}

func hasAnyCardsInHand(cardsInHands map[int][]int) bool {
	hasAnyCardsInHand := false
	for _, cardsInHand := range cardsInHands {
		hasAnyCardsInHand = len(cardsInHand) != 0
		if hasAnyCardsInHand {
			break
		}
	}
	return hasAnyCardsInHand
}
func levelFinish(communicator *GameManager, gameState *GameState) {
	gameState.GameStateEvent.Name = "levelUp"
	if communicator.levelCards[communicator.CardsOnTable.Level].lifeAsBonus {
		communicator.CardsOnTable.Lives++
		if !gameState.GameStateEvent.LivesDecrease {
			gameState.GameStateEvent.LivesIncrease = true
		} else {
			gameState.GameStateEvent.LivesIncrease = false
			gameState.GameStateEvent.LivesDecrease = false
		}
	}
	if communicator.levelCards[communicator.CardsOnTable.Level].starAsBonus {
		communicator.CardsOnTable.Stars++
		if !gameState.GameStateEvent.StarsDecrease {
			gameState.GameStateEvent.StarsIncrease = true
		} else {
			gameState.GameStateEvent.StarsIncrease = false
			gameState.GameStateEvent.StarsDecrease = false
		}
	}
	communicator.CardsOnTable.Level++
	gameState.GameStateEvent.LevelTitle = communicator.levelCards[communicator.CardsOnTable.Level].levelTitle
	communicator.CardsOnTable.TopCard = 0
}
func handleGameoverOrLevelFinish(communicator *GameManager, gameState *GameState) {
	if communicator.CardsOnTable.Level == len(communicator.levelCards) {
		gameState.GameStateEvent.Name = "gameOver"
	} else {
		levelFinish(communicator, gameState)
		communicator.CardsManager.handoutCards(len(gameState.PlayerNames))
		gameState.ReadyEvent.setReadyEventToCencentrate()
	}
}

func actIfUsingStar(communicator *GameManager, gameState *GameState) {
	nrOfPlayer := len(gameState.PlayerNames)
	if nrOfPlayer == len(gameState.ProcessStarEvent.ProStar) {
		for playerIdx, cardsInHand := range communicator.cardsInHands {
			communicator.discardedCards[playerIdx], communicator.cardsInHands[playerIdx] = cardsInHand[0], cardsInHand[1:]
		}
		gameState.PlaceCardEvent.setPlaceCardEvent("useStar", 0, communicator.discardedCards)
		communicator.CardsOnTable.Stars--
		gameState.GameStateEvent.StarsDecrease = true
		gameState.GameStateEvent.StarsIncrease = false
		gameState.GameStateEvent.LivesIncrease = false
		gameState.GameStateEvent.LivesDecrease = false
		if hasAnyCardsInHand(communicator.cardsInHands) {
			gameState.ReadyEvent.setReadyEventToCencentrate()
		} else {
			handleGameoverOrLevelFinish(communicator, gameState)
		}
	}
}

func wrongPlacedCard(currentCard int, manager *GameManager) bool {
	for playerIdx, cardsInHand := range manager.cardsInHands {
		if cardsInHand[0] < currentCard {
			manager.discardedCards[playerIdx], manager.cardsInHands[playerIdx] = cardsInHand[0], cardsInHand[1:]
		}
	}
	return len(manager.discardedCards) != 0
}
func actDueToWrongPlacedCard(communicator *GameManager, gameState *GameState, currplayerIdx int) {
	gameState.PlaceCardEvent.setPlaceCardEvent("placeCard", currplayerIdx, communicator.discardedCards)
	communicator.CardsOnTable.Lives--
	gameState.GameStateEvent.LivesDecrease = true
	gameState.GameStateEvent.LivesIncrease = false
	gameState.GameStateEvent.StarsIncrease = false
	gameState.GameStateEvent.StarsDecrease = false
	if communicator.CardsOnTable.Lives == 0 {
		gameState.GameStateEvent.Name = "gameOver"
	} else if hasAnyCardsInHand(communicator.cardsInHands) {
		gameState.GameStateEvent.Name = "lostLife"
		gameState.ReadyEvent.setReadyEventToCencentrate()
	} else {
		handleGameoverOrLevelFinish(communicator, gameState)
	}
}
func actDueToRightPlacedCard(communicator *GameManager, gameState *GameState, currplayerIdx int, currentCard int) {
	communicator.CardsOnTable.TopCard = currentCard
	communicator.discardedCards[currplayerIdx], communicator.cardsInHands[currplayerIdx] = communicator.cardsInHands[currplayerIdx][0], communicator.cardsInHands[currplayerIdx][1:]
	gameState.PlaceCardEvent.setPlaceCardEvent("placeCard", currplayerIdx, communicator.discardedCards)
	if !hasAnyCardsInHand(communicator.cardsInHands) {
		handleGameoverOrLevelFinish(communicator, gameState)
	}
}

func convertFromGameManagerToChannelOutput(communicator *GameManager, game *GameState) {
	for playerToken, playerChannel := range communicator.subs {
		game.PlayerId = communicator.playerId2Token[playerToken]
		game.PlayerToken = playerToken
		game.PlayerName = game.PlayerNames[game.PlayerId]
		game.CardsOfPlayer = cardsInHandOfPlayer(game.PlayerId, communicator.cardsInHands)
		game.CardsOnTable = communicator.CardsOnTable
		select {
		case playerChannel <- *game:
			// handled by goroutine in main.go
		default:
		}
	}
}
func cardsInHandOfPlayer(playerIdx int, cardsInHands map[int][]int) (cardsOfPlayer CardsOfPlayer) {
	cardsOfPlayer.CardsInHand = cardsInHands[playerIdx]
	for playerId, cards := range cardsInHands {
		cardsOfPlayer.NrCardsOfOtherPlayers[playerId] = cap(cards)
	}
	return cardsOfPlayer
}
func (game *GameState) updateEventsAfterProcessedEvent(started bool) {
	if (len(game.ReadyEvent.Ready) == len(game.PlayerNames)) && started {
		game.ReadyEvent.resetReadyEvent()
	}
	if game.ProcessStarEvent.Name == "proposeStar" {
		game.ProcessStarEvent.Name = "agreeStar"
	} else if ((game.ProcessStarEvent.Name == "agreeStar") && (len(game.ProcessStarEvent.ProStar) == len(game.PlayerNames))) || (game.ProcessStarEvent.Name == "rejectStar") {
		game.ProcessStarEvent.resetProcessStarEvent()
	}
	game.PlaceCardEvent.resetPlaceCardEvent()
	game.GameStateEvent.resetGameStateEvent()
}

func (readyEvent *ReadyEvent) triggerReadyEvent(triggedby int) {
	readyEvent.Name = "concentrate"
	readyEvent.TriggeredBy = triggedby
	readyEvent.Ready = append(readyEvent.Ready, triggedby)
}
func (readyEvent *ReadyEvent) resetReadyEvent() {
	readyEvent.Name = ""
	readyEvent.TriggeredBy = 0
	readyEvent.Ready = nil
}
func (readyEvent *ReadyEvent) setReadyEventToCencentrate() {
	readyEvent.Name = "concentrate"
	readyEvent.TriggeredBy = 0
	readyEvent.Ready = nil
}
func (placeCardEvent *PlaceCardEvent) setPlaceCardEvent(name string, triggeredBy int, discardedCards map[int]int) {
	placeCardEvent.Name = name
	placeCardEvent.TriggeredBy = triggeredBy
	placeCardEvent.DiscardedCard = discardedCards
}
func (placeCardEvent *PlaceCardEvent) resetPlaceCardEvent() {
	placeCardEvent.Name = ""
	placeCardEvent.TriggeredBy = 0
	placeCardEvent.DiscardedCard = nil
}
func (processCardEvent *ProcessStarEvent) resetProcessStarEvent() {
	processCardEvent.Name = ""
	processCardEvent.TriggeredBy = 0
	processCardEvent.ProStar = nil
	processCardEvent.ConStar = nil
}
func (processCardEvent *ProcessStarEvent) triggerProcessStarEvent(triggedby int) {
	processCardEvent.Name = "propose-star"
	processCardEvent.TriggeredBy = triggedby
	processCardEvent.ProStar = append(processCardEvent.ProStar, triggedby)
	processCardEvent.ConStar = nil
}
func (gameStateEvent *GameStateEvent) resetGameStateEvent() {
	gameStateEvent.Name = ""
	gameStateEvent.LevelTitle = ""
	gameStateEvent.StarsIncrease = false
	gameStateEvent.LivesIncrease = false
	gameStateEvent.StarsDecrease = false
	gameStateEvent.LivesDecrease = false
}
