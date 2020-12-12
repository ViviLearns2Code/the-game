package main

// public for automatic test
type InputDetails struct {
	GameToken   string `json:"gameToken, omitempty"`
	PlayerToken string `json:"playerToken, omitempty"`
	PlayerName  string `json:"playerName, omitempty"`
	ActionId    string `json:"actionId, omitempty"`
	CardId      int    `json:"cardId, omitempty"`
}

type GameStateOutput struct {
	TopCard int   `json:"topCard, omitempty"`
	Level   int   `json:"level, omitempty"`
	Lives   int   `json:"lives, omitempty"`
	Stars   int   `json:"stars, omitempty"`
	Hand    []int `json:"hand, omitempty"`
}

type GameOutput struct {
	GameToken   string           `json:"gameId"`
	PlayerToken string           `json:"playerId"`
	PlayerName  string           `json:"playerName"`
	GameState   *GameStateOutput `json:"gameState, omitempty"`
}

// structs for game core
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
	topCard int
	level   int
	lives   int
	stars   int
}

type CardsOfPlayer struct{
	cardsInHand []int
	nrCardsOfOtherPlayers map[int]int
}

type ReadyEvents struct {
	name string  // Lobby, Concentrate
	triggeredBy int // playerId, 0 iff Lobby
	ready []int // playerId
}

type PlaceCardEvents struct {
	name string // PlacedCards, UsedStar
	triggeredBy int // playerId, 0 iff UsedStar
	discardedCard map[int]int // playerId to card number
}

type ProcessingStarEvent struct {
	name string  // ProposeStar, AgreeStar, RejectStar
	triggeredBy int
	proStar []int
	conStar []int
}

type GameStateEvent struct {
	name string // GameOver, LifeLost, LevelUp
	levelTitle string
	starsIncrease bool
	starsDecrease bool
	livesIncrease bool
	livesDecrease bool
}

type GameState struct {
	gameToken   string
	playerToken string
	playerName  string
	playerId int
	CardsOfPlayer
	playerNames map[int]string
	CardsOnTable
	// events
	GameStateEvent
	ReadyEvents
	PlaceCardEvents
	ProcessingStarEvent
	err error
}