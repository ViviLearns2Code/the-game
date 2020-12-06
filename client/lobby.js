
class LobbyUI extends PIXI.Container {
  constructor() {
    super()

    const titleText = new PIXI.Text('The Game\'s Lobby');
    this.addChild(titleText);
    titleText.x = 50;
    titleText.y = 100;

    const joinButton = new PIXI.Text('Leave');
    this.addChild(joinButton);
    joinButton.x = 50;
    joinButton.y = 200;

    joinButton.interactive = true;
    joinButton.buttonMode = true;
    joinButton.on('pointerdown', onJoinButtonClick);
    function onJoinButtonClick() {

    }

    const readyButton = new PIXI.Text('Ready');
    this.addChild(readyButton);
    readyButton.x = 50;
    readyButton.y = 300;

    readyButton.interactive = true;
    readyButton.buttonMode = true;
    readyButton.on('pointerdown', onReadyButtonClick);
    function onReadyButtonClick() {

    }
  }

  parseGameState(gameState) {
    this.visible = false;//gameState.state == "lobby";
    if (!this.visible)
      return

  }
}