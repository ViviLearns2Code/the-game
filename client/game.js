import * as PIXI from './pixi.mjs';
"use strict";

export class GameUI extends PIXI.Container {
  constructor(websocket) {
    super()

    const titleText = new PIXI.Text('The Game');
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 100;

    this.levelText = new PIXI.Text('0');
    this.addChild(this.levelText);
    this.levelText.x = 200;
    this.levelText.y = 100;

    this.livesText = new PIXI.Text('0');
    this.addChild(this.livesText);
    this.livesText.x = 200;
    this.livesText.y = 150;

    this.starsText = new PIXI.Text('0');
    this.addChild(this.starsText);
    this.starsText.x = 200;
    this.starsText.y = 200;

    this.topCardText = new PIXI.Text('0');
    this.addChild(this.topCardText);
    this.topCardText.x = 200;
    this.topCardText.y = 250;

    this.handText = new PIXI.Text('0');
    this.addChild(this.handText);
    this.handText.x = 200;
    this.handText.y = 300;

    this.playerNames = new PIXI.Text('');
    this.addChild(this.playerNames);
    this.playerNames.x = 400;
    this.playerNames.y = 100;

    this.proposeStarButton = new PIXI.Text('Prop. Star');
    this.addChild(this.proposeStarButton);
    this.proposeStarButton.x = 50;
    this.proposeStarButton.y = 450;

    this.proposeStarButton.interactive = true;
    this.proposeStarButton.buttonMode = true;
    this.proposeStarButton.on('pointerdown', onProposeStarButtonClick);

    var self = this
    function onProposeStarButtonClick() {
      var text = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "propose-star",
          "playerName": self.playerName,
          "cardId": "",
        }
      );
      console.debug(text)
      websocket.send(text);
    }

    const proposeConcentrationButton = new PIXI.Text('Prop. Concentration');
    this.addChild(proposeConcentrationButton);
    proposeConcentrationButton.x = 50;
    proposeConcentrationButton.y = 400;

    proposeConcentrationButton.interactive = true;
    proposeConcentrationButton.buttonMode = true;
    proposeConcentrationButton.on('pointerdown', onProposeConcentrationButtonClick);
    function onProposeConcentrationButtonClick() {
      var text = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "concentrate",
          "playerName": self.playerName,
          "cardId": "",
        }
      )
      console.debug(text)
      websocket.send(text);
    }

    this.playCardButton = new PIXI.Text('Play card');
    this.addChild(this.playCardButton);
    this.playCardButton.x = 50;
    this.playCardButton.y = 500;

    this.playCardButton.interactive = true;
    this.playCardButton.buttonMode = true;
    this.playCardButton.on('pointerdown', onPlayCardButtonClick.bind(this));
    function onPlayCardButtonClick() {
      var message = JSON.stringify(
        {
          "gameToken": self.gameToken,
          "playerToken": self.playerToken,
          "actionId": "card",
          "playerName": self.playerName,
          "cardId": this.hand[0],
        }
      );
      console.debug(message)
      websocket.send(message);
    }

    this.visible = false
    this.targetVisible = false
  }

  parseGameState(gameState) {

    // Main switch animation
    var visibleBefore = this.targetVisible;
    this.targetVisible = (gameState.gameState?.readyEvent)
                         && (gameState.gameState?.readyEvent?.name !== "lobby");

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


    this.levelText.text = "Level " + gameState.gameState.cardsOnTable.level;
    this.livesText.text = "Lives " + gameState.gameState.cardsOnTable.lives;
    this.starsText.text = "Stars " + gameState.gameState.cardsOnTable.stars;
    this.topCardText.text = "TopCard " + gameState.gameState.cardsOnTable.topCard;
    this.handText.text = "Hand " + gameState.gameState.cardsOfPlayer.cardsInHand;

    this.playerNames.text = ""
    for (const [key, value] of Object.entries(gameState.gameState.playerNames)) {
      this.playerNames.text += value
      this.playerNames.text += "(" + gameState.gameState.cardsOfPlayer.nrCardOfOtherPlayers[key] + " cards)"
      this.playerNames.text += "\n"
    }

    this.proposeStarButton.visible = gameState.gameState.cardsOnTable.stars > 0;
    this.playCardButton.visible = gameState.gameState.cardsOfPlayer.cardsInHand.length > 0;

    this.hand = gameState.gameState.cardsOfPlayer.cardsInHand;



    if (gameState.gameState?.placeCardEvent?.name === "placeCard") {
      var triggeredById = gameState.gameState.placeCardEvent.triggeredBy;
      var cardPlayedText = new PIXI.Text('Player ' + gameState.gameState.playerNames[triggeredById] + ' played card ' + gameState.gameState.placeCardEvent.discardedCard);
      cardPlayedText.anchor.set(0.5);
      cardPlayedText.x = 400
      cardPlayedText.y = 250
      cardPlayedText.visible = false;
      this.addChild(cardPlayedText);

      var self = this;
      const coords = {scale: 0, pos_y: cardPlayedText.y}
      var tweenShowCardPlayed = new TWEEN.Tween(coords)
        .to({scale: 1, pos_y: cardPlayedText.y}, 250)
        .easing(TWEEN.Easing.Exponential.Out)
        .onStart(()=>{
          cardPlayedText.visible = true;
        })
        .onUpdate(() => {
          cardPlayedText.scale.x = coords.scale;
          cardPlayedText.scale.y = coords.scale;
          cardPlayedText.y = coords.pos_y;
        })
        .start()
      var tweenHideCardPlayed = new TWEEN.Tween(coords)
        .to({scale: 0, pos_y: cardPlayedText.y + 100}, 2000)
        .easing(TWEEN.Easing.Quadratic.In)
        .onUpdate(() => {
          cardPlayedText.scale.x = coords.scale;
          cardPlayedText.scale.y = coords.scale;
          cardPlayedText.y = coords.pos_y;
        })
        .onComplete(()=>{
          self.removeChild(cardPlayedText);
        })
        tweenShowCardPlayed.chain(tweenHideCardPlayed);
    }

    if (gameState.gameState?.gameStateEvent?.name === "gameOver") {
      const skewStyle = new PIXI.TextStyle({
        fontFamily: 'Arial',
        dropShadow: true,
        dropShadowAlpha: 0.8,
        dropShadowAngle: 2.1,
        dropShadowBlur: 4,
        dropShadowColor: "0x111111",
        dropShadowDistance: 10,
        fill: ['#ffffff'],
        stroke: '#ff0000', //'#004620',
        fontSize: 40,
        fontWeight: "lighter",
        lineJoin: "round",
        strokeThickness: 12
      });

      var gameOverText = new PIXI.Text('Big fail. You not in sync. Totally wrong. >:(', skewStyle);
      gameOverText.anchor.set(0.5);
      gameOverText.x = 400
      gameOverText.y = 250
      gameOverText.visible = false;
      this.addChild(gameOverText);

      var self = this;

      const coords = {scale: 0, pos_y: gameOverText.y}
      var tweenShowGameOver = new TWEEN.Tween(coords)
        .to({scale: 1.0, pos_y: gameOverText.y}, 15000)
        .easing(TWEEN.Easing.Exponential.Out)
        .onStart(()=>{
          gameOverText.visible = true;
        })
        .onUpdate(() => {
          gameOverText.scale.x = coords.scale;
          gameOverText.scale.y = coords.scale;
          gameOverText.y = coords.pos_y;
        })
        .onComplete(()=>{
          self.removeChild(gameOverText);
        })
        .start()
    }

  }

}

