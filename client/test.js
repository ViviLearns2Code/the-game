
class TestUI extends PIXI.Container {
  constructor() {
    super()

    const testb1 = new PIXI.Text('Readytest');
    this.addChild(testb1);
    testb1.y = 0;
    testb1.interactive = true;
    testb1.buttonMode = true;
    testb1.on('pointerdown', onTestB1Click);
    function onTestB1Click() {
      var text = '{ "state":"lobby"}';
      parseGameStateGlobal(text)
    }

    const testb2 = new PIXI.Text('Playtest');
    this.addChild(testb2);
    testb2.y = 25;
    testb2.interactive = true;
    testb2.buttonMode = true;
    testb2.on('pointerdown', onTestB2Click);
    function onTestB2Click() {
      var text = '{ "state":"game"}';
      parseGameStateGlobal(text)
    }
  }

}