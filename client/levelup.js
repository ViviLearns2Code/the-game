import * as PIXI from './pixi.mjs';
import { Styles } from './style.js'

export class LevelUpUI extends PIXI.Container {
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

    this.titleText = new PIXI.Text('', Styles.headingStyle);
    this.addChild(this.titleText);
    this.titleText.x = 0;
    this.titleText.y = 0;

    this.perksText = new PIXI.Text('', Styles.infoStyle);
    this.addChild(this.perksText);
    this.perksText.x = 0;
    this.perksText.y = 50;

    this.visible = false
  }

  parseGameState(gameState) {

    if (!(gameState.gameState?.gameStateEvent?.name === "levelUp"))
      return;

    this.titleText.text = gameState.gameState.gameStateEvent.levelTitle;
    this.perksText.text = "You reached a new level"
    if (gameState.gameState.gameStateEvent.starsIncrease)
      this.perksText.text += "\n+1 ★";
    else if (gameState.gameState.gameStateEvent.livesIncrease)
      this.perksText.text += "\n+1 ❤";

    const coords = {scale: 0}
    var self = this;
    var wait = 0
    if (gameState.gameState?.gameStateEvent?.livesDecrease){
      wait = 4000
    }

    var tweenShow = new TWEEN.Tween(coords)
      .to({scale: 1}, 750)
      .easing(TWEEN.Easing.Quadratic.In)
      .onUpdate(() => {
        self.scale.x = coords.scale;
        self.scale.y = coords.scale;
      })
      .onStart(()=>{
        self.visible = true
      })
      .start()

    var tweenWait = new TWEEN.Tween(coords)
      .to({scale: 1}, wait)

    var tweenHide = new TWEEN.Tween(coords)
      .to({scale: 0}, 4000)
      .easing(TWEEN.Easing.Quadratic.In)
      .onUpdate(() => {
        self.scale.x = coords.scale;
        self.scale.y = coords.scale;
      })
      .onComplete(()=>{
        self.visible = false;
      })
    tweenShow.chain(tweenWait)
    tweenWait.chain(tweenHide);

  }


}