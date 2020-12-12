package main

import (
	"errors"
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

func cardsInHandOfPlayer(playerIdx int, cardsInHands map[int][]int) (cardsOfPlayer CardsOfPlayer){
 cardsOfPlayer.cardsInHand=cardsInHands[playerIdx]
 for playerId, cards := range cardsInHands{
	 cardsOfPlayer.nrCardsOfOtherPlayers[playerId] = cap(cards)
 }
 return cardsOfPlayer
}
func (g *Game) Start() {
	// loop
	var subs = make(map[string]chan GameState)
	var players = make(map[int]string)
	var cardsInHands = make(map[int][]int)
	cardsOnTable := CardsOnTable{0, 0,0,0}
	var err error
	playerIdx := 1
	for {
		select {
		case raw := <-g.inputCh:
			describe(raw)
			g.publishCh <- GameState{
				gameToken:   g.token,
				playerToken: raw.PlayerToken,
				playerName:  raw.PlayerName,
				playerId: playerIdx,
				CardsOfPlayer: cardsInHandOfPlayer(playerIdx, cardsInHands),
				CardsOnTable : cardsOnTable,
				GameStateEvent: GameStateEvent{"", "", false, false,false,false},
				ReadyEvents: ReadyEvents{"", 0,nil},
				PlaceCardEvents: PlaceCardEvents{
					name:          "",
					triggeredBy:   0,
					discardedCard: nil,
				},
				
				ProcessingStarEvent:ProcessingStarEvent{
					name:        "",
					triggeredBy: 0,
					proStar:     nil,
					conStar:     nil,
				},
				playerNames: players,
				err:         err,
			}
			playerIdx++
		case subscriber := <-g.subCh:
			if len(players) >= 4 {
				err = errors.New("cannot join game anymore")
				subscriber.playerChannel <- GameState{}
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
