package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func testReadyEventResult(t *testing.T, gameState GameState, nrOfPlayer int, expectedReadyEvent ReadyEvent) {
	assert.Equal(t, len(gameState.PlayerNames), 2)
	assert.Equal(t, gameState.ReadyEvent.Name, expectedReadyEvent.Name)
	assert.Equal(t, len(gameState.ReadyEvent.Ready), len(expectedReadyEvent.Ready))
	assert.Equal(t, gameState.ReadyEvent.TriggeredBy, expectedReadyEvent.TriggeredBy)
}

func testCardsInHandsAndOnTable(t *testing.T, gameState GameState, nrOfCard int, cardOnTop int, level int, lives int, stars int) {
	assert.Equal(t, len(gameState.CardsOfPlayer.CardsInHand), nrOfCard)
	assert.Equal(t, gameState.CardsOnTable.TopCard, cardOnTop)
	assert.Equal(t, gameState.CardsOnTable.Level, level)
	assert.Equal(t, gameState.CardsOnTable.Lives, lives)
	assert.Equal(t, gameState.CardsOnTable.Stars, stars)
}

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
	expectedReadyEvent = ReadyEvent{"lobby", 0, []int{bobID}}
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
	expectedReadyEvent = ReadyEvent{"lobby", 0, []int{bobID}}
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

	expectedGameState := GameStateEvent{"levelUp", "Erhöhte Sensibilitat", false, false, false, false}
	expectedReadyEvent = ReadyEvent{"concentrate", 0, make([]int, 0)}
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
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{bobID}}
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
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{bobID, maryID}}
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

	// wrong card is placed
	// the cards smaller than placed are placed
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
	// lost one life
	// but the level is up, everyone has 2 cards in hand
	// concentrate
	expectedPlaceCardEvent := PlaceCardEvent{"placeCard", triggeredBy, map[int][]int{1: {cardOfMary}, 2: {cardOfBob}}}
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

	// bob and mary are ready for level 2
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	for i := 0; i < 2; i++ {
		select {
		case <-maryChannel:
		case <-bobChannel:
		}
	}
	var cardsOfMary []int
	var cardsOfBob []int
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			cardsOfMary = c1.CardsOfPlayer.CardsInHand
		case c2 := <-bobChannel:
			cardsOfBob = c2.CardsOfPlayer.CardsInHand
		}
	}
	// the correct card is played
	// check cards on table and in hands
	if cardsOfMary[0] < cardsOfBob[0] {
		maryInput.ActionId = "card"
		maryInput.CardId = cardsOfMary[0]
		myGame.inputCh <- *maryInput
		expectedPlaceCardEvent.TriggeredBy = maryID
		expectedPlaceCardEvent.DiscardedCard = map[int][]int{maryID: {maryInput.CardId}}
	} else {
		bobInput.ActionId = "card"
		bobInput.CardId = cardsOfBob[0]
		myGame.inputCh <- *bobInput
		expectedPlaceCardEvent.TriggeredBy = bobID
		expectedPlaceCardEvent.DiscardedCard = map[int][]int{bobID: {bobInput.CardId}}
	}

	nrOfCardsMary := 2
	if expectedPlaceCardEvent.TriggeredBy == maryID {
		nrOfCardsMary--
	}
	topCard := expectedPlaceCardEvent.DiscardedCard[expectedPlaceCardEvent.TriggeredBy][0]
	nrOfCardsBob := 3 - nrOfCardsMary
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testCardsInHandsAndOnTable(t, c1, nrOfCardsMary, topCard, 2, 1, 1)
			assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
		case c2 := <-bobChannel:
			testCardsInHandsAndOnTable(t, c2, nrOfCardsBob, topCard, 2, 1, 1)
			assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
		}
	}

	// concentration triggered by mary
	maryInput.CardId = 0
	maryInput.ActionId = "concentrate"
	myGame.inputCh <- *maryInput
	expectedReadyEvent = ReadyEvent{"concentrate", maryID, make([]int, 0)}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}
	bobInput.CardId = 0
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	expectedReadyEvent = ReadyEvent{"concentrate", maryID, []int{bobID}}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}
	maryInput.CardId = 0
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput
	expectedReadyEvent = ReadyEvent{"concentrate", maryID, []int{bobID, maryID}}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}
	// prposal star by bob
	bobInput.ActionId = "propose-star"
	myGame.inputCh <- *bobInput
	expectedProcessStarEvent := ProcessStarEvent{"proposeStar", bobID, []int{bobID}, make([]int, 0)}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			assert.Equal(t, c1.ProcessStarEvent, expectedProcessStarEvent)
		case c2 := <-bobChannel:
			assert.Equal(t, c2.ProcessStarEvent, expectedProcessStarEvent)
		}
	}
	// rejected by mary
	maryInput.ActionId = "reject-star"
	myGame.inputCh <- *maryInput
	expectedProcessStarEvent = ProcessStarEvent{"rejectStar", bobID, []int{bobID}, []int{maryID}}
	expectedReadyEvent = ReadyEvent{"concentrate", 0, make([]int, 0)}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			assert.Equal(t, c1.ProcessStarEvent, expectedProcessStarEvent)
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
			assert.Equal(t, c2.ProcessStarEvent, expectedProcessStarEvent)
		}
	}
	// constratation after star
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{maryID}}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{maryID, bobID}}
	for i := 0; i < 2; i++ {
		select {
		case c1 := <-maryChannel:
			testReadyEventResult(t, c1, 2, expectedReadyEvent)
		case c2 := <-bobChannel:
			testReadyEventResult(t, c2, 2, expectedReadyEvent)
		}
	}
	// place wrong card, game over
	if cardsOfMary[2-nrOfCardsMary] > cardsOfBob[2-nrOfCardsBob] {
		maryInput.ActionId = "card"
		maryInput.CardId = cardsOfMary[2-nrOfCardsMary]
		myGame.inputCh <- *maryInput
		expectedPlaceCardEvent.TriggeredBy = maryID
		expectedPlaceCardEvent.DiscardedCard = map[int][]int{maryID: {cardsOfMary[2-nrOfCardsMary]}}
		for n := 2 - nrOfCardsBob; n < 2; n++ {
			if cardsOfBob[n] < maryInput.CardId {
				expectedPlaceCardEvent.DiscardedCard[bobID] = append(expectedPlaceCardEvent.DiscardedCard[bobID], cardsOfBob[n])
			} else {
				break
			}
		}
	} else {
		bobInput.ActionId = "card"
		bobInput.CardId = cardsOfBob[2-nrOfCardsBob]
		myGame.inputCh <- *bobInput
		expectedPlaceCardEvent.TriggeredBy = bobID
		expectedPlaceCardEvent.DiscardedCard = map[int][]int{bobID: {cardsOfBob[2-nrOfCardsBob]}}
		for n := 2 - nrOfCardsMary; n < 2; n++ {
			if cardsOfMary[n] < bobInput.CardId {
				expectedPlaceCardEvent.DiscardedCard[maryID] = append(expectedPlaceCardEvent.DiscardedCard[maryID], cardsOfMary[n])
			} else {
				break
			}
		}
	}

	expectedGameState = GameStateEvent{"gameOver", "", false, false, false, true}
	nrOfCardsMary = nrOfCardsMary - len(expectedPlaceCardEvent.DiscardedCard[maryID])
	nrOfCardsBob = nrOfCardsBob - len(expectedPlaceCardEvent.DiscardedCard[bobID])
	c1, ok1 := <-maryChannel
	assert.True(t, ok1)
	testCardsInHandsAndOnTable(t, c1, nrOfCardsMary, topCard, 2, 0, 1)
	assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c1.GameStateEvent, expectedGameState)
	c2, ok2 := <-bobChannel
	assert.True(t, ok2)
	testCardsInHandsAndOnTable(t, c2, nrOfCardsBob, topCard, 2, 0, 1)
	assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c2.GameStateEvent, expectedGameState)
}
