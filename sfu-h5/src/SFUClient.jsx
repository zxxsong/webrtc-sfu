import React from 'react'
import '../styles/css/sfu.scss';
import SFULogin from './SFULogin';
import { Button,List } from 'antd';
import SFU from './SFU';
import Player from './Player';
import AudioDetecter from './AudioDetecter';
import SoundMeter from './SoundMeter'; 

var sfu;
/* var connected = false; */
var connected = true;
var published = false;
var marked_speaker = false;
var players = new Map();

// soundMeterProcess()用到的变量
var startTotal = 0;
var endTotal = 0;
var startThreshold = 3;
var endThreshold = 3;

class SFUClient extends React.Component{

    constructor(props){
        super(props);

        /* this.state = {
            isLogin:false,
            userName:'',
            roomId:'',
            userId:this.getRandomUserId(),
        }; */
        this.state = {
            isLogin:true,
            roomId:'1',
            userId:this.getRandomUserId(),
            userName:''
        };
    }

    onPublishBtnClick = () => {
        if(!connected){
            console.log('客户端还没有连接到服务器...');
            return;
        }
        if(published){
            console.log('已经发布了音视频...');
            return;
        }
        console.log('开始发布音视频...');
        sfu.publish();
        published = true;
    }
    
    // 开启检测主讲人, 暂时不用 AudioDetecter
    // 这里的处理是只对自己处理，下面的sfu.on处理的是别人用户发来的请求
    onMarkBtnClick = () => {
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
            setInterval(soundMeterProcess, 1000);
        })
        .catch(this.onErr);

    }

    soundMeterProcess = () => {
        var val = (window.soundMeter.instant.toFixed(2) * 348) + 1;

        console.log("audioLevel:" + val);
        // 前者为模式判断累加过程，后者为已经进入主讲人模式，需要判断是否结束
        if(val > 5 || marked_speaker) {
            startTotal++;
            console.log("startTotal: " + startTotal + "!!!!!!!!!!!!!!!!!!!!!");
            // startThreshold 和 intervaltime 决定开始阈值的时间
            if(startTotal >= startThreshold) {
                // 主讲人模式开始
                marked_speaker = true;
                console.log("!!!!!!!!!!!!findSpeaker!!!!!!!!!!!!");
                
                // 监测主讲人模式结束
                if(val < 10) {
                    endTotal++;
                    // endThreshold 和 intervaltime 决定结束阈值的时间
                    if(endTotal == endThreshold) {
                        marked_speaker = false;
                        startTotal = 0;
                        endTotal = 0;
                    }
                }
            }
        } else {
            // 累加过程中途断掉则重新开始探测，因为主讲人模式假设为连续的发言
            startTotal = 0;
        }


/* 有bug！！！！！！
在没有发生状态改变时，
仍旧一直向服务器端发送消息 */


        if (marked_speaker) {
            // 通知服务器端进行主讲人标识
            sfu.findMark();
            // 进行自己主讲人标识
            var player = players.get(this.state.userId);
            if(player){
                console.log("player change red");
                player.name.style.color = 'red';
            }
        } else {
            // 通知服务器端结束主讲人标识
            sfu.findEndMark();
            // 结束自己的主讲人标识
            var player = players.get(this.state.userId);
            if(player){
                console.log("player change black");
                player.name.style.color = 'black';
            }
        }
    }
      
    onErr(error) {
        const errorMessage = '报错 navigator.MediaDevices.getUserMedia : ' + error.message + ' ' + error.name;
        console.error(errorMessage);
    }

    onJoinBtnClick = (userName,roomId) => {
        // 固定用户名为其id，房间为1
        userName = this.state.userId
        roomId = "1"
        sfu = new SFU(this.state.userId,roomId);
        sfu.on('connect',() =>{
            console.log('Connected to SFU!');
            connected = true;

            this.setState({
                isLogin:true,
                userName:userName,
                roomId:roomId,
            });
            sfu.join(this.state.userId,this.state.userName,this.state.roomId);

        });

        sfu.on('disconnect',() => {
            connected = false;
        });

        sfu.on('addLocalStream',(id,stream) => {
            var player = new Player({id,stream,parent:'localVideoDiv'});
            players.set(id,player);
        });

        sfu.on('addRemoteStream',(id,stream) => {
            var player = new Player({id,stream,parent:'remoteVideoDiv'});
            players.set(id,player);
        });

        sfu.on('removeRemoteStream',(id,stream) => {
            var player = players.get(id);
            if(player){
                player.destory();
            }
        });

        // 增加主讲人标注事件 处理
        sfu.on('markSpeakerStream', (id, stream) => {
            var player = players.get(id);
            if(player){
                player.name.style.color = 'red';
            }
        });
        
        // 增加主讲人结束标注事件 处理
        sfu.on('endMarkSpeakerStream', (id, stream) => {
            var player = players.get(id);
            if(player){
                player.name.style.color = 'black';
            }
        });
    }

    getRandomUserId() {
        var num = "";
        for (var i = 0; i < 6; i++) {
            num += Math.floor(Math.random() * 10);
        }
        return num;
    }

    render(){
        return(
            <div>
            {!this.state.isLogin ?
                <div className="login-container">
                <h2>SFU-demo</h2>
        
                </div>
                :
                <div>
                    <Button onClick={()=>this.onJoinBtnClick()}>加入会议</Button>
                    <Button onClick={()=>this.onPublishBtnClick()}>发布</Button>
                    <Button onClick={()=>this.onMarkBtnClick()}>监测</Button>
                    <p>本地视频</p>
                    <div id="localVideoDiv"></div>
                    <p>远端视频</p>
                    <div id="remoteVideoDiv"></div>
                </div>
            }
                
            </div>
        );
    }

}

export default SFUClient;