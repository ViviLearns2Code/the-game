
class TestUI extends PIXI.Container {
  constructor() {
    super()

    const style_small = new PIXI.TextStyle({
      fontSize: 12,
    });
    this.msgtext = new PIXI.Text('JSON', style_small);
    this.addChild(this.msgtext);
    this.msgtext.x = 150;

    const style = new PIXI.TextStyle({
      fontSize: 16,
    });

    const testb1 = new PIXI.Text('Readytest', style);
    this.addChild(testb1);
    testb1.y = 0;
    testb1.interactive = true;
    testb1.buttonMode = true;
    testb1.on('pointerdown', onTestB1Click);
    function onTestB1Click() {
      var text = '{ "state":"lobby"}';
      parseGameStateGlobal(JSON.parse(text))
    }

    const testb2 = new PIXI.Text('Playtest', style);
    this.addChild(testb2);
    testb2.y = 15;
    testb2.interactive = true;
    testb2.buttonMode = true;
    testb2.on('pointerdown', onTestB2Click);
    function onTestB2Click() {
      var text = '{ "state":"game"}';
      parseGameStateGlobal(JSON.parse(text))
    }

    const testb3 = new PIXI.Text('Playtest', style);
    this.addChild(testb3);
    testb3.y = 30;
    testb3.interactive = true;
    testb3.buttonMode = true;
    testb3.on('pointerdown', onTestB3Click);
    function onTestB3Click() {
      var text = '{"gameId": "1","playerId": "1","gameState": {"hand": [20,43,61,62,68,90,100],"playersCardCount": [{"bob": 7,"alice": 4}],"topCard": 10,"level": 1,"lives": 3,"stars": 1},"event": {"type": "propose-star","triggeredBy": "bob","proStar": ["bob"],"conStar": [],"discarded": []}}'
      parseGameStateGlobal(JSON.parse(text))
    }

  }

  parseGameState(gameState) {
    this.msgtext.text = JSON.stringify(gameState);
  }

}