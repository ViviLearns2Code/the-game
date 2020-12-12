import * as PIXI from './pixi.mjs';

export class LobbyUI extends PIXI.Container {
  constructor(websocket) {
    super()

    const titleText = new PIXI.Text('The Game\'s Lobby');
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 100;

    const joinButton = new PIXI.Text('Leave');
    this.addChild(joinButton);
    joinButton.x = 50;
    joinButton.y = 200;

    joinButton.interactive = true;
    joinButton.buttonMode = true;
    joinButton.on('pointerdown', onJoinButtonClick);
    function onJoinButtonClick() {

    }

    const readyButton = new PIXI.Text('Ready');
    this.addChild(readyButton);
    readyButton.x = 50;
    readyButton.y = 300;

    readyButton.interactive = true;
    readyButton.buttonMode = true;
    readyButton.on('pointerdown', onReadyButtonClick);
    function onReadyButtonClick() {

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


  }


}