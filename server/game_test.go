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
	assert.Equal(t, gameState.CardsOfPlayer.NrCardsOfOtherPlayers[gameState.PlayerId], nrOfCard)
	assert.Equal(t, gameState.CardsOnTable.TopCard, cardOnTop)
	assert.Equal(t, gameState.CardsOnTable.Level, level)
	assert.Equal(t, gameState.CardsOnTable.Lives, lives)
	assert.Equal(t, gameState.CardsOnTable.Stars, stars)
}

func TestUseStar(t *testing.T) {
	var myGame Game
	myGame = *NewGame()
	go myGame.Start(false)

	var maryInput = &InputDetails{
		PlayerName: "mary",
		ActionId:   "create",
	}
	maryToken, maryChannel := myGame.Subscribe("mary", 1)
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
	bobToken, bobChannel := myGame.Subscribe("bob", 2)
	bobInput.PlayerToken = bobToken
	bobInput.GameToken = myGame.token
	bobID := 2
	// bob joint the room
	// test result, 2 players are in lobby

	expectedReadyEvent := ReadyEvent{"lobby", 0, make([]int, 0)}
	c1, c2 := <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)

	bobInput.ActionId = "start"
	myGame.inputCh <- *bobInput

	// bob is ready for start
	expectedReadyEvent = ReadyEvent{"lobby", 0, []int{bobID}}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	bobInput.ActionId = "start"
	myGame.inputCh <- *bobInput

	// bob is ready for start
	expectedReadyEvent = ReadyEvent{"lobby", 0, []int{bobID}}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)

	maryInput.ActionId = "start"
	myGame.inputCh <- *maryInput

	// mary is ready for start --> game starts
	// bob and mary has 1 card in hand, level 1, lives 2, stars1
	// GameStateEvent, levelup with the titel
	// the players need cencentration

	expectedGameState := GameStateEvent{"levelUp", "Initialization: Game Trek", false, false, false, false}
	expectedReadyEvent = ReadyEvent{"concentrate", 0, make([]int, 0)}
	c1, c2 = <-maryChannel, <-bobChannel
	testCardsInHandsAndOnTable(t, c1, 1, 0, 1, 2, 1)
	assert.Equal(t, c1.GameStateEvent, expectedGameState)
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testCardsInHandsAndOnTable(t, c2, 1, 0, 1, 2, 1)
	assert.Equal(t, c2.GameStateEvent, expectedGameState)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)

	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput

	// bob is ready for play
	// GameStateEvent is cleaned up
	expectedGameState.Name = ""
	expectedGameState.LevelTitle = ""
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{bobID}}
	c1, c2 = <-maryChannel, <-bobChannel

	assert.Equal(t, c1.GameStateEvent, expectedGameState)
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	assert.Equal(t, c2.GameStateEvent, expectedGameState)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)

	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput

	// mary is ready for play --> all are ready
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{bobID, maryID}}
	var cardOfMary, cardOfBob int
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	cardOfMary = c1.CardsOfPlayer.CardsInHand[0]
	cardOfBob = c2.CardsOfPlayer.CardsInHand[0]
	testReadyEventResult(t, c2, 2, expectedReadyEvent)

	//propose
	bobInput.ActionId = "propose-star"
	myGame.inputCh <- *bobInput
	expectedProcessStarEvent := ProcessStarEvent{"proposeStar", bobID, []int{bobID}, make([]int, 0)}
	c1, c2 = <-maryChannel, <-bobChannel
	assert.Equal(t, c1.ProcessStarEvent, expectedProcessStarEvent)
	assert.Equal(t, c2.ProcessStarEvent, expectedProcessStarEvent)
	// agree by mary
	maryInput.ActionId = "agree-star"
	myGame.inputCh <- *maryInput
	expectedProcessStarEvent = ProcessStarEvent{"agreeStar", bobID, []int{bobID, maryID}, make([]int, 0)}
	expectedReadyEvent = ReadyEvent{"concentrate", 0, make([]int, 0)}
	c1, c2 = <-maryChannel, <-bobChannel
	assert.Equal(t, c1.ProcessStarEvent, expectedProcessStarEvent)
	assert.Equal(t, c2.ProcessStarEvent, expectedProcessStarEvent)
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	expectedPlaceCardEvent := PlaceCardEvent{"useStar", 0, map[int][]int{1: {cardOfMary}, 2: {cardOfBob}}}
	assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
}

