package main

import (
	"fmt"
)

// public for automatic test
type InputDetails struct {
	GameToken   string `json:"gameToken, omitempty"`
	PlayerToken string `json:"playerToken, omitempty"`
	PlayerName  string `json:"playerName, omitempty"`
	ActionId    string `json:"actionId, omitempty"`
	CardId      int    `json:"cardId, omitempty"`
}

type GameOutput struct {
	ErrorMsg  string     `json:"errorMsg"`
	GameState *GameState `json:"gameState, omitempty"`
}

// structs for game core
type gameError struct {
	severity string
	msg      string
}

func NewGameError(severity string, msg string) *gameError {
	return &gameError{
		severity: severity,
		msg:      msg,
	}
}
func (e *gameError) Error() string {
	var info = fmt.Sprintf("[%s] %s", e.severity, e.msg)
	return info
}

type subscription struct {
	playerToken   string
	playerName    string
	playerChannel chan GameState
}

type Game struct {
	token     string
	inputCh   chan InputDetails
	publishCh chan GameState
	subCh     chan subscription
	unsubCh   chan subscription
}

type CardsOnTable struct {
	TopCard int `json:"topCard, omitempty"`
	Level   int `json:"level, omitempty"`
	Lives   int `json:"lives, omitempty"`
	Stars   int `json:"stars, omitempty"`
}

type CardsOfPlayer struct {
	CardsInHand           []int       `json:"cardsInHand, omitempty"`
	NrCardsOfOtherPlayers map[int]int `json:"nrCardOfOtherPlayers, omitempty"`
}

type ReadyEvent struct {
	// lobby, concentrate
	Name string `json:"name, omitempty"`
	// playerId, 0 if lobby
	TriggeredBy int `json:"triggeredBy, omitempty"`
	// playerId
	Ready []int `json:"ready, omitempty"`
}

type PlaceCardEvent struct {
	// placeCard, useStar
	Name string `json:"name, omitempty"`
	// playerId, 0 if useStar
	TriggeredBy int `json:"triggeredBy, omitempty"`
	// playerId to card number
	DiscardedCard map[int]int `json:"discardedCard, omitempty"`
}

type ProcessStarEvent struct {
	// proposeStar, agreeStar, rejectStar
	Name        string `json:"name, omitempty"`
	TriggeredBy int    `json:"triggeredBy, omitempty"`
	ProStar     []int  `json:"proStar, omitempty"`
	ConStar     []int  `json:"conStar, omitempty"`
}

type GameStateEvent struct {
	// gameOver, lostLife, levelUp
	Name          string `json:"name, omitempty"`
	LevelTitle    string `json:"levelTitle, omitempty"`
	StarsIncrease bool   `json:"starsIncrease, omitempty"`
	StarsDecrease bool   `json:"starsDecrease, omitempty"`
	LivesIncrease bool   `json:"livesIncrease, omitempty"`
	LivesDecrease bool   `json:"livesDecrease, omitempty"`
}

type GameState struct {
	GameToken     string         `json:"gameToken"`
	PlayerToken   string         `json:"playerToken"`
	PlayerName    string         `json:"playerName"`
	PlayerId      int            `json:"PlayerId"`
	CardsOfPlayer CardsOfPlayer  `json:"cardsOfPlayer, omitempty"`
	PlayerNames   map[int]string `json:"playerNames, omitempty"`
	CardsOnTable  CardsOnTable   `json:"cardsOnTable, omitempty"`
	// events
	GameStateEvent   `json:"gameStateEvent, omitempty"`
	ReadyEvent       `json:"readyEvent, omitempty"`
	PlaceCardEvent   `json:"placeCardEvent, omitempty"`
	ProcessStarEvent `json:"processStarEvent, omitempty"`
	err              *gameError
}
