import * as PIXI from './pixi.mjs';
import * as PIXITEXTINPUT from './PIXI.TextInput.js';
import { Styles } from './style.js'

export class WelcomeUI extends PIXI.Container {
  constructor(websocket) {
    super()

    this.icons = [];
    this.selectedIcon = Math.floor(Math.random() * 4.999);

    var self = this;

    for (var i = 0; i < 5; i++) {
      var icon = new PIXI.Sprite.from(`artefacts/${i}.png`);
      icon.anchor.set(0.5)

      icon.interactive = true;
      icon.buttonMode = true;
      var generateHandler = function(iteration){
        return function(){
          self.selectedIcon = iteration;
          self.updateIconsScale();
        }
      }
      icon.on('pointerdown', generateHandler(i));

      this.addChild(icon)
      icon.x = 100 + i*150;
      icon.y = 425;
      icon.width = 100;
      icon.height = 100;
      this.icons.push(icon)
    }
    this.updateIconsScale();

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
          "playerIconId": self.selectedIcon,
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
          "playerIconId": self.selectedIcon,
          "cardId": "",
        }
      );
      console.debug("sending: " + message)
      websocket.send(message);
    }

  }

  updateIconsScale() {
    for (var j = 0; j < 5; j++) {
      var coords = {scale: this.icons[j].width}
      var self = this;
      var generateHandler = function(coords, j){
        return function(){
          self.icons[j].width = coords.scale;
          self.icons[j].height = coords.scale;
        }
      }
      var tween = new TWEEN.Tween(coords)
        .to({scale: j === this.selectedIcon ? 140 : 100}, 500)
        .easing(TWEEN.Easing.Exponential.Out)
        .onUpdate(generateHandler(coords, j))
        .start()
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