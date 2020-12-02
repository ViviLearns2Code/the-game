# Interfaces
The stream of information between client and server is bidirectional. The communication interfaces are defined below.
## Create Game
Bob creates a game
```json
// push to server from bob
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
// push to server from alice
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
// push to server from bob
{
  "action": "concentrate",
  "game_id": "1",
  "player_id": "1",
  "player_name": "bob"
}
```
* Server pushes to both players
```json
// push to bob from server
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
    "not_ready": ["bob", "alice"],
  }
}
// push to alice from server
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
    "not_ready": ["bob", "alice"],
  }
}
```
* Alice notifies the server that she is ready
```json
// push to server from alice
{
  "action": "ready",
  "game_id": "1",
  "player_id": "2",
  "player_name": "alice"
}
```
* Server pushes to both players
```json
// push to bob from server
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
    "not_ready": ["bob"],
  }
}
// push to alice from server
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
    "not_ready": ["bob"],
  }
}
```
### Request & Use Star
* Bob requests a star
```json
// push to server from bob
{
  "action": "propose-star",
  "game_id": "1",
  "player_id": "1",
  "player_name": "bob"
}
```
* Server pushes to both players, below is the example for Bob
```json
// push to bob from server
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
    "type": "propose-star",
    "triggered_by": "bob",
    "pro_star": ["bob"],
    "con_star": [],
    "discarded": []
  }
}
```
* Alice agrees
```json
// push to server from alice
{
  "action": "agree-star",
  "game_id": "1",
  "player_id": "2",
  "player_name": "alice"
}
```
* Server pushes to both players, below is the example for Bob
```json
// push to bob from server
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
    "type": "agree-star",
    "triggered_by": "alice",
    "pro_star": ["bob", "alice"],
    "con_star": [],
    "discarded": [["bob", 20], ["alice", 12]]
  }
}
```
## Play Card
* Alice plays a card
```json
// push to server from alice
{
  "card": 12,
  "game_id": "1",
  "player_id": "2",
  "player_name": "alice"
}
```
* Sunny day scenario: The card was placed in the correct order
```json
// push to alice from server
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
Rainy day scenario (Bob has 11 on his hand)
```json
// push to alice from server
{
  "game_id": "1",
  "player_id": "2",
  "game_state": {
    "hand": [35,40,99],
    "players_card_count": [{
      "bob": 6,
      "alice": 3
    }],
    "top_card": 12,
    "level": 1,
    "lives": 2,
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
      "discarded": [["bob", 11]]
      "dead": false
    }
  }
}
```
