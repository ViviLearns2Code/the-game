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
		unsubCh:   make(chan subscription, 1),
	}
}

func describe(i interface{}) {
	log.Printf("(%v, %T)\n", i, i)
}

func cardsInHandOfPlayer(playerIdx int, cardsInHands map[int][]int) (cardsOfPlayer CardsOfPlayer) {
	cardsOfPlayer.CardsInHand = cardsInHands[playerIdx]
	for playerId, cards := range cardsInHands {
		cardsOfPlayer.NrCardsOfOtherPlayers[playerId] = cap(cards)
	}
	return cardsOfPlayer
}
func (g *Game) Start() {
	// loop
	var subs = make(map[string]chan GameState)
	var players = make(map[int]string)
	var cardsInHands = make(map[int][]int)
	cardsOnTable := CardsOnTable{0, 0, 0, 0}
	var err *gameError = nil
	playerIdx := 1
	for {
		select {
		case raw := <-g.inputCh:
			g.publishCh <- GameState{
				GameToken:      g.token,
				PlayerToken:    raw.PlayerToken,
				PlayerName:     raw.PlayerName,
				PlayerId:       playerIdx,
				CardsOfPlayer:  cardsInHandOfPlayer(playerIdx, cardsInHands),
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
				PlayerNames: players,
				err:         err,
			}
		case subscriber := <-g.subCh:
			if len(players) >= 4 {
				var err = NewGameError("error", "cannot join game anymore")
				subscriber.playerChannel <- GameState{
					err: err,
				}
				return
			} else {
				players[playerIdx] = subscriber.playerName
				playerIdx++
			}
			subs[subscriber.playerToken] = subscriber.playerChannel
		case subscriber := <-g.unsubCh:
			delete(subs, subscriber.playerToken)
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

func (g *Game) Unsubscribe(playerToken string, playerName string, playerChannel chan GameState) {
	g.unsubCh <- subscription{
		playerToken:   playerToken,
		playerName:    playerName,
		playerChannel: playerChannel,
	}
}

func (g *Game) PublishState(gameState GameState) {
	g.publishCh <- gameState
}
