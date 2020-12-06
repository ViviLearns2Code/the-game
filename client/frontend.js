var websocket = new WebSocket("wss://www.example.com/socketserver");
gameId = -1
playerId = -1
playerName = ""

const app = new PIXI.Application({ backgroundColor: 0x1099bb });
document.body.appendChild(app.view);

welcomeContainer = new WelcomeUI();
app.stage.addChild(welcomeContainer);

lobbyContainer = new LobbyUI();
app.stage.addChild(lobbyContainer);

gameContainer = new GameUI();
app.stage.addChild(gameContainer);

testContainer = new TestUI()
app.stage.addChild(testContainer)

function parseGameStateGlobal(gameStateJSON) {
  var gameState = JSON.parse(gameStateJSON);
  welcomeContainer.parseGameState(gameState);
  lobbyContainer.parseGameState(gameState);
  gameContainer.parseGameState(gameState);
  testContainer.parseGameState(gameState);
}

websocket.onmessage = function (event) {
  console.log(event.data);

  var gameState = JSON.parse(event.data); // TODO: Avoid duplicate JSON extraction

  if ("gameState" in gameState)
    parseGameStateGlobal(event.data);
  else {
    gameId = gameState.gameId;
    playerId = gameState.playerId;
    playerName = gameState.playerName;
  }

}

parseGameStateGlobal('{"state" : "welcome"}');
//switchView("welcome")
/*
const basicText = new PIXI.Text('Basic text in pixi');
basicText.x = 50;
basicText.y = 100;

app.stage.addChild(basicText);

const style = new PIXI.TextStyle({
    fontFamily: 'Arial',
    fontSize: 36,
    fontStyle: 'italic',
    fontWeight: 'bold',
    fill: ['#ffffff', '#00ff99'], // gradient
    stroke: '#4a1850',
    strokeThickness: 5,
    dropShadow: true,
    dropShadowColor: '#000000',
    dropShadowBlur: 4,
    dropShadowAngle: Math.PI / 6,
    dropShadowDistance: 6,
    wordWrap: true,
    wordWrapWidth: 440,
    lineJoin: 'round'
});

const richText = new PIXI.Text('Rich text with a lot of options and across multiple lines', style);
richText.x = 50;
richText.y = 220;

// Opt-in to interactivity
richText.interactive = true;
// Shows hand cursor
richText.buttonMode = true;
// Pointers normalize touch and mouse
richText.on('pointerdown', onClick);
// Alternatively, use the mouse & touch events:
// sprite.on('click', onClick); // mouse-only
// sprite.on('tap', onClick); // touch-only
function onClick() {
    richText.scale.x *= 1.25;
    richText.scale.y *= 1.25;
}


app.stage.addChild(richText);

const skewStyle = new PIXI.TextStyle({
    fontFamily: 'Arial',
    dropShadow: true,
    dropShadowAlpha: 0.8,
    dropShadowAngle: 2.1,
    dropShadowBlur: 4,
    dropShadowColor: "0x111111",
    dropShadowDistance: 10,
    fill: ['#ffffff'],
    stroke: '#004620',
    fontSize: 60,
    fontWeight: "lighter",
    lineJoin: "round",
    strokeThickness: 12
});

const skewText = new PIXI.Text('SKEW IS COOL', skewStyle);
skewText.skew.set(0.65,-0.3);
skewText.anchor.set(0.5, 0.5);
skewText.x = 300;
skewText.y = 480;

app.stage.addChild(skewText);

skewText.interactive = true;
// Shows hand cursor
skewText.buttonMode = true;
// Pointers normalize touch and mouse
skewText.on('pointerdown', onClick2);
// Alternatively, use the mouse & touch events:
// sprite.on('click', onClick); // mouse-only
// sprite.on('tap', onClick); // touch-only
function onClick2() {
    richText.scale.x /= 1.25;
    richText.scale.y /= 1.25;
}
*/
