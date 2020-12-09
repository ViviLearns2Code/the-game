package main

// structs for API layer
type inputDetails struct {
	gameId     string
	playerId   string
	playerName string
	actionId   string
	cardId     int
}

type GameStateOutput struct {
	TopCard int   `json:"topCard, omitempty"`
	Level   int   `json:"level, omitempty"`
	Lives   int   `json:"lives, omitempty"`
	Stars   int   `json:"stars, omitempty"`
	Hand    []int `json:"hand, omitempty"`
}

type GameOutput struct {
	GameId     string           `json:"gameId"`
	PlayerId   string           `json:"playerId"`
	PlayerName string           `json:"playerName"`
	GameState  *GameStateOutput `json:"gameState, omitempty"`
}

// structs for game core
type subscription struct {
	playerId      string
	playerName    string
	playerChannel chan GameState
}

type Game struct {
	id        string
	inputCh   chan inputDetails
	publishCh chan GameState
	subCh     chan subscription
	unsubCh   chan subscription
}
type GameState struct {
	gameId         string
	playerIdToName map[string]string
	started        bool
	err            error
}
