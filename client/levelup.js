import * as PIXI from './pixi.mjs';

export class LevelUpUI extends PIXI.Container {
  constructor(websocket) {
    super()

    this.pivot.x = 400 / 2;
    this.pivot.y = 400 / 2;

    const bkg = new PIXI.Sprite(PIXI.Texture.WHITE);
    this.addChild(bkg);
    bkg.x = 0;
    bkg.y = 0;
    bkg.width = 400;
    bkg.height = 400;

    this.titleText = new PIXI.Text('');
    this.addChild(this.titleText);
    this.titleText.x = 0;
    this.titleText.y = 0;

    this.perksText = new PIXI.Text('');
    this.addChild(this.perksText);
    this.perksText.x = 0;
    this.perksText.y = 50;

    this.visible = false
  }

  parseGameState(gameState) {

    if (!(gameState.gameState?.gameStateEvent?.name === "levelUp"))
      return;

    this.titleText.text = gameState.gameState.gameStateEvent.levelTitle;
    if (gameState.gameState.gameStateEvent.starsIncrease)
      this.perksText.text = "New star!";
    else if (gameState.gameState.gameStateEvent.livesIncrease)
      this.perksText.text = "New life!";

    const coords = {scale: 0}
    var self = this;
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
    var tweenHide = new TWEEN.Tween(coords)
      .to({scale: 0}, 2000)
      .easing(TWEEN.Easing.Quadratic.In)
      .onUpdate(() => {
        self.scale.x = coords.scale;
        self.scale.y = coords.scale;
      })
      .onComplete(()=>{
        self.visible = false;
      })
      tweenShow.chain(tweenHide);

  }


}