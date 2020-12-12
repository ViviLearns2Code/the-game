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
type GameState struct {
	gameToken      string
	playerIdToName map[string]string
	started        bool
	err            error
}
