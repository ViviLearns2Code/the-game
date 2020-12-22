package main

import (
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (g *Game) Start(isBorg bool) {
	// game loop
	var manager = newGameManager()
	var gameState = newGameState(g.token)
	nextPlayerID := 1
	for {
		select {
		case inputDetails := <-g.inputCh:
			//log.Printf("inputDetails := <-g.inputCh")
			if actionCheck(inputDetails, gameState, manager) {
				gameLogicBasedOnAction(inputDetails, manager, gameState)
			}
			changeNamesToBorgsIfNeeded(manager, gameState, isBorg)
			convertFromGameManagerToChannelOutput(manager, gameState)
			if gameState.GameStateEvent.Name == "gameOver" {
				return
			}
			gameState.updateEventsAfterProcessedEvent(manager.started)
		case subscriber := <-g.subCh:
			//log.Printf("subscriber := <-g.subCh")
			if (len(gameState.PlayerNames) >= 4) || manager.started {
				var err = NewGameError("error", "cannot join game anymore")
				var gs = newGameState(g.token)
				gs.err = err
				subscriber.playerChannel <- *gs
			} else {
				gameState.PlayerNames[nextPlayerID] = subscriber.playerName
				gameState.PlayerIconIds[nextPlayerID] = subscriber.playerIconId
				manager.playerTokenToID[subscriber.playerToken] = nextPlayerID
				manager.subs[subscriber.playerToken] = subscriber.playerChannel
				nextPlayerID++
				convertFromGameManagerToChannelOutput(manager, gameState)
				gameState.updateEventsAfterProcessedEvent(manager.started)
			}
		case playerToken := <-g.unsubCh:
			// if playertoken not valid, ignore unsubscribe
			if playerChannel, ok := manager.subs[playerToken]; ok {
				gameState.GameStateEvent.Name = "gameOver"
				close(playerChannel)
				delete(manager.subs, playerToken)
				convertFromGameManagerToChannelOutput(manager, gameState)
				gameState.updateEventsAfterProcessedEvent(manager.started)
				return
			}
		}
	}
}
func changeNamesToBorgsIfNeeded(manager *GameManager, gameState *GameState, isBorg bool) {
	if isBorg && manager.CardsOnTable.Level == 7 {
		str := "of"
		str += strconv.Itoa(len(gameState.PlayerNames))
		for playerID, _ := range gameState.PlayerNames {
			str1 := strconv.Itoa(playerID)
			newname := str1 + str
			gameState.PlayerNames[playerID] = newname
		}
	}
}
func NewGame() *Game {
	return &Game{
		token:   uuid.New().String()[:6], //concurrent reads only!
		inputCh: make(chan InputDetails),
		subCh:   make(chan subscription, 1),
		unsubCh: make(chan string, 1),
	}
}

func describe(i interface{}) {
	log.Printf("(%v, %T)\n", i, i)
}

func (g *Game) Subscribe(playerName string, playerIconId int) (string, chan GameState) {
	playerToken := uuid.New().String()
	playerChannel := make(chan GameState, 1)
	g.subCh <- subscription{
		playerToken:   playerToken,
		playerIconId:  playerIconId,
		playerName:    playerName,
		playerChannel: playerChannel,
	}
	return playerToken, playerChannel
}

func (g *Game) Unsubscribe(playerToken string) {
	g.unsubCh <- playerToken
}

func newGameManager() *GameManager {
	return &GameManager{
		playerTokenToID: make(map[string]int),
		subs:            make(map[string]chan GameState),
		started:         false,
		CardsManager: CardsManager{cardsInHands: make(map[int][]int),
			CardsOnTable:   CardsOnTable{TopCard: 0, Level: 0, Lives: 0, Stars: 0},
			levelCards:     make(map[int]LevelCard),
			discardedCards: make(map[int][]int)},
	}
}

func newGameState(gt string) *GameState {
	return &GameState{
		GameToken:        gt,
		PlayerToken:      "",
		PlayerName:       "",
		PlayerId:         0,
		CardsOfPlayer:    CardsOfPlayer{CardsInHand: make([]int, 0), NrCardsOfOtherPlayers: make(map[int]int)},
		PlayerNames:      make(map[int]string),
		PlayerIconIds:    make(map[int]int),
		CardsOnTable:     CardsOnTable{TopCard: 0, Level: 0, Lives: 0, Stars: 0},
		GameStateEvent:   GameStateEvent{Name: "", LevelTitle: "", StarsIncrease: false, StarsDecrease: false, LivesIncrease: false, LivesDecrease: false},
		ReadyEvent:       ReadyEvent{Name: "lobby", TriggeredBy: 0, Ready: make([]int, 0)},
		PlaceCardEvent:   PlaceCardEvent{Name: "", TriggeredBy: 0, DiscardedCard: make(map[int][]int)},
		ProcessStarEvent: ProcessStarEvent{Name: "", TriggeredBy: 0, ProStar: make([]int, 0), ConStar: make([]int, 0)},
		err:              nil,
	}
}
func actionCheckReadyEventRelevant(inputDetails InputDetails, gameState *GameState, communicator *GameManager) bool {
	x0 := inputDetails.ActionId != "start"
	x1 := inputDetails.ActionId != "ready"
	if (gameState.ReadyEvent.Name == "lobby") && x0 {
		gameState.err = NewGameError("warning", "wrong action:  game is not started, start action is expected")
		return false
	}
	if (gameState.ReadyEvent.Name == "concentrate") && x1 {
		gameState.err = NewGameError("warning", "wrong action:  game is in concentrating, ready action is expected")
		return false
	}
	if gameState.ReadyEvent.Name == "" && (!x0 || !x1) {
		gameState.err = NewGameError("warning", "wrong action:  no ready action is expected")
		return false
	}
	return true
}

func actionCheckStarEventRelevant(inputDetails InputDetails, gameState *GameState, communicator *GameManager) bool {
	if inputDetails.ActionId == "propose-star" && communicator.CardsOnTable.Stars == 0 {
		gameState.err = NewGameError("warning", "wrong action:  no stars left")
		return false
	}
	x0 := inputDetails.ActionId == "reject-star"
	x1 := inputDetails.ActionId == "agree-star"
	if gameState.ProcessStarEvent.Name == "agreeStar" {
		if !x0 && !x1 {
			gameState.err = NewGameError("warning", "wrong action:  star is proposed, agree/reject star action is expected")
			return false
		}
	} else {
		if x0 || x1 {
			gameState.err = NewGameError("warning", "wrong action:  no star is proposed or star proposal has been handled, agree/reject star action is not accepted")
			return false
		}
	}
	return true
}
func actionCheck(inputDetails InputDetails, gameState *GameState, communicator *GameManager) bool {
	x1 := inputDetails.ActionId == "create"
	x2 := inputDetails.ActionId == "join"
	x3 := inputDetails.ActionId == "leave"
	if x1 || x2 || x3 {
		gameState.err = NewGameError("warning", "the actions are happened")
		return false
	}
	if !actionCheckReadyEventRelevant(inputDetails, gameState, communicator) {
		return false
	}
	if !actionCheckStarEventRelevant(inputDetails, gameState, communicator) {
		return false
	}
	return true
}

func gameLogicBasedOnAction(raw InputDetails, manager *GameManager, gameState *GameState) {
	currplayerIdx := manager.playerTokenToID[raw.PlayerToken]
	switch raw.ActionId {
	case "start":
		appendElementIfUnique(&gameState.ReadyEvent.Ready, currplayerIdx)
		actIfStartGame(manager, gameState)
	case "concentrate":
		gameState.ReadyEvent.triggerReadyEvent(currplayerIdx)
	case "ready":
		appendElementIfUnique(&gameState.ReadyEvent.Ready, currplayerIdx)
	case "propose-star":
		gameState.ProcessStarEvent.triggerProcessStarEvent(currplayerIdx)
	case "agree-star":
		appendElementIfUnique(&gameState.ProcessStarEvent.ProStar, currplayerIdx)
		actIfUsingStar(manager, gameState)
	case "reject-star":
		gameState.ProcessStarEvent.Name = "rejectStar"
		appendElementIfUnique(&gameState.ProcessStarEvent.ConStar, currplayerIdx)
		gameState.ReadyEvent.setReadyEventToCencentrate()
	default:
		if raw.CardId != 0 {
			if raw.CardId != manager.cardsInHands[currplayerIdx][0] {
				gameState.err = NewGameError("warning", "wrong action: you have a smaller card")
			} else {
				if wrongPlacedCard(raw.CardId, manager) {
					actDueToWrongPlacedCard(manager, gameState, currplayerIdx, raw.CardId)
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
		gameState.GameStateEvent.Name = "levelUp"
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
			cards.levelCards[1] = LevelCard{"Initialization: Game Trek", false, false}
		case 2:
			cards.levelCards[2] = LevelCard{"Systems: online", false, true}
		case 3:
			cards.levelCards[3] = LevelCard{"Tactical Maneuver 1.A", true, false}
		case 4:
			cards.levelCards[4] = LevelCard{"United Federation", false, false}
		case 5:
			cards.levelCards[5] = LevelCard{"Black Alert", false, true}
		case 6:
			cards.levelCards[6] = LevelCard{"Jump", true, false}
		case 7:
			cards.levelCards[7] = LevelCard{"My mind to your mind, my thoughts to your thoughts", false, false}
		case 8:
			cards.levelCards[8] = LevelCard{"Absolute Mind Meld", false, true}
		case 9:
			cards.levelCards[9] = LevelCard{"Collective Consciousness", true, false}
		case 10:
			cards.levelCards[10] = LevelCard{"Absolute synchronized Public Spirit", false, false}
		case 11:
			cards.levelCards[11] = LevelCard{"Metamorphosis of Spirit: Seperation of Space-Time", false, false}
		case 12:
			cards.levelCards[12] = LevelCard{"Q-Continuum", false, false}
		}
	}
}
func appendElementIfUnique(intslice *([]int), element int) {
	doesElementExist := false
	for _, e := range *intslice {
		if e == element {
			doesElementExist = true
			break
		}
	}
	if !doesElementExist {
		*intslice = append(*intslice, element)
	}
}
func (cards *CardsManager) handoutCards(nrOfPlayer int) {
	cards.cardsInHands = make(map[int][]int)
	rand.Seed(time.Now().UTC().UnixNano())
	numberCards := rand.Perm(100)
	for i := 1; i <= cards.CardsOnTable.Level; i++ {
		for j := 1; j <= nrOfPlayer; j++ {
			cards.cardsInHands[j] = append(cards.cardsInHands[j], numberCards[len(numberCards)-1]+1)
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
		if communicator.CardsOnTable.Lives < 5 {
			communicator.CardsOnTable.Lives++
			if !gameState.GameStateEvent.LivesDecrease {
				gameState.GameStateEvent.LivesIncrease = true
			} else {
				gameState.GameStateEvent.LivesIncrease = false
				gameState.GameStateEvent.LivesDecrease = false
			}
		}
	}
	if communicator.levelCards[communicator.CardsOnTable.Level].starAsBonus {
		if communicator.CardsOnTable.Stars < 3 {
			communicator.CardsOnTable.Stars++
			if !gameState.GameStateEvent.StarsDecrease {
				gameState.GameStateEvent.StarsIncrease = true
			} else {
				gameState.GameStateEvent.StarsIncrease = false
				gameState.GameStateEvent.StarsDecrease = false
			}
		}
	}
	communicator.CardsOnTable.Level++
	gameState.GameStateEvent.LevelTitle = communicator.levelCards[communicator.CardsOnTable.Level].levelTitle
	communicator.CardsOnTable.TopCard = 0
}
func handleGameoverOrLevelFinish(manager *GameManager, gameState *GameState) {
	if manager.CardsOnTable.Level == len(manager.levelCards) {
		gameState.GameStateEvent.Name = "gameWon"
	} else {
		levelFinish(manager, gameState)

		manager.CardsManager.handoutCards(len(gameState.PlayerNames))
		gameState.ReadyEvent.setReadyEventToCencentrate()
	}
}

func actIfUsingStar(manager *GameManager, gameState *GameState) {
	nrOfPlayer := len(gameState.PlayerNames)
	if nrOfPlayer == len(gameState.ProcessStarEvent.ProStar) {
		smallestCard := 100
		for playerIdx, cardsInHand := range manager.cardsInHands {
			manager.discardedCards[playerIdx], manager.cardsInHands[playerIdx] = cardsInHand[:1], cardsInHand[1:]
			if manager.discardedCards[playerIdx][0] < smallestCard {
				smallestCard = manager.discardedCards[playerIdx][0]
			}
		}
		gameState.PlaceCardEvent.setPlaceCardEvent("useStar", 0, &manager.discardedCards)
		manager.CardsOnTable.Stars--
		manager.CardsOnTable.TopCard = smallestCard
		gameState.GameStateEvent.StarsDecrease = true
		gameState.GameStateEvent.StarsIncrease = false
		gameState.GameStateEvent.LivesIncrease = false
		gameState.GameStateEvent.LivesDecrease = false
		if hasAnyCardsInHand(manager.cardsInHands) {
			gameState.ReadyEvent.setReadyEventToCencentrate()
		} else {
			handleGameoverOrLevelFinish(manager, gameState)
		}
	}
}

func wrongPlacedCard(currentCard int, manager *GameManager) bool {
	for playerIdx, cardsInHand := range manager.cardsInHands {
		hasSmallerCard := false
		n := 0
		for ; n < len(cardsInHand); n++ {
			if cardsInHand[n] >= currentCard {
				break
			} else {
				hasSmallerCard = true
			}
		}
		if hasSmallerCard {
			manager.discardedCards[playerIdx], manager.cardsInHands[playerIdx] = cardsInHand[:n], cardsInHand[n:]
		}
	}
	return len(manager.discardedCards) != 0
}

func setTopCardsUpdateCurrPlayerCardsInHandAndPlaceCardEvent(manager *GameManager, gameState *GameState, currplayerIdx int, currentCard int) {
	manager.CardsOnTable.TopCard = currentCard
	manager.discardedCards[currplayerIdx], manager.cardsInHands[currplayerIdx] = manager.cardsInHands[currplayerIdx][:1], manager.cardsInHands[currplayerIdx][1:]
	gameState.PlaceCardEvent.setPlaceCardEvent("placeCard", currplayerIdx, &manager.discardedCards)
}
func actDueToWrongPlacedCard(manager *GameManager, gameState *GameState, currplayerIdx int, currentCard int) {
	setTopCardsUpdateCurrPlayerCardsInHandAndPlaceCardEvent(manager, gameState, currplayerIdx, currentCard)
	manager.CardsOnTable.Lives--
	gameState.GameStateEvent.LivesDecrease = true
	gameState.GameStateEvent.LivesIncrease = false
	gameState.GameStateEvent.StarsIncrease = false
	gameState.GameStateEvent.StarsDecrease = false
	if manager.CardsOnTable.Lives == 0 {
		gameState.GameStateEvent.Name = "gameOver"
	} else if hasAnyCardsInHand(manager.cardsInHands) {
		gameState.GameStateEvent.Name = "lostLife"
		gameState.ReadyEvent.setReadyEventToCencentrate()
	} else {
		handleGameoverOrLevelFinish(manager, gameState)
	}
}
func actDueToRightPlacedCard(manager *GameManager, gameState *GameState, currplayerIdx int, currentCard int) {
	setTopCardsUpdateCurrPlayerCardsInHandAndPlaceCardEvent(manager, gameState, currplayerIdx, currentCard)
	if !hasAnyCardsInHand(manager.cardsInHands) {
		handleGameoverOrLevelFinish(manager, gameState)
	}
}
func convertFromGameManagerToChannelOutput(manager *GameManager, gameState *GameState) {
	var isGameOver = (gameState.GameStateEvent.Name == "gameOver")
	// sequential processing
	for playerToken, playerChannel := range manager.subs {
		gameState.PlayerId = manager.playerTokenToID[playerToken]
		gameState.PlayerToken = playerToken
		gameState.PlayerName = gameState.PlayerNames[gameState.PlayerId]
		gameState.PlayerIconId = gameState.PlayerIconIds[gameState.PlayerId]
		gameState.CardsOfPlayer.CardsInHand = manager.cardsInHands[gameState.PlayerId]
		for playerID, cards := range manager.cardsInHands {
			gameState.CardsOfPlayer.NrCardsOfOtherPlayers[playerID] = len(cards)
		}
		gameState.CardsOnTable = manager.CardsOnTable
		playerChannel <- *gameState // handled by goroutine in main.go
		if isGameOver {
			close(playerChannel)
		}
	}
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
	game.GameStateEvent.resetGameStateEvent()
	game.err = nil
}

func (readyEvent *ReadyEvent) triggerReadyEvent(triggedby int) {
	readyEvent.Name = "concentrate"
	readyEvent.TriggeredBy = triggedby
}
func (readyEvent *ReadyEvent) resetReadyEvent() {
	readyEvent.Name = ""
	readyEvent.TriggeredBy = 0
	readyEvent.Ready = make([]int, 0)
}
func (readyEvent *ReadyEvent) setReadyEventToCencentrate() {
	readyEvent.Name = "concentrate"
	readyEvent.TriggeredBy = 0
	readyEvent.Ready = make([]int, 0)
}
func (placeCardEvent *PlaceCardEvent) setPlaceCardEvent(name string, triggeredBy int, discardedCards *map[int][]int) {
	placeCardEvent.Name = name
	placeCardEvent.TriggeredBy = triggeredBy
	placeCardEvent.DiscardedCard = *discardedCards
	*discardedCards = make(map[int][]int)
}
func (processCardEvent *ProcessStarEvent) resetProcessStarEvent() {
	processCardEvent.Name = ""
	processCardEvent.TriggeredBy = 0
	processCardEvent.ProStar = make([]int, 0)
	processCardEvent.ConStar = make([]int, 0)
}
func (processCardEvent *ProcessStarEvent) triggerProcessStarEvent(triggedby int) {
	processCardEvent.Name = "proposeStar"
	processCardEvent.TriggeredBy = triggedby
	appendElementIfUnique(&processCardEvent.ProStar, triggedby)
	processCardEvent.ConStar = make([]int, 0)
}
func (gameStateEvent *GameStateEvent) resetGameStateEvent() {
	gameStateEvent.Name = ""
	gameStateEvent.LevelTitle = ""
	gameStateEvent.StarsIncrease = false
	gameStateEvent.LivesIncrease = false
	gameStateEvent.StarsDecrease = false
	gameStateEvent.LivesDecrease = false
}
