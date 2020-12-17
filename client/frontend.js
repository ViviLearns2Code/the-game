//import * as PIXI from "pixi.js";
import * as PIXI from './pixi.mjs'
import { WelcomeUI } from './welcome.js'
import { LobbyUI } from './lobby.js'
import { GameUI } from './game.js'
import { TestUI } from './test.js'

var websocket = new WebSocket("wss://game-backend.linusseelinger.de/socket");


const app = new PIXI.Application({ backgroundColor: 0x1099bb });
document.body.appendChild(app.view);

var welcomeContainer = new WelcomeUI(websocket);
app.stage.addChild(welcomeContainer);

var lobbyContainer = new LobbyUI(websocket);
app.stage.addChild(lobbyContainer);

var gameContainer = new GameUI(websocket);
app.stage.addChild(gameContainer);

var testContainer = new TestUI(parseGameStateGlobal)
app.stage.addChild(testContainer)

function parseGameStateGlobal(gameState) {
  welcomeContainer.parseGameState(gameState);
  lobbyContainer.parseGameState(gameState);
  gameContainer.parseGameState(gameState);
  testContainer.parseGameState(gameState);
}

websocket.onmessage = function (event) {
  console.log(event.data);

  var gameState = JSON.parse(event.data);
  parseGameStateGlobal(gameState);
}
websocket.onclose = function (event) {
  console.debug("SOCKET CLOSED")
  console.debug(event)
  console.debug(event.reason)
}
websocket.onerror = function (event) {
  console.debug("SOCKET ONERROR")
  console.debug(event)
}
websocket.onopen = function (event) {
  console.debug("OPENED SOCKET")
}



// Setup the animation loop.
function animate(time) {
	requestAnimationFrame(animate)
	TWEEN.update(time)
}
requestAnimationFrame(animate)

parseGameStateGlobal(JSON.parse('{}'));
