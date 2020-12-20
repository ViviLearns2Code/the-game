import * as PIXI from './pixi.mjs';
import * as PIXITEXTINPUT from './PIXI.TextInput.js';
import { Styles } from './style.js'

export class LobbyUI extends PIXI.Container {
  constructor(websocket) {
    super()

    const titleText = new PIXI.Text('Silent Launchpad', Styles.headingStyle);
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 50;
    new PIXI.Text
    this.tokenText = new PIXITEXTINPUT.TextInput({
      input: {fontSize: '25px'},
      box: {fill: 0xEEEEEE}
    })
    this.tokenText.x = 50;
    this.tokenText.y = 150;
    this.addChild(this.tokenText)

    this.playerIcons = [5];
    this.playerStateText = [5];
    for (var i = 1; i < 5; i++) {
      this.playerStateText[i] = new PIXI.Text('', Styles.infoStyle);
      this.addChild(this.playerStateText[i]);
      this.playerStateText[i].x = 110;
      this.playerStateText[i].y = 180 + i * 60;

      this.playerIcons[i] = new PIXI.Sprite(PIXI.Texture.WHITE);
      this.addChild(this.playerIcons[i]);
      this.playerIcons[i].x = 50;
      this.playerIcons[i].y = 175 + i * 60;
      this.playerIcons[i].width = 50;
      this.playerIcons[i].height = 50;
    }


    const readyButton = new PIXI.Text('Ready', Styles.buttonStyle);
    this.addChild(readyButton);
    readyButton.x = 50;
    readyButton.y = 500;

    readyButton.interactive = true;
    readyButton.buttonMode = true;

    var self = this;
    console.debug(self)
    readyButton.on('pointerdown', onReadyButtonClick);
    function onReadyButtonClick() {
      console.debug(self.gameToken)
      var text = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "start",
          "playerName": self.playerName,
          "cardId": "",
        }
      )
      console.debug(text)
      websocket.send(text);
    }

    this.visible = false
    this.targetVisible = false
  }

  parseGameState(gameState) {

    // Main switch animation
    var visibleBefore = this.targetVisible;
    this.targetVisible = gameState.gameState?.readyEvent?.name === "lobby";

    if (visibleBefore != this.targetVisible) {
      if (this.tween)
        this.tween.stop()

      const coords = {scale: visibleBefore ? 1 : 0}

      var self = this;
      this.tween = new TWEEN.Tween(coords)
        .to({scale: this.targetVisible ? 1 : 0}, 750)
        .easing(TWEEN.Easing.Quadratic.In)
        .onUpdate(() => {
          self.scale.x = coords.scale;
          self.scale.y = coords.scale;
        })
        .onStart(()=>{
          if (this.targetVisible)
            self.visible = true
        })
        .onComplete(()=>{
          if (!this.targetVisible)
            self.visible = false
        })
        .start()
    }

    if (!this.targetVisible)
      return

    this.playerName = gameState.gameState.playerName;
    this.playerToken = gameState.gameState.playerToken;
    this.gameToken = gameState.gameState.gameToken;

    this.tokenText.text = gameState.gameState.gameToken;

    for (var playerId = 1; playerId <= 4; playerId++) {
      if (playerId in gameState.gameState.playerNames) {
        this.playerStateText[playerId].visible = true;
        this.playerStateText[playerId].text = gameState.gameState.playerNames[playerId];
        this.playerStateText[playerId].text += " " + (gameState.gameState.readyEvent.ready.includes(parseInt(playerId)) ? "ready" : "not ready")

        this.playerIcons[playerId].visible = true;
        var iconId = gameState.gameState.playerIconIds[playerId];
        this.playerIcons[playerId].texture = PIXI.Texture.from(`artefacts/${iconId}.png`);
      } else {
        this.playerStateText[playerId].visible = false;
        this.playerIcons[playerId].visible = false;
      }
    }

  }


}