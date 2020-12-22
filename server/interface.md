# Interfaces
The stream of information between client and server is bidirectional. The communication interfaces are defined below.

## Requests
The request json has a static structure, but fields can be null depending on the context.
```json
{
  "playerName": "bob",
  "playerToken": "049281ac-104f-4e9b-9cc6-e494e61ecea2",
  "playerIconId": 1,
  "gameToken": "01w281ac-524f-4w3f-1hh5-i135z71rtre2",
  "actionId": "concentrate",
  "card": 0
}
```
Possible `actionIds`
* create
* join
* start
* leave
* card
* concentrate
* ready
* propose-star
* agree-star
* reject-star

## Response
The response json has a static structure.

| event | possible values |
|:------|:----------------|
|readyEvent| lobby, concentrate |
|placeCardEvent| placeCard, useStar |
|processStarEvent| proposeStar, agreeStar, rejectStar |
|gameStateEvent| gameOver, lostLife, levelUp, gameWon |

```json
{
  "errorMsg": "",
  "gameState": {
    "gameToken": "049281ac-104f-4e9b-9cc6-e494e61ecea2",
    "playerToken": "01w281ac-524f-4w3f-1hh5-i135z71rtre2",
    "playerName": "bob",
    "playerId": 1,
    "playerIconId": 1,
    "cardsOfPlayer": {
      "cardsInHand": [40, 53, 88],
      "nrCardOfOtherPlayers": {
        "1": 3,
        "2": 2
      }
    },
    "playerNames": {
      "1": "bob",
      "2": "alice"
    },
    "playerIconIds": {
      "1": 1,
      "2": 1
    },
    "cardsOnTable": {
      "topCard": 10,
      "level": 3,
      "lives": 2,
      "stars": 1
    },
    "gameStateEvent": {
      "name": "",
      "levelTitle": "",
      "starsIncrease": false,
      "starsDecrease": false,
      "livesIncrease": false,
      "livesDecrease": false
    },
    "readyEvent": {
      "name": "",
      "triggeredBy": 0,
      "ready": []
    },
    "placeCardEvent": {
      "name": "placeCard",
      "triggeredBy": 1,
      "discardedCard": {
        "1": [10]
      }
    },
    "processingStarEvent": {
      "name": "",
      "triggeredBy": 0,
      "proStar": [],
      "conStar": []
    }
  }
}
```
