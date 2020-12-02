# The Game
WIP

## Interface
### Create Game
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
### Join Game
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
### Simple Game Actions
* concentrate
* ready
* propose-star
* agree-star
* reject-star

Example: concentrate+ready
```json
// request from bob
{
  "action": "concentrate",
  "card": 0,
  "game_id": "1",
  "player_id": "1",
  "player_name": "bob"
}
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
```
```json
// request from alice
{
  "action": "ready",
  "card": 0,
  "game_id": "1",
  "player_id": "2",
  "player_name": "alice"
}
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
```
For propose, agree, reject it is similar but with different event objects.

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
### Play Card Action

```json
// request
{
  "action": "play-card",
  "card": 12,
  "game_id": "1",
  "player_id": "2",
}
```
Sunny day scenario
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
    "lose_life": false,
    "new_level": false,
    "star_used": false
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
    "lives": 3,
    "stars": 1,
    "not_ready": [],
    "pro_star": [],
    "con_star": []
  },
  "event": {
    "lose_life": true,
    "new_level": false,
    "star_requested": false,
    "ready_requested": false
  }
}
```