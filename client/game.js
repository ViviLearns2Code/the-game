
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

    const proposeStarButton = new PIXI.Text('Prop. Star');
    this.addChild(proposeStarButton);
    proposeStarButton.x = 50;
    proposeStarButton.y = 300;

    proposeStarButton.interactive = true;
    proposeStarButton.buttonMode = true;
    proposeStarButton.on('pointerdown', onProposeStarButtonClick);
    function onProposeStarButtonClick() {
      websocket.send(JSON.stringify(
        {
          "action": "propose-star",
          "gameId": "1",
          "playerId": "1",
          "player_name": "bob"
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
          "action": "concentrate",
          "gameId": "1",
          "playerId": "1",
          "player_name": "bob"
        }
      ));
    }

    const playCardButton = new PIXI.Text('Play card');
    this.addChild(playCardButton);
    playCardButton.x = 50;
    playCardButton.y = 500;

    playCardButton.interactive = true;
    playCardButton.buttonMode = true;
    playCardButton.on('pointerdown', onPlayCardButtonClick);
    function onPlayCardButtonClick() {
      websocket.send(JSON.stringify(
        {
          "card": this.hand[0],
          "gameId": "1",
          "playerId": "2",
          "playerName": "alice"
        }
      ));
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

    this.hand = gameState.gameState.hand;
  }

}