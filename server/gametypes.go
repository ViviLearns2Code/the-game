package main

import (
	"fmt"
)

// public for automatic test
type InputDetails struct {
	GameToken    string `json:"gameToken"`
	PlayerToken  string `json:"playerToken"`
	PlayerName   string `json:"playerName"`
	PlayerIconId int    `json:"playerIconId"`
	ActionId     string `json:"actionId"`
	CardId       int    `json:"cardId"`
}

type GameOutput struct {
	ErrorMsg  string     `json:"errorMsg"`
	GameState *GameState `json:"gameState"`
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
	playerIconId  int
	playerChannel chan GameState
}

type LevelCard struct {
	levelTitle  string
	lifeAsBonus bool
	starAsBonus bool
}

type CardsManager struct {
	cardsInHands map[int][]int
	CardsOnTable
	levelCards     map[int]LevelCard
	discardedCards map[int][]int
}

type GameManager struct {
	playerTokenToID map[string]int            // token to int
	subs            map[string]chan GameState // token to channel
	started         bool
	CardsManager
}

type Game struct {
	token     string
	inputCh   chan InputDetails
	publishCh chan bool
	subCh     chan subscription
	unsubCh   chan string
}

type CardsOnTable struct {
	TopCard int `json:"topCard"`
	Level   int `json:"level"`
	Lives   int `json:"lives"`
	Stars   int `json:"stars"`
}

type CardsOfPlayer struct {
	CardsInHand           []int       `json:"cardsInHand"`
	NrCardsOfOtherPlayers map[int]int `json:"nrCardOfOtherPlayers"`
}

type ReadyEvent struct {
	// lobby, concentrate
	Name string `json:"name"`
	// playerId, 0 if lobby
	TriggeredBy int `json:"triggeredBy"`
	// playerId
	Ready []int `json:"ready"`
}

type PlaceCardEvent struct {
	// placeCard, useStar
	Name string `json:"name"`
	// playerId, 0 if useStar
	TriggeredBy int `json:"triggeredBy"`
	// playerId to card number
	DiscardedCard map[int][]int `json:"discardedCard"`
}

type ProcessStarEvent struct {
	// proposeStar, agreeStar, rejectStar
	Name        string `json:"name"`
	TriggeredBy int    `json:"triggeredBy"`
	ProStar     []int  `json:"proStar"`
	ConStar     []int  `json:"conStar"`
}

type GameStateEvent struct {
	// gameOver, lostLife, levelUp
	Name          string `json:"name"`
	LevelTitle    string `json:"levelTitle"`
	StarsIncrease bool   `json:"starsIncrease"`
	StarsDecrease bool   `json:"starsDecrease"`
	LivesIncrease bool   `json:"livesIncrease"`
	LivesDecrease bool   `json:"livesDecrease"`
}

// GameState per Player
type GameState struct {
	GameToken     string         `json:"gameToken"`
	PlayerToken   string         `json:"playerToken"`
	PlayerName    string         `json:"playerName"`
	PlayerId      int            `json:"PlayerId"`
	PlayerIconId  int            `json:"PlayerIconId"`
	CardsOfPlayer CardsOfPlayer  `json:"cardsOfPlayer"`
	PlayerIconIds map[int]int    `json:"playerIconIds"`
	PlayerNames   map[int]string `json:"playerNames"`
	CardsOnTable  CardsOnTable   `json:"cardsOnTable"`
	// events
	GameStateEvent   `json:"gameStateEvent"`
	ReadyEvent       `json:"readyEvent"`
	PlaceCardEvent   `json:"placeCardEvent"`
	ProcessStarEvent `json:"processStarEvent"`
	err              *gameError
}
