# Interfaces
The stream of information between client and server is bidirectional. The communication interfaces are defined below.

## Requests
The request json has a static structure, but fields can be null depending on the context.
```json
{
  "playerName": "bob", // uuid string
  "playerToken": "1135dtr-ndtrn7365", // uuid string
  "gameToken": "uiaoy12-46247dnr", // uuid string
  "actionId": "concentrate",
  "card": null
}
```
Possible `actionIds`
* create
* join
* start
* leave
* concentrate
* ready
* propose-star
* agree-star
* reject-star

## Response
The response json has a static structure, but fields can be null depending on the context.
```json
{
  "gameToken": "usxu34vywr12-1346174", // uuid string
  "playerName": "bob",
  "playerId": "1", // 1-4 unique per player
  "playerToken": "uiodu-241346147uiaedtrnu-", // uuid string
  "playerNames": [{
    "1": "bob",
    "2": "alice",
    "3": "bob", // playerName does not need to be unique
    "4": ""
  }],
  "cards": {
    "hand": [35,40,99],
    "playersCardCount": [{ // playerId -> #cards
      "1": 7,
      "2": 3,
      "3": 3,
      "4": 0
    }],
    "topCard": 12,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "placed-card": {
    "active": true,
    "triggeredBy": "3", // playerId
    "discarded": [["4", 11]], // [playerId, discarded card]
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
  "concentrate": {
    "active": false,
    "triggeredBy": "2",
    "ready": ["1", "2"],
  },
  "propose-star": {
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
  "used-star": {
    "active": true,
    "discarded": [["4", 11]], // [playerId, discarded card]
  }
}
```