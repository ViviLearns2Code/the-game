import * as PIXI from './pixi.mjs';
import { Styles } from './style.js'

export class ConcentrationUI extends PIXI.Container {
  constructor(websocket) {
    super()

    this.pivot.x = 600 / 2;
    this.pivot.y = 400 / 2;

    const bkg = new PIXI.Sprite(PIXI.Texture.WHITE);
    this.addChild(bkg);
    bkg.tint = Styles.popupTint;
    bkg.x = 0;
    bkg.y = 0;
    bkg.width = 600;
    bkg.height = 400;

    const titleText = new PIXI.Text('Concentration', Styles.headingStyle);
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 50;

    this.playerStatesText = new PIXI.Text('', Styles.infoStyle);
    this.addChild(this.playerStatesText);
    this.playerStatesText.x = 50;
    this.playerStatesText.y = 200;


    const readyButton = new PIXI.Text('Ready', Styles.buttonStyle);
    this.addChild(readyButton);
    readyButton.x = 50;
    readyButton.y = 350;

    readyButton.interactive = true;
    readyButton.buttonMode = true;

    var self = this;
    readyButton.on('pointerdown', onReadyButtonClick);
    function onReadyButtonClick() {
      var text = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "ready",
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
    var inConcentrationPhase = gameState.gameState?.readyEvent?.name === "concentrate";
    var everyoneReady = gameState.gameState?.readyEvent?.ready.length === Object.keys(gameState.gameState?.playerNames).length;
    this.targetVisible = inConcentrationPhase && !everyoneReady;

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

    if (!inConcentrationPhase)
      return

    this.playerName = gameState.gameState.playerName;
    this.playerToken = gameState.gameState.playerToken;
    this.gameToken = gameState.gameState.gameToken;

    this.playerStatesText.text = "";
    for (const [playerId, playerName] of Object.entries(gameState.gameState.playerNames)) {
      this.playerStatesText.text += playerName
      this.playerStatesText.text += " " + (gameState.gameState.readyEvent.ready.includes(parseInt(playerId)) ? "ready" : "not ready")
      this.playerStatesText.text += "\n"
    }

  }


}