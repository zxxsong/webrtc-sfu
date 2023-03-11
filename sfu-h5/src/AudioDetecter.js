import { EventEmitter } from 'events';
import SoundMeter from './SoundMeter';

export default class AudioDetecter extends EventEmitter{
    constructor(speaker) {
      super();
      this.isSpeaker = speaker;
    }

    detect () {
      try {
          window.AudioContext = window.AudioContext || window.webkitAudioContext;
          window.audioContext = new AudioContext();
      } catch (e) {
          alert('Web Audio API 不支持.');
      }
      const soundMeter = window.soundMeter = new SoundMeter(window.audioContext);
      let soundMeterProcess = this.soundMeterProcess;

      navigator.mediaDevices.getUserMedia({video:false, audio:true})
      .then(function (stream) {
          window.stream = stream; // make stream available to console
          console.log('对接麦克风的音频');
          soundMeter.connectToSource(stream);
          setTimeout(soundMeterProcess, 1000); 
          
          // 增加监测线程
          // let myworker = new Worker('soundWorker.js');
          // myworker.onmessage = function (events) {
          //     console.log('worker received message ' + events.data);
          //     if(events.data == "findSpeaker"){
          //       this.isSpeaker = true;
          //     }
          //     setInterval(soundMeterProcess, 500);
          // }
            
      })
      .catch(this.onErr);

      return this.isSpeaker;
    }

    soundMeterProcess = () => {
        var val = (window.soundMeter.instant.toFixed(2) * 348) + 1;
        console.log("audioLevel:" + val);
        if(val > 10) {
            this.isSpeaker = true;
            console.log("findSpeaker!!!!!!!!!!!!");
        }
        this.isSpeaker = false;
        // this.setState({ audioLevel: val });
        // if (this.state.visible)
        //     setTimeout(this.soundMeterProcess, 100);
    }
      
    onErr(error) {
        const errorMessage = '报错 navigator.MediaDevices.getUserMedia : ' + error.message + ' ' + error.name;
        console.error(errorMessage);
    }
}

