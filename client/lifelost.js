import * as PIXI from './pixi.mjs';
import { Styles } from './style.js'

export class LifeLostUI extends PIXI.Container {
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

    this.detailText = new PIXI.Text('', Styles.infoStyle);
    this.addChild(this.detailText);
    this.detailText.x = 0;
    this.detailText.y = 50;

    this.visible = false
  }

  parseGameState(gameState) {

    if (!(gameState.gameState?.gameStateEvent?.livesDecrease))
      return;

    this.titleText.text = "Out of synch!";
    this.detailText.text = "You lose a life";
    for (const [playerId, cards] of Object.entries(gameState.gameState.placeCardEvent.discardedCard)){
      var verb = " has "
      if(cards.length == 1 && cards[0] == gameState.gameState.cardsOnTable.topCard){
        verb = " played "
      }
      this.detailText.text += "\n" + gameState.gameState.playerNames[playerId] + verb + cards
    }

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
      .to({scale: 0}, 4000)
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