func TestStart(t *testing.T) {

	var myGame Game
	myGame = *NewGame()
	go myGame.Start(false)

	// mary create room
	// tet result only 1 player in lobby
	var maryInput = &InputDetails{
		PlayerName: "mary",
		ActionId:   "create",
	}
	maryToken, maryChannel := myGame.Subscribe("mary", 1)
	maryInput.PlayerToken = maryToken
	maryInput.GameToken = myGame.token
	maryID := 1
	maryGameState := <-maryChannel
	assert.Equal(t, len(maryGameState.PlayerNames), 1)
	assert.Equal(t, maryGameState.ReadyEvent.Name, "lobby")

	// bob joint the room
	// test result, 2 players are in lobby
	var bobInput = &InputDetails{
		PlayerName: "bob",
		ActionId:   "join",
	}
	bobToken, bobChannel := myGame.Subscribe("bob", 2)
	bobInput.PlayerToken = bobToken
	bobInput.GameToken = myGame.token
	bobID := 2
	<-maryChannel
	<-bobChannel

	// bob is ready for start

	bobInput.ActionId = "start"
	myGame.inputCh <- *bobInput
	<-maryChannel
	<-bobChannel
	// mary is ready for start
	// mary is ready for start --> game starts
	// bob and mary has 1 card in hand, level 1, lives 2, stars1
	// GameStateEvent, levelup with the titel
	// the players need cencentration
	maryInput.ActionId = "start"
	myGame.inputCh <- *maryInput
	<-maryChannel
	<-bobChannel

	// bob is ready for play
	// GameStateEvent is cleaned up
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	<-maryChannel
	<-bobChannel

	// mary is ready for play --> all are ready
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput

	var cardOfMary, cardOfBob int
	c1, c2 := <-maryChannel, <-bobChannel
	cardOfMary = c1.CardsOfPlayer.CardsInHand[0]
	cardOfBob = c2.CardsOfPlayer.CardsInHand[0]

	// wrong card is placed
	// the cards smaller than placed are placed
	expectedPlaceCardEvent := PlaceCardEvent{"placeCard", 0, make(map[int][]int)}
	if cardOfMary > cardOfBob {
		maryInput.ActionId = "card"
		maryInput.CardId = cardOfMary
		myGame.inputCh <- *maryInput
		expectedPlaceCardEvent.TriggeredBy = maryID
		expectedPlaceCardEvent.DiscardedCard[maryID] = []int{cardOfMary}
		expectedPlaceCardEvent.DiscardedCard[bobID] = []int{cardOfBob}
	} else {
		bobInput.ActionId = "card"
		bobInput.CardId = cardOfBob
		myGame.inputCh <- *bobInput
		expectedPlaceCardEvent.TriggeredBy = bobID
		expectedPlaceCardEvent.DiscardedCard[bobID] = []int{cardOfBob}
		expectedPlaceCardEvent.DiscardedCard[maryID] = []int{cardOfMary}
	}
	// lost one life
	// but the level is up, everyone has 2 cards in hand
	// concentrate

	expectedGameState := GameStateEvent{"levelUp", "Systems: online", false, false, false, true} // lost life
	expectedReadyEvent := ReadyEvent{"concentrate", 0, make([]int, 0)}

	c1, c2 = <-maryChannel, <-bobChannel

	testCardsInHandsAndOnTable(t, c1, 2, 0, 2, 1, 1)
	assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c1.GameStateEvent, expectedGameState)
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testCardsInHandsAndOnTable(t, c2, 2, 0, 2, 1, 1)
	assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c2.GameStateEvent, expectedGameState)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	// bob and mary are ready for level 2
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	<-maryChannel
	<-bobChannel
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput
	c1, c2 = <-maryChannel, <-bobChannel
	var cardsOfMary []int
	var cardsOfBob []int
	cardsOfMary = c1.CardsOfPlayer.CardsInHand
	cardsOfBob = c2.CardsOfPlayer.CardsInHand
	nrOfCardsMary := len(cardsOfMary)
	nrOfCardsBob := len(cardsOfBob)
	expectedPlaceCardEvent.DiscardedCard = make(map[int][]int)
	topCard := 0
	// the correct card is played
	// check cards on table and in hands
	if cardsOfMary[0] < cardsOfBob[0] {
		maryInput.ActionId = "card"
		maryInput.CardId = cardsOfMary[0]
		topCard = maryInput.CardId
		myGame.inputCh <- *maryInput
		expectedPlaceCardEvent.TriggeredBy = maryID
		expectedPlaceCardEvent.DiscardedCard[maryID] = []int{cardsOfMary[0]}
		nrOfCardsMary--
	} else {
		bobInput.ActionId = "card"
		bobInput.CardId = cardsOfBob[0]
		topCard = bobInput.CardId
		myGame.inputCh <- *bobInput
		expectedPlaceCardEvent.TriggeredBy = bobID
		expectedPlaceCardEvent.DiscardedCard[bobID] = []int{cardsOfBob[0]}
		nrOfCardsBob--
	}
	c1, c2 = <-maryChannel, <-bobChannel
	testCardsInHandsAndOnTable(t, c1, nrOfCardsMary, topCard, 2, 1, 1)
	assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
	testCardsInHandsAndOnTable(t, c2, nrOfCardsBob, topCard, 2, 1, 1)
	assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
	// concentration triggered by mary
	maryInput.CardId = 0
	maryInput.ActionId = "concentrate"
	myGame.inputCh <- *maryInput
	expectedReadyEvent = ReadyEvent{"concentrate", maryID, make([]int, 0)}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)

	bobInput.CardId = 0
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	expectedReadyEvent = ReadyEvent{"concentrate", maryID, []int{bobID}}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	maryInput.CardId = 0
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput
	expectedReadyEvent = ReadyEvent{"concentrate", maryID, []int{bobID, maryID}}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	// prposal star by bob
	bobInput.ActionId = "propose-star"
	myGame.inputCh <- *bobInput
	expectedProcessStarEvent := ProcessStarEvent{"proposeStar", bobID, []int{bobID}, make([]int, 0)}
	c1, c2 = <-maryChannel, <-bobChannel
	assert.Equal(t, c1.ProcessStarEvent, expectedProcessStarEvent)
	assert.Equal(t, c2.ProcessStarEvent, expectedProcessStarEvent)
	// rejected by mary
	maryInput.ActionId = "reject-star"
	myGame.inputCh <- *maryInput
	expectedProcessStarEvent = ProcessStarEvent{"rejectStar", bobID, []int{bobID}, []int{maryID}}
	expectedReadyEvent = ReadyEvent{"concentrate", 0, make([]int, 0)}
	c1, c2 = <-maryChannel, <-bobChannel
	assert.Equal(t, c1.ProcessStarEvent, expectedProcessStarEvent)
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	assert.Equal(t, c2.ProcessStarEvent, expectedProcessStarEvent)
	// constratation after star
	maryInput.ActionId = "ready"
	myGame.inputCh <- *maryInput
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{maryID}}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	bobInput.ActionId = "ready"
	myGame.inputCh <- *bobInput
	expectedReadyEvent = ReadyEvent{"concentrate", 0, []int{maryID, bobID}}
	c1, c2 = <-maryChannel, <-bobChannel
	testReadyEventResult(t, c1, 2, expectedReadyEvent)
	testReadyEventResult(t, c2, 2, expectedReadyEvent)
	// place wrong card, game over
	expectedPlaceCardEvent.DiscardedCard = make(map[int][]int)
	if cardsOfMary[2-nrOfCardsMary] > cardsOfBob[2-nrOfCardsBob] {
		maryInput.ActionId = "card"
		maryInput.CardId = cardsOfMary[2-nrOfCardsMary]
		myGame.inputCh <- *maryInput
		expectedPlaceCardEvent.TriggeredBy = maryID
		topCard = cardsOfMary[2-nrOfCardsMary]
		expectedPlaceCardEvent.DiscardedCard[maryID] = []int{topCard}
		for n := 2 - nrOfCardsBob; n < 2; n++ {
			if cardsOfBob[n] < maryInput.CardId {
				expectedPlaceCardEvent.DiscardedCard[bobID] = append(expectedPlaceCardEvent.DiscardedCard[bobID], cardsOfBob[n])
			} else {
				break
			}
		}
		nrOfCardsMary--
		nrOfCardsBob = nrOfCardsBob - len(expectedPlaceCardEvent.DiscardedCard[bobID])
	} else {
		bobInput.ActionId = "card"
		bobInput.CardId = cardsOfBob[2-nrOfCardsBob]
		myGame.inputCh <- *bobInput
		expectedPlaceCardEvent.TriggeredBy = bobID
		topCard = cardsOfBob[2-nrOfCardsBob]
		expectedPlaceCardEvent.DiscardedCard[bobID] = []int{topCard}
		for n := 2 - nrOfCardsMary; n < 2; n++ {
			if cardsOfMary[n] < bobInput.CardId {
				expectedPlaceCardEvent.DiscardedCard[maryID] = append(expectedPlaceCardEvent.DiscardedCard[maryID], cardsOfMary[n])
			} else {
				break
			}
		}
		nrOfCardsBob--
		nrOfCardsMary = nrOfCardsMary - len(expectedPlaceCardEvent.DiscardedCard[maryID])
	}

	expectedGameState = GameStateEvent{"gameOver", "", false, false, false, true}
	c1, ok1 := <-maryChannel
	assert.Equal(t, ok1, true)
	testCardsInHandsAndOnTable(t, c1, nrOfCardsMary, topCard, 2, 0, 1)
	assert.Equal(t, c1.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c1.GameStateEvent, expectedGameState)
	c2, ok2 := <-bobChannel
	assert.Equal(t, ok2, true)
	testCardsInHandsAndOnTable(t, c2, nrOfCardsBob, topCard, 2, 0, 1)
	assert.Equal(t, c2.PlaceCardEvent, expectedPlaceCardEvent)
	assert.Equal(t, c2.GameStateEvent, expectedGameState)
}
