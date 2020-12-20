import * as PIXI from './pixi.mjs';
import * as PIXITEXTINPUT from './PIXI.TextInput.js';
import { Styles } from './style.js'

export class WelcomeUI extends PIXI.Container {
  constructor(websocket) {
    super()

    const titleText = new PIXI.Text('Silence', Styles.headingStyle);
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 100;

    var inputName = new PIXITEXTINPUT.TextInput({
      input: {fontSize: '25px'},
      box: {fill: 0xEEEEEE}
    })
    inputName.x = 50;
    inputName.y = 200;
    inputName.placeholder = 'Player name...'
    this.addChild(inputName)
    inputName.focus()

    var inputGameToken = new PIXITEXTINPUT.TextInput({
      input: {fontSize: '25px'},
      box: {fill: 0xEEEEEE}
    })
    inputGameToken.x = 400
    inputGameToken.y = 200
    inputGameToken.placeholder = 'Game token...'
    this.addChild(inputGameToken)


    const hostButton = new PIXI.Text('Host Game', Styles.buttonStyle);
    this.addChild(hostButton);
    hostButton.x = 50;
    hostButton.y = 300;

    hostButton.interactive = true;
    hostButton.buttonMode = true;
    hostButton.on('pointerdown', onHostButtonClick);
    function onHostButtonClick() {
      if (inputName.text == "")
        return;
      var message = JSON.stringify(
        {
          "gameToken": "",
          "playerToken": "",
          "actionId": "create",
          "playerName": inputName.text,
          "cardId": "",
        }
      );
      console.debug("sending: " + message)
      websocket.send(message);
    }

    const joinButton = new PIXI.Text('Join Game', Styles.buttonStyle);
    this.addChild(joinButton);
    joinButton.x = 400;
    joinButton.y = 300;

    joinButton.interactive = true;
    joinButton.buttonMode = true;
    joinButton.on('pointerdown', onJoinButtonClick);
    function onJoinButtonClick() {
      if (inputName.text === "" || inputGameToken.text === "")
        return;
      var message = JSON.stringify(
        {
          "gameToken": inputGameToken.text,
          "playerToken": "",
          "actionId": "join",
          "playerName": inputName.text,
          "cardId": "",
        }
      );
      console.debug("sending: " + message)
      websocket.send(message);
    }

  }

  parseGameState(gameState) {

    // Main switch animation
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

//export default WelcomeUI;
//export { WelcomeUI as default };