
class WelcomeUI extends PIXI.Container {
  constructor() {
    super()

    const titleText = new PIXI.Text('The Game');
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 100;

    var inputName = new PIXI.TextInput({
      input: {fontSize: '25px'},
      box: {fill: 0xEEEEEE}
    })
    inputName.x = 50
    inputName.y = 300
    inputName.placeholder = 'Player name...'
    this.addChild(inputName)
    inputName.focus()

    const hostButton = new PIXI.Text('Host New Game');
    this.addChild(hostButton);
    hostButton.x = 50;
    hostButton.y = 200;

    hostButton.interactive = true;
    hostButton.buttonMode = true;
    hostButton.on('pointerdown', onHostButtonClick);
    function onHostButtonClick() {
      if (inputName.text == "")
        return;
      var message = JSON.stringify(
        {
          "actionId": "create",
          "playerName": inputName.text,
        }
      );
      console.debug("sending: " + message)
      websocket.send(message);
    }
    this.visible = false
  }

  parseGameState(gameState) {

    var visibleBefore = this.targetVisible;
    this.targetVisible = !gameState.gameState;

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

  }
}
