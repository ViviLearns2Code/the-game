package main

import (
	"errors"
	"github.com/google/uuid"
	"log"
)

type Game struct {
	id        string
	inputCh   chan map[string]string
	publishCh chan GameState
	subCh     chan chan GameState
	unsubCh   chan chan GameState
	players   map[string]string
	started   bool
}

func NewGame() *Game {
	return &Game{
		id:        uuid.New().String(),
		inputCh:   make(chan map[string]string),
		publishCh: make(chan GameState, 1),
		subCh:     make(chan chan GameState, 1),
		unsubCh:   make(chan chan GameState, 1),
		players:   make(map[string]string),
		started:   false,
	}
}

type GameState struct {
	gameId         string
	playerIdToName map[string]string
}

func describe(i interface{}) {
	log.Printf("(%v, %T)\n", i, i)
}

func (g *Game) Start() {
	// loop
	var subs = make(map[chan GameState]struct{})
	for {
		select {
		case raw := <-g.inputCh:
			describe(raw)
			g.publishCh <- GameState{
				g.id,
				g.players,
			}
		case msgCh := <-g.subCh:
			subs[msgCh] = struct{}{}
		case msgCh := <-g.unsubCh:
			delete(subs, msgCh)
		case msg := <-g.publishCh:
			log.Printf("New game state published")
			for msgCh := range subs {
				// msgCh is buffered, use non-blocking send to protect the broker:
				select {
				case msgCh <- msg:
					// handled by goroutine in main.go
				default:
				}
			}
		}
	}
}

func (g *Game) Subscribe(playerName string) (string, chan GameState, error) {
	playerId := uuid.New().String()
	playerChannel := make(chan GameState)
	var err error
	g.subCh <- playerChannel
	if len(g.players) >= 4 || g.started {
		err = errors.New("Cannot join game anymore")
	} else {
		g.players[playerId] = playerName
	}
	return playerId, playerChannel, err
}

func (g *Game) Unsubscribe(playerChannel chan GameState) {
	g.unsubCh <- playerChannel
}

func (g *Game) PublishState(gameState GameState) {
	g.publishCh <- gameState
}
