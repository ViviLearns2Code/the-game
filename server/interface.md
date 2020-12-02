# Interfaces
The stream of information between client and server is bidirectional. The communication interfaces are defined below.
## Create Game
Bob creates a game
```json
// request
{
  "action": "create",
  "player_name": "bob"
}
// response
{
  "game_id": "1",
  "player_id": "1",
  "player_name": "bob"
}
```
## Join Game
Alice joins the game Bob created
```json
// request
{
  "action": "join",
  "game_id": "1",
  "player_name": "alice"
}
// response
{
  "game_id": "1",
  "player_id": "2",
  "player_name": "alice"
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
// request from bob
{
  "action": "concentrate",
  "card": 0,
  "game_id": "1",
  "player_id": "1",
  "player_name": "bob"
}
```
* The server pushes responses to both players
```json
// response for bob
{
  "game_id": "1",
  "player_id": "1",
  "game_state": {
    "hand": [20,43,61,62,68,90,100],
    "players_card_count": [{
      "bob": 7,
      "alice": 4
    }],
    "top_card": 10,
    "level": 1,
    "lives": 3,
    "stars": 1,
  },
  "event": {
    "type": "concentrate",
    "triggered_by": "bob",
    "not_ready": ["alice"],
  }
}
// response for alice
{
  "game_id": "1",
  "player_id": "2",
  "game_state": {
    "hand": [12,35,40,99],
    "players_card_count": [{
      "bob": 7,
      "alice": 4
    }],
    "top_card": 10,
    "level": 1,
    "lives": 3,
    "stars": 1,
  },
  "event": {
    "type": "concentrate",
    "triggered_by": "bob",
    "not_ready": ["alice"],
  }
}
```
* Alice notifies the server that she is ready
```json
// request from alice
{
  "action": "ready",
  "card": 0,
  "game_id": "1",
  "player_id": "2",
  "player_name": "alice"
}
```
```json
// response for bob
{
  "game_id": "1",
  "player_id": "1",
  "game_state": {
    "hand": [12,35,40,99],
    "players_card_count": [{
      "bob": 7,
      "alice": 4
    }],
    "top_card": 10,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "event": {
    "type": "ready",
    "triggered_by": "alice",
    "not_ready": [],
  }
}
// response for alice
{
  "game_id": "1",
  "player_id": "2",
  "game_state": {
    "hand": [12,35,40,99],
    "players_card_count": [{
      "bob": 7,
      "alice": 4
    }],
    "top_card": 10,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "event": {
    "type": "ready",
    "triggered_by": "alice",
    "not_ready": [],
  }
}
```
* For propose, agree, reject the process is similar but with different event objects.

```json
{
  "event": {
    "type": "propose-star",
    "triggered_by": "bob",
    "pro_star": ["bob"],
    "con_star": []
  }
}
```
## Play Card
* Alice plays a card
```json
// request
{
  "action": "play-card",
  "card": 12,
  "game_id": "1",
  "player_id": "2",
}
```
* Sunny day scenario: The card was placed in the correct order
```json
// response
{
  "game_id": "1",
  "player_id": "2",
  "game_state": {
    "hand": [35,40,99],
    "players_card_count": [{
      "bob": 7,
      "alice": 3
    }],
    "top_card": 12,
    "level": 1,
    "lives": 3,
    "stars": 1
  },
  "event": {
    "type": "placed_card",
    "triggered_by": "alice",
    "success": true,
    "details": {
      "level_up": false,
      "new_skill": "",
      "stars_increase": 0,
      "lives_increase": 0
    }
  }
}
```
Rainy day scenario (bob has 11 on his hand)
```json
// response
{
  "game_id": "1",
  "player_id": "2",
  "game_state": {
    "hand": ??,
    "players_card_count": [{
      "bob": ??,
      "alice": ??
    }],
    "top_card": ??,
    "level": 1,
    "lives": ??,
    "stars": 1,
    "not_ready": [],
    "pro_star": [],
    "con_star": []
  },
  "event": {
    "type": "placed_card",
    "triggered_by": "alice",
    "success": false,
    "details": {
      ??
    }
  }
}
```