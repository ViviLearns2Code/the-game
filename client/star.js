import * as PIXI from './pixi.mjs';

export class StarUI extends PIXI.Container {
  constructor(websocket) {
    super()

    this.width = 400;
    this.height = 400;
    this.pivot.x = this.width / 2;
    this.pivot.y = this.height / 2;

    const titleText = new PIXI.Text('Star proposed');
    this.addChild(titleText);
    titleText.x = 0;
    titleText.y = 0;

    this.playerStatesText = new PIXI.Text('');
    this.addChild(this.playerStatesText);
    this.playerStatesText.x = 0;
    this.playerStatesText.y = 50;


    const rejectButton = new PIXI.Text('Reject');
    this.addChild(rejectButton);
    rejectButton.x = 0;
    rejectButton.y = 300;

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

    const acceptButton = new PIXI.Text('Accept');
    this.addChild(acceptButton);
    acceptButton.x = 0;
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
    this.targetVisible = gameState.gameState?.processStarEvent?.name === "proposeStar";

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
