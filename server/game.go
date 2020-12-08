package main

import (
	"errors"
	"github.com/google/uuid"
	"log"
)

type subscription struct {
	playerId      string
	playerName    string
	playerChannel chan GameState
}

type Game struct {
	id        string
	inputCh   chan map[string]string
	publishCh chan GameState
	subCh     chan subscription
	unsubCh   chan subscription
}

func NewGame() *Game {
	return &Game{
		id:        uuid.New().String(), //concurrent reads only!
		inputCh:   make(chan map[string]string),
		publishCh: make(chan GameState, 1),
		subCh:     make(chan subscription, 1),
		unsubCh:   make(chan subscription, 1),
	}
}

type GameState struct {
	gameId         string
	playerIdToName map[string]string
	started        bool
	err            error
}

func describe(i interface{}) {
	log.Printf("(%v, %T)\n", i, i)
}

func (g *Game) Start() {
	// loop
	var subs = make(map[string]chan GameState)
	var players = make(map[string]string)
	var started = false
	var err error
	for {
		select {
		case raw := <-g.inputCh:
			describe(raw)
			g.publishCh <- GameState{
				g.id,
				players,
				started,
				err,
			}
		case subscriber := <-g.subCh:
			if len(players) >= 4 || started {
				err = errors.New("Cannot join game anymore")
				subscriber.playerChannel <- GameState{}
			} else {
				players[subscriber.playerId] = subscriber.playerName
			}
			subs[subscriber.playerId] = subscriber.playerChannel
		case subscriber := <-g.unsubCh:
			delete(subs, subscriber.playerId)
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
	playerId := uuid.New().String()
	playerChannel := make(chan GameState)
	g.subCh <- subscription{
		playerId:      playerId,
		playerName:    playerName,
		playerChannel: playerChannel,
	}
	return playerId, playerChannel
}

func (g *Game) Unsubscribe(playerId string, playerName string, playerChannel chan GameState) {
	g.unsubCh <- subscription{
		playerId:      playerId,
		playerName:    playerName,
		playerChannel: playerChannel,
	}
}

func (g *Game) PublishState(gameState GameState) {
	g.publishCh <- gameState
}
