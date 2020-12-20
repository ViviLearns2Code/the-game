import * as PIXI from './pixi.mjs';
import { Styles } from './style.js'

export class StarUI extends PIXI.Container {
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

    const titleText = new PIXI.Text('Star proposed', Styles.headingStyle);
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 50;

    this.playerStatesText = new PIXI.Text('', Styles.infoStyle);
    this.addChild(this.playerStatesText);
    this.playerStatesText.x = 50;
    this.playerStatesText.y = 100;


    const rejectButton = new PIXI.Text('Reject', Styles.buttonStyle);
    this.addChild(rejectButton);
    rejectButton.x = 350;
    rejectButton.y = 350;

    rejectButton.interactive = true;
    rejectButton.buttonMode = true;

    var self = this;
    rejectButton.on('pointerdown', onRejectButtonClick);
    function onRejectButtonClick() {
      console.debug(self.gameToken)
      var text = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "reject-star",
          "playerName": self.playerName,
          "cardId": "",
        }
      )
      console.debug(text)
      websocket.send(text);
    }

    const acceptButton = new PIXI.Text('Accept', Styles.buttonStyle);
    this.addChild(acceptButton);
    acceptButton.x = 50;
    acceptButton.y = 350;

    acceptButton.interactive = true;
    acceptButton.buttonMode = true;

    acceptButton.on('pointerdown', onAcceptButtonClick);
    function onAcceptButtonClick() {
      console.debug(self.gameToken)
      var text = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "agree-star",
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
    this.targetVisible = gameState.gameState?.processStarEvent?.name === "proposeStar"
                         || gameState.gameState?.processStarEvent?.name === "agreeStar"
                         || gameState.gameState?.processStarEvent?.name === "rejectStar";

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

    this.playerStatesText.text = "";
    for (const [playerId, playerName] of Object.entries(gameState.gameState.playerNames)) {
      this.playerStatesText.text += playerName
      this.playerStatesText.text += " " + (gameState.gameState.processStarEvent.proStar.includes(parseInt(playerId)) ? "ready" : "not ready")
      this.playerStatesText.text += "\n"
    }

  }


}
