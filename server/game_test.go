package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestStart(t *testing.T) {

	var myGame Game
	myGame = *NewGame()
	go myGame.Start()

	var maryInput = &InputDetails{
		PlayerName: "mary",
		ActionId:   "create",
	}
	maryToken, maryChannel := myGame.Subscribe("mary")
	maryInput.PlayerToken = maryToken
	maryInput.GameToken = myGame.token
	maryID := 1
	// mary create room
	// tet result only 1 player in lobby
	maryGameState := <-maryChannel
	assert.Equal(t, len(maryGameState.PlayerNames), 1)
	assert.Equal(t, maryGameState.ReadyEvent.Name, "lobby")

	var bobInput = &InputDetails{
		PlayerName: "bob",
		ActionId:   "join",
	}
	bobToken, bobChannel := myGame.Subscribe("bob")
	bobInput.PlayerToken = bobToken
	bobInput.GameToken = myGame.token
	bobID := 2
	// bob joint the room
	// test result, 2 players are in lobby
	var testReadyEventResult = func(t *testing.T, gameState GameState, nrOfPlayer int, expectedReadyEvent ReadyEvent) {
		assert.Equal(t, len(gameState.PlayerNames), 2)
		assert.Equal(t, gameState.ReadyEvent.Name, expectedReadyEvent.Name)
		assert.Equal(t, len(gameState.ReadyEvent.Ready), len(expectedReadyEvent.Ready))
		assert.Equal(t, gameState.ReadyEvent.TriggeredBy, expectedReadyEvent.TriggeredBy)
	}
	expectedReadyEvent := ReadyEvent{"lobby", 0, make([]int, 0)}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}

	bobInput.ActionId = "start"
	myGame.inputCh <- *bobInput

	// bob is ready for start
	expectedReadyEvent.Ready = append(expectedReadyEvent.Ready, bobID)
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}

	maryInput.ActionId = "start"
	myGame.inputCh <- *maryInput

	// mary is ready for start --> game starts
	// bob and mary has 1 card in hand, level 1, lives 2, stars1
	// GameStateEvent, levelup with the titel
	// the players need cencentration
	var testCardsInHandsAndOnTable = func(t *testing.T, gameState GameState, nrOfCard int, cardOnTop int, level int, lives int, stars int) {
		assert.Equal(t, len(gameState.CardsOfPlayer.CardsInHand), nrOfCard)
		assert.Equal(t, gameState.CardsOnTable.TopCard, cardOnTop)
		assert.Equal(t, gameState.CardsOnTable.Level, level)
		assert.Equal(t, gameState.CardsOnTable.Lives, lives)
		assert.Equal(t, gameState.CardsOnTable.Stars, stars)
	}
	expectedGameState := GameStateEvent{"levelUp", "Erhöhte Sensibilitat", false, false, false, false}
	expectedReadyEvent.Name = "concentrate"
	expectedReadyEvent.Ready = make([]int, 0)
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testCardsInHandsAndOnTable(t, c1, 1, 0, 1, 2, 1)
			assert.Equal(t, c1.GameStateEvent, expectedGameState)
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testCardsInHandsAndOnTable(t, c2, 1, 0, 1, 2, 1)
			assert.Equal(t, c2.GameStateEvent, expectedGameState)
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}

	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput

	// bob is ready for play
	// GameStateEvent is cleaned up
	expectedGameState.Name = ""
	expectedGameState.LevelTitle = ""
	expectedReadyEvent.Ready = append(expectedReadyEvent.Ready, bobID)
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			assert.Equal(t, c1.GameStateEvent, expectedGameState)
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			assert.Equal(t, c2.GameStateEvent, expectedGameState)
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}

	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput

	// mary is ready for play --> all are ready
	expectedReadyEvent.Ready = append(expectedReadyEvent.Ready, 1)
	var cardOfMary, cardOfBob int
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
			cardOfMary = c1.CardsOfPlayer.CardsInHand[0]
		case c2 := <-bobChannel:
			cardOfBob = c2.CardsOfPlayer.CardsInHand[0]
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}

	var triggeredBy int
	if cardOfMary > cardOfBob {
		maryInput.ActionId = "card"
		maryInput.CardId = cardOfMary
		myGame.inputCh <- *maryInput
		triggeredBy = maryID
	} else {
		bobInput.ActionId = "card"
		bobInput.CardId = cardOfBob
		myGame.inputCh <- *bobInput
		triggeredBy = bobID
	}

	// wrong card is placed
	// the cards smaller than placed are placed
	// lost one life
	// but the level is up, everyone has 2 cards in hand
	// concentrate
	expectedPlaceCardEvent := PlaceCardEvent{"placeCard", triggeredBy, map[int]int{1: cardOfMary, 2: cardOfBob}}
	expectedGameState = GameStateEvent{"levelUp", "Verstärkte Empathie", false, false, false, true} // lost life
	expectedReadyEvent = ReadyEvent{"concentrate", 0, make([]int, 0)}

	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testCardsInHandsAndOnTable(t, c1, 2, 0, 2, 1, 1)
			assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
			assert.Equal(t, c1.GameStateEvent, expectedGameState)
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testCardsInHandsAndOnTable(t, c2, 2, 0, 2, 1, 1)
			assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
			assert.Equal(t, c2.GameStateEvent, expectedGameState)
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}
}
