//import * as PIXI from "pixi.js";
import * as PIXI from './pixi.mjs'
import { WelcomeUI } from './welcome.js'
import { LobbyUI } from './lobby.js'
import { GameUI } from './game.js'
import { ConcentrationUI } from './concentration.js'
import { StarUI } from './star.js'
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

var starContainer = new StarUI(websocket);
app.stage.addChild(starContainer);

var concentrationContainer = new ConcentrationUI(websocket)
app.stage.addChild(concentrationContainer)
concentrationContainer.x = app.stage.width / 2;
concentrationContainer.y = app.stage.height / 2;

var testContainer = new TestUI(parseGameStateGlobal)
app.stage.addChild(testContainer)


function parseGameStateGlobal(gameState) {
  welcomeContainer.parseGameState(gameState);
  lobbyContainer.parseGameState(gameState);
  gameContainer.parseGameState(gameState);
  concentrationContainer.parseGameState(gameState);
  starContainer.parseGameState(gameState);
  testContainer.parseGameState(gameState);

  if (gameState.errorMsg === "") {
    return;
  }
  showErrorToast('Server error! ' + gameState.errorMsg);
}

websocket.onmessage = function (event) {
  console.log(event.data);

  var gameState = JSON.parse(event.data);
  parseGameStateGlobal(gameState);
}
websocket.onclose = function (event) {
  showErrorToast('Websocket disconnected! ' + event.reason);
}
websocket.onerror = function (event) {
  showErrorToast('Websocket error! ' + event.reason);
}
websocket.onopen = function (event) {
  console.debug("OPENED SOCKET")
}

function showErrorToast(errorText) {
  var socketErrorText = new PIXI.Text(errorText);
  socketErrorText.x = 0;
  socketErrorText.y = -socketErrorText.height;
  socketErrorText.visible = false;
  app.stage.addChild(socketErrorText);

  const coords = {pos_y: -socketErrorText.height}
  var tweenShow = new TWEEN.Tween(coords)
    .to({pos_y: 0}, 250)
    .easing(TWEEN.Easing.Exponential.Out)
    .onStart(()=>{
      socketErrorText.visible = true;
    })
    .onUpdate(() => {
      socketErrorText.y = coords.pos_y;
    })
    .start()
  var tweenHide = new TWEEN.Tween(coords)
    .to({pos_y: -socketErrorText.height}, 5000)
    .easing(TWEEN.Easing.Quadratic.In)
    .onUpdate(() => {
      socketErrorText.y = coords.pos_y;
    })
    .onComplete(()=>{
      app.stage.removeChild(socketErrorText);
    })
  tweenShow.chain(tweenHide);
}

// Setup the animation loop.
function animate(time) {
	requestAnimationFrame(animate)
	TWEEN.update(time)
}
requestAnimationFrame(animate)
