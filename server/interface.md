# Interfaces
The stream of information between client and server is bidirectional. The communication interfaces are defined below.
## Create Game
Bob creates a game
```json
// push to server from bob
{
  "actionId": "create",
  "playerName": "bob"
}
// response
// -> The great big response structure
```
## Join Game
Alice joins the game Bob created
```json
// push to server from alice
{
  "actionId": "join",
  "gameId": "1",
  "playerName": "alice"
}
// response
// -> The great big response structure
```
## The great not so big request structure
```json
{
  "actionId": "concentrate",
  "gameToken": "uiaoy1246247dnr",
  "playerToken": "1135dtrndtrn7365",
  "card": null
}
```

## The great big response structure
```json
{
  "gameToken": "usxu34vywr12-1346174", // random token
  "playerName": "Tom",
  "playerId": "2", // 1-4
  "playerToken": "uiodu-241346147uiaedtrnu-", // random token
  "playerNames": [{
    "1": "Tom",
    "2": "Jerry",
    "3": "Tom",
    "4": "Tom"
  }],
  "cards": {
    "hand": [35,40,99],
    "playersCardCount": [{
      "1": 7, // playerId!
      "2": 3,
      "3": 3,
      "4": 3
    }],
    "topCard": 12,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "placed-card": {
    "active": "true",
    "triggeredBy": "3", // playerId
    "discarded": [["4", 11]], // playerId's and discarded cards
  },
  "game-over": {
    "active": false
  },
  "level-finished": {
    "active": false,
    "levelUp": false,
    "levelTitle": "",
    "starsIncrease": true,
    "livesIncrease": false,
  },
  "lobby": {
    "active": false,
    "ready": ["2", "4"],
  },
  "concentrating": {
    "active": false,
    "triggeredBy": "2",
    "ready": ["1", "2"],
  },
  "proposed-star": {
    "active": false,
    "triggeredBy": "1",
    "proStar": ["1", "3"],
    "conStar": null,
  },
  "agree-star": {
    "active": true,
    "triggeredBy": "3",
    "proStar": ["3", "4"],
    "conStar": [],
  },
  "reject-star": {
    "active": false,
    "triggeredBy": "2",
    "proStar": ["2", "4"],
    "conStar": [],
  },
  "star-accepted": {
    "active": true,
    "lowest-discarded": [["4", 11]], // playerId's and discarded cards
  }
}
```



## Simple Actions
* concentrate
* ready
* propose-star
* agree-star
* reject-star

### Request Concentration & Player Readiness
* Bob requests concentration
```json
// push to server from bob
{
  "actionId": "concentrate",
  "gameToken": "uiaoy1246247dnr",
  "playerToken": "1135dtrndtrn7365",
  "cardId": null
}
```
* Server pushes to both players
```json
// push to bob from server
{
  "gameId": "1",
  "playerId": "1",
  "gameState": {
    "hand": [20,43,61,62,68,90,100],
    "playersCardCount": [{
      "bob": 7,
      "alice": 4
    }],
    "topCard": 10,
    "level": 1,
    "lives": 3,
    "stars": 1,
  },
  "event": {
    "type": "concentrate",
    "triggeredBy": "bob",
    "notReady": ["bob", "alice"],
  }
}
// push to alice from server
{
  "gameId": "1",
  "playerId": "2",
  "gameState": {
    "hand": [12,35,40,99],
    "playersCardCount": [{
      "bob": 7,
      "alice": 4
    }],
    "topCard": 10,
    "level": 1,
    "lives": 3,
    "stars": 1,
  },
  "event": {
    "type": "concentrate",
    "triggeredBy": "bob",
    "notReady": ["bob", "alice"],
  }
}
```
* Alice notifies the server that she is ready
```json
// push to server from alice
{
  "actionId": "ready",
  "gameId": "1",
  "playerId": "2",
  "playerName": "alice"
}
```
* Server pushes to both players
```json
// push to bob from server
{
  "gameId": "1",
  "playerId": "1",
  "gameState": {
    "hand": [12,35,40,99],
    "playersCardCount": [{
      "bob": 7,
      "alice": 4
    }],
    "topCard": 10,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "event": {
    "type": "ready",
    "triggeredBy": "alice",
    "notReady": ["bob"],
  }
}
// push to alice from server
{
  "gameId": "1",
  "playerId": "2",
  "gameState": {
    "hand": [12,35,40,99],
    "playersCardCount": [{
      "bob": 7,
      "alice": 4
    }],
    "topCard": 10,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "event": {
    "type": "ready",
    "triggeredBy": "alice",
    "notReady": ["bob"],
  }
}
```
### Request & Use Star
* Bob requests a star
```json
// push to server from bob
{
  "actionId": "propose-star",
  "gameId": "1",
  "playerId": "1",
  "playerName": "bob"
}
```
* Server pushes to both players, below is the example for Bob
```json
// push to bob from server
{
  "gameId": "1",
  "playerId": "1",
  "gameState": {
    "hand": [20,43,61,62,68,90,100],
    "playersCardCount": [{
      "bob": 7,
      "alice": 4
    }],
    "topCard": 10,
    "level": 1,
    "lives": 3,
    "stars": 1,
  },
  "event": {
    "type": "propose-star",
    "triggeredBy": "bob",
    "proStar": ["bob"],
    "conStar": [],
    "discarded": []
  }
}
```
* Alice agrees
```json
// push to server from alice
{
  "actionId": "agree-star",
  "gameId": "1",
  "playerId": "2",
  "playerName": "alice"
}
```
* Server pushes to both players, below is the example for Bob
```json
// push to bob from server
{
  "gameId": "1",
  "playerId": "1",
  "gameState": {
    "hand": [20,43,61,62,68,90,100],
    "playersCardCount": [{
      "bob": 7,
      "alice": 4
    }],
    "topCard": 10,
    "level": 1,
    "lives": 3,
    "stars": 1,
  },
  "event": {
    "type": "agree-star",
    "triggeredBy": "alice",
    "proStar": ["bob", "alice"],
    "conStar": [],
    "discarded": [["bob", 20], ["alice", 12]]
  }
}
```
## Play Card
* Alice plays a card
```json
// push to server from alice
{
  "actionId": "card",
  "cardId": 12,
  "gameId": "1",
  "playerId": "2",
  "playerName": "alice"
}
```
* Sunny day scenario: The card was placed in the correct order
```json
// push to alice from server

```
Rainy day scenario (Bob has 11 on his hand)
```json
// push to alice from server
{
  "gameId": "1",
  "playerId": "2",
  "gameState": {
    "hand": [35,40,99],
    "playersCardCount": [{
      "bob": 6,
      "alice": 3
    }],
    "topCard": 12,
    "level": 1,
    "lives": 2,
    "stars": 1,
    "notReady": [],
    "proStar": [],
    "conStar": []
  },
  "event": {
    "type": "placed-card",
    "triggeredBy": "alice",
    "details": {
      "discarded": [["bob", 11]],
      "dead": false
    }
  }
}
```
