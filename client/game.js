
class GameUI extends PIXI.Container {
  constructor() {
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

    /*joinButton.interactive = true;
    joinButton.buttonMode = true;
    joinButton.on('pointerdown', onJoinButtonClick);
    function onJoinButtonClick() {

    }*/

    this.proposeStarButton = new PIXI.Text('Prop. Star');
    this.addChild(this.proposeStarButton);
    this.proposeStarButton.x = 50;
    this.proposeStarButton.y = 300;

    this.proposeStarButton.interactive = true;
    this.proposeStarButton.buttonMode = true;
    this.proposeStarButton.on('pointerdown', onProposeStarButtonClick);
    function onProposeStarButtonClick() {
      websocket.send(JSON.stringify(
        {
          "actionId": "propose-star",
          "gameId": gameId,
          "playerId": playerId,
          "playerName": playerName
        }
      ));
    }

    const proposeConcentrationButton = new PIXI.Text('Prop. Concentration');
    this.addChild(proposeConcentrationButton);
    proposeConcentrationButton.x = 50;
    proposeConcentrationButton.y = 400;

    proposeConcentrationButton.interactive = true;
    proposeConcentrationButton.buttonMode = true;
    proposeConcentrationButton.on('pointerdown', onProposeConcentrationButtonClick);
    function onProposeConcentrationButtonClick() {
      websocket.send(JSON.stringify(
        {
          "actionId": "concentrate",
          "gameId": gameId,
          "playerId": playerId,
          "playerName": playerName
        }
      ));
    }

    this.playCardButton = new PIXI.Text('Play card');
    this.addChild(this.playCardButton);
    this.playCardButton.x = 50;
    this.playCardButton.y = 500;

    this.playCardButton.interactive = true;
    this.playCardButton.buttonMode = true;
    this.playCardButton.on('pointerdown', onPlayCardButtonClick.bind(this));
    function onPlayCardButtonClick() {
      console.debug(this.hand)
      var message = JSON.stringify(
        {
          "actionId": "card",
          "card": String(this.hand[0]),
          "gameId": gameId,
          "playerId": playerId,
          "playerName": playerName
        }
      );
      console.debug(message)
      websocket.send(message);
    }

  }

  parseGameState(gameState) {
    this.visible = "gameState" in gameState;
    if (!this.visible)
      return
    this.levelText.text = "Level " + gameState.gameState.level;
    this.livesText.text = "Lives " + gameState.gameState.lives;
    this.starsText.text = "Stars " + gameState.gameState.stars;
    this.topCardText.text = "TopCard " + gameState.gameState.topCard;
    this.handText.text = "Hand " + gameState.gameState.hand;

    this.proposeStarButton.visible = gameState.gameState.stars > 0;
    this.playCardButton.visible = gameState.gameState.hand.length > 0;

    this.hand = gameState.gameState.hand;

    console.debug(this.hand)
  }

}