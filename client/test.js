import * as PIXI from './pixi.mjs';

export class TestUI extends PIXI.Container {
  constructor(jsonInjectCallback) {
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
      jsonInjectCallback(JSON.parse(text))
    }

    const testb2 = new PIXI.Text('lobby', style);
    this.addChild(testb2);
    testb2.y = 15;
    testb2.interactive = true;
    testb2.buttonMode = true;
    testb2.on('pointerdown', onTestB2Click);
    function onTestB2Click() {
      var text = `
      {
        "errorMsg":"",
        "gameState":{
          "gameToken":"049281ac-104f-4e9b-9cc6-e494e61ecea2",
          "playerToken":"",
          "playerName":"Bla",
          "PlayerId":2,
          "cardsOfPlayer":{
            "cardsInHand":null,
            "nrCardOfOtherPlayers":null
          },
          "playerNames":{
            "1":"Bla"
          },
          "cardsOnTable":{
            "topCard":0,
            "level":0,
            "lives":0,
            "stars":0
          },
          "gameStateEvent":{
            "name":"",
            "levelTitle":"",
            "starsIncrease":false,
            "starsDecrease":false,
            "livesIncrease":false,
            "LivesDecrease":false
          },
          "readyEvent":{
            "name":"lobby",
            "triggeredBy":0,
            "ready":null
          },
          "placeCardEvent":{
            "name":"",
            "triggeredBy":0,
            "discardedCard":null
          },
          "processStarEvent":{
            "name":"",
            "triggeredBy":0,
            "proStar":null,
            "conStar":null}
          }
        }
      `;
      jsonInjectCallback(JSON.parse(text))
    }

    const testb3 = new PIXI.Text('game', style);
    this.addChild(testb3);
    testb3.y = 30;
    testb3.interactive = true;
    testb3.buttonMode = true;
    testb3.on('pointerdown', onTestB3Click);
    function onTestB3Click() {
      var text = `
        {
          "errorMsg": "",
          "gameState": {
            "gameToken": "049281ac-104f-4e9b-9cc6-e494e61ecea2",
            "playerToken": "01w281ac-524f-4w3f-1hh5-i135z71rtre2",
            "playerName": "bob",
            "PlayerId": 1,
            "cardsOfPlayer": {
              "cardsInHand": [40, 53, 88],
              "nrCardOfOtherPlayers": {
                "1": 3,
                "2": 2
              }
            },
            "playerNames": {
              "1": "bob",
              "2": "alice"
            },
            "cardsOnTable": {
              "topCard": 10,
              "level": 3,
              "lives": 2,
              "stars": 1
            },
            "gameStateEvent": {
              "name": "",
              "levelTitle": "",
              "starsIncrease": false,
              "starsDecrease": false,
              "livesIncrease": false,
              "livesDecrease": false
            },
            "readyEvent": {
              "name": "",
              "triggeredBy": 0,
              "ready": null
            },
            "placeCardEvent": {
              "name": "",
              "triggeredBy": 0,
              "discardedCard": null
            },
            "processingStarEvent": {
              "name": "",
              "triggeredBy": 0,
              "proStar": null,
              "conStar": null
            }
          }
        }
      `;
      jsonInjectCallback(JSON.parse(text))
    }

  }

  parseGameState(gameState) {
    this.msgtext.text = JSON.stringify(gameState);
  }

}