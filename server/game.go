package main

import (
	"log"

	"github.com/google/uuid"
)

func NewGame() *Game {
	return &Game{
		token:     uuid.New().String(), //concurrent reads only!
		inputCh:   make(chan InputDetails),
		publishCh: make(chan GameState, 1),
		subCh:     make(chan subscription, 1),
		unsubCh:   make(chan string, 1),
	}
}

func describe(i interface{}) {
	log.Printf("(%v, %T)\n", i, i)
}

func cardsInHandOfPlayer(playerId int, cardsInHands map[int][]int) (cardsOfPlayer CardsOfPlayer) {
	cardsOfPlayer.CardsInHand = cardsInHands[playerId]
	for playerId, cards := range cardsInHands {
		cardsOfPlayer.NrCardsOfOtherPlayers[playerId] = cap(cards)
	}
	return cardsOfPlayer
}
func (g *Game) Start() {
	// loop
	var subs = make(map[string]chan GameState)
	var playerId2Name = make(map[int]string)
	var playerToken2Id = make(map[string]int)
	var cardsInHands = make(map[int][]int)
	var cardsOnTable = CardsOnTable{0, 0, 0, 0}
	var err *gameError = nil
	var nextPlayerId = 1
	for {
		select {
		case inputDetails := <-g.inputCh:
			g.publishCh <- GameState{
				GameToken:      g.token,
				PlayerToken:    inputDetails.PlayerToken,
				PlayerName:     inputDetails.PlayerName,
				PlayerId:       playerToken2Id[inputDetails.PlayerToken],
				CardsOfPlayer:  cardsInHandOfPlayer(playerToken2Id[inputDetails.PlayerToken], cardsInHands),
				CardsOnTable:   cardsOnTable,
				GameStateEvent: GameStateEvent{"", "", false, false, false, false},
				ReadyEvent:     ReadyEvent{"", 0, nil},
				PlaceCardEvent: PlaceCardEvent{
					Name:          "",
					TriggeredBy:   0,
					DiscardedCard: nil,
				},

				ProcessStarEvent: ProcessStarEvent{
					Name:        "",
					TriggeredBy: 0,
					ProStar:     nil,
					ConStar:     nil,
				},
				PlayerNames: playerId2Name,
				err:         err,
			}
		case subscriber := <-g.subCh:
			if len(playerId2Name) >= 4 {
				var err = NewGameError("error", "cannot join game anymore")
				subscriber.playerChannel <- GameState{
					err: err,
				}
				return
			} else {
				playerId2Name[nextPlayerId] = subscriber.playerName
				playerToken2Id[subscriber.playerToken] = nextPlayerId
				nextPlayerId++
			}
			subs[subscriber.playerToken] = subscriber.playerChannel
		case <-g.unsubCh:
			//trigger gameOver
		case gameState := <-g.publishCh:
			log.Printf("New game state published")
			for _, playerChannel := range subs {
				select {
				case playerChannel <- gameState:
					// handled by goroutine in main.go
				default:
				}
			}
		}
	}
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

func (g *Game) PublishState(gameState GameState) {
	g.publishCh <- gameState
}
