import * as PIXI from './pixi.mjs';
import { Styles } from './style.js'
"use strict";

export class GameUI extends PIXI.Container {
  constructor(websocket) {
    super()

    const titleText = new PIXI.Text('Silence', Styles.headingStyle);
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 50;

    this.levelText = new PIXI.Text('0', Styles.infoStyle);
    this.addChild(this.levelText);
    this.levelText.x = 400;
    this.levelText.y = 75;

    this.livesText = new PIXI.Text('0', Styles.infoStyle);
    this.addChild(this.livesText);
    this.livesText.x = 500;
    this.livesText.y = 75;

    this.starsText = new PIXI.Text('0', Styles.infoStyle);
    this.addChild(this.starsText);
    this.starsText.x = 600;
    this.starsText.y = 75;

    this.topCardText = new PIXI.Text('0', Styles.importantStyle);
    this.addChild(this.topCardText);
    this.topCardText.anchor.set(0.5);
    this.topCardText.x = 400;
    this.topCardText.y = 300;

    this.handText = new PIXI.Text('0', Styles.infoStyle);
    this.addChild(this.handText);
    this.handText.anchor.set(0.5);
    this.handText.x = 400;
    this.handText.y = 460;

    this.discardedText = new PIXI.Text('', Styles.smallStyle);
    this.addChild(this.discardedText);
    this.discardedText.anchor.set(0.5);
    this.discardedText.x = 400;
    this.discardedText.y = 500;

    this.playerIcon = new PIXI.Sprite(PIXI.Texture.WHITE);
    this.playerIcon.anchor.set(0.5);
    this.addChild(this.playerIcon);
    this.playerIcon.x = 400;
    this.playerIcon.y = 400;
    this.playerIcon.width = 50;
    this.playerIcon.height = 50;

    this.coplayers = []
    var xList = [100, 400, 675]
    var yList = [250, 150, 250]
    for (var i = 0; i < 3; i++) {
      this.coplayers[i] = {
        "text": new PIXI.Text("", Styles.infoStyle),
        "icon": new PIXI.Sprite(PIXI.Texture.WHITE),
        "discarded": new PIXI.Text("", Styles.smallStyle)
      }
      this.addChild(this.coplayers[i].text);
      this.coplayers[i].text.anchor.set(0.5);
      this.coplayers[i].text.x = xList[i];
      this.coplayers[i].text.y = yList[i];
      this.coplayers[i].text.visible = false;

      this.addChild(this.coplayers[i].discarded);
      this.coplayers[i].discarded.anchor.set(0.5);
      this.coplayers[i].discarded.x = xList[i];
      this.coplayers[i].discarded.y = yList[i]+100;
      this.coplayers[i].discarded.visible = false;

      this.addChild(this.coplayers[i].icon);
      this.coplayers[i].icon.x = xList[i];
      this.coplayers[i].icon.y = yList[i]+50;
      this.coplayers[i].icon.width = 50;
      this.coplayers[i].icon.height = 50;
      this.coplayers[i].icon.anchor.set(0.5);
      this.coplayers[i].icon.visible = false;
    }

    this.proposeStarButton = new PIXI.Text('Star!', Styles.buttonStyle);
    this.addChild(this.proposeStarButton);
    this.proposeStarButton.x = 400;
    this.proposeStarButton.y = 550;

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

    const proposeConcentrationButton = new PIXI.Text('Concentrate!', Styles.buttonStyle);
    this.addChild(proposeConcentrationButton);
    proposeConcentrationButton.x = 100;
    proposeConcentrationButton.y = 550;

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

    this.playCardButton = new PIXI.Text('Play!', Styles.buttonStyle);
    this.addChild(this.playCardButton);
    this.playCardButton.x = 600;
    this.playCardButton.y = 550;

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
          "cardId": parseInt(this.hand[0]),
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


    this.levelText.text = "⚑ " + gameState.gameState.cardsOnTable.level;
    this.livesText.text = "❤ " + gameState.gameState.cardsOnTable.lives;
    this.starsText.text = "★ " + gameState.gameState.cardsOnTable.stars;
    this.topCardText.text = "Top " + gameState.gameState.cardsOnTable.topCard;
    this.handText.text = gameState.gameState.cardsOfPlayer.cardsInHand;

    var coplayersIndex = 0;
    for(const id of Object.keys(gameState.gameState.playerNames)){
      if (id == gameState.gameState.playerId) {
        // render player icon
        this.playerIcon.visible = true;
        var iconId = gameState.gameState.playerIconId;
        this.playerIcon.texture = PIXI.Texture.from(`artefacts/${iconId}.png`);
        // render player's discarded card for star
        if (id in gameState.gameState.placeCardEvent.discardedCard) {
          this.discardedText.visible = true;
          this.discardedText.text = "discard: " + gameState.gameState.placeCardEvent.discardedCard[id];
        } else {
          this.discardedText.visible = false;
        }
      } else {
        this.coplayers[coplayersIndex].text.visible = true;
        this.coplayers[coplayersIndex].text.text = gameState.gameState.playerNames[id];
        this.coplayers[coplayersIndex].text.text += " (" + gameState.gameState.cardsOfPlayer.nrCardOfOtherPlayers[id] + ")";

        if (id in gameState.gameState.placeCardEvent.discardedCard) {
          this.coplayers[coplayersIndex].discarded.visible = true;
          this.coplayers[coplayersIndex].discarded.text = "discard: " + gameState.gameState.placeCardEvent.discardedCard[id];
        } else {
          this.coplayers[coplayersIndex].discarded.visible = false;
        }
        this.coplayers[coplayersIndex].icon.visible = true;
        var iconId = gameState.gameState.playerIconIds[id];
        this.coplayers[coplayersIndex].icon.texture = PIXI.Texture.from(`artefacts/${iconId}.png`);
        coplayersIndex += 1;
      }
    };

    this.proposeStarButton.visible = gameState.gameState.cardsOnTable.stars > 0;
    this.playCardButton.visible = gameState.gameState.cardsOfPlayer.cardsInHand.length > 0;

    this.hand = gameState.gameState.cardsOfPlayer.cardsInHand;



    if (gameState.gameState?.placeCardEvent?.name === "placeCard") {
      var triggeredById = gameState.gameState.placeCardEvent.triggeredBy;
      var discardedCards = gameState.gameState.placeCardEvent.discardedCard;
      var cardPlayedText = new PIXI.Text('Player ' + gameState.gameState.playerNames[triggeredById] + ' played card ' + discardedCards[triggeredById], Styles.infoStyle);
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
          location.reload();
        })
        .start()
    }

    if (gameState.gameState?.gameStateEvent?.name === "gameWon") {
      const skewStyle = new PIXI.TextStyle({
        fontFamily: 'Arial',
        dropShadow: true,
        dropShadowAlpha: 0.8,
        dropShadowAngle: 2.1,
        dropShadowBlur: 4,
        dropShadowColor: "0x111111",
        dropShadowDistance: 10,
        fill: ['#ffffff'],
        stroke: '#00ff00', //'#004620',
        fontSize: 40,
        fontWeight: "lighter",
        lineJoin: "round",
        strokeThickness: 12
      });

      var gameOverText = new PIXI.Text('Congratulations! You are part of the collective mind now.', skewStyle);
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
          location.reload();
        })
        .start()
    }

  }
}
