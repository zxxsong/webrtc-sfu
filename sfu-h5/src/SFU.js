import { EventEmitter } from 'events';
import RTC from './RTC';

export default class SFU extends EventEmitter{

    constructor(userId,roomId){
        super();
        this._rtc = new RTC();

        var sfuUrl = "wss://localhost:3000/ws?userId=" + userId + "&roomId=" + roomId;

        this.socket = new WebSocket(sfuUrl);

        this.socket.onopen = () => {
            console.log("WebSocket连接成功...");
            this._onRoomConnect();
        };

        this.socket.onmessage = (e) => {
            var parseMessage = JSON.parse(e.data);
            
            // 服务器发送给客户端的消息
            switch(parseMessage.type){
                case 'joinRoom':
                    console.log(parseMessage);
                    break;
                case 'onJoinRoom':
                    console.log(parseMessage);
                    break;   
                case 'onPublish':
                    this.onPublish(parseMessage);
                    break;
                case 'onUnpublish':
                    this.onUnpublish(parseMessage);
                    break;
                case 'onSubscribe':
                    this.onSubscribe(parseMessage);
                    break; 
                case 'mark':          //增加标记主讲人数据包
                    this.mark(parseMessage);
                    break;
                case 'endMark':          //增加标记主讲人数据包
                    this.endMark(parseMessage);
                    break;  
                case 'heartPackage':
                    console.log("heartPackage:::");
                    break; 
                default:
                    console.error('未知消息',parseMessage);
            }
        };

        this.socket.onerror = (e) => {
            console.log('onerror::' + e.data);
        };

        this.socket.onclose = (e) => {
            console.log('onclose::' + e.data);
        };
    }

    // RTC.js的消息，上传至sfu.js这一层，这一层再传给sfuclient.jsx
    _onRoomConnect = () => {
        console.log('onRoomConnect');

        this._rtc.on('localstream',(id,stream) => {
            this.emit('addLocalStream',id,stream);
        })

        this._rtc.on('addstream',(id,stream) => {
            this.emit('addRemoteStream',id,stream);
        })

        this._rtc.on('removestream',(id,stream) => {
            this.emit('removeRemoteStream',id,stream);
        })

        // 增加主讲人标注事件
        this._rtc.on('markstream',(id,stream) => {
            this.emit('markSpeakerStream',id,stream);
        })

        this._rtc.on('endmarkstream',(id,stream) => {
            this.emit('endMarkSpeakerStream',id,stream);
        })

        this.emit('connect');
    }

    join(userId,userName,roomId){
        console.log('Join to [' + roomId + ']');
        this.userId = userId;
        this.userName = userName;
        this.roomId = roomId;

        let message = {
            'type':'join',
            'data':{
                'userName':userName,
                'userId':userId,
                'roomId':roomId,
            }
        };
        this.send(message);

    }

    // 客户端发现自己为主讲人时，发送消息给服务器端
    findMark() {
        console.log('dectect speaker is me :[' + this.userId + ']');

        let message = {
            'type':'findMark',
            'data':{
                'userName':this.userName,
                'userId':this.userId,
                'roomId':this.roomId,
                'pubid':this.pubid
            }
        };
        this.send(message);
    }

    // 客户端发现自己结束主讲人身份时，发送消息给服务器端
    findEndMark() {
        console.log('find end speaker is me :[' + this.userId + ']');

        let message = {
            'type':'findEndMark',
            'data':{
                'userName':this.userName,
                'userId':this.userId,
                'roomId':this.roomId,
                'pubid':this.pubid
            }
        };
        this.send(message);
    }

    send = (data) => {
        this.socket.send(JSON.stringify(data));
    }


    publish(){
        console.log('publish stream :[' + this.userId + ']');
        this._createSender(this.userId);
    }

    async _createSender(pubid){

        try{
            //创建一个sender
            let sender = await this._rtc.createSender(pubid);
            this.sender = sender;

            //监听IceCandidate回调
            sender.pc.onicecandidate = async (e) => {
                if(!sender.senderOffer){
                    var offer = sender.pc.localDescription;
                    sender.senderOffer = true;
                   await this.publishToServer(offer,pubid);
                }
            }
            //创建Offer
            let desc = await sender.pc.createOffer({ offerToReceiveVideo:false,offerToReceiveAudio:false})
            sender.pc.setLocalDescription(desc);

        }catch(error){
            console.log('onCreateSender error =>' + error);
        }

    }

    async publishToServer(offer,pubid){
        let message = {
            'type':'publish',
            'data':{
                'jsep':offer,
                'pubid':pubid,
                'userName':this.userName,
                'userId':this.userId,
                'roomId':this.roomId,
            }
        };
        this.send(message);
    }

    onPublish(message){

        //服务器返回的Answer信息 如A--->Offer--->SFU--->Answer--->A
        if(this.sender && message['data']['userId'] == this.userId){
            console.log('onPublish:::自已发布的Id:::' + message['data']['userId']);
            this.sender.pc.setRemoteDescription(message['data']['jsep']);
        }

        //服务器返回其他人发布的信息 如A--->Pub--->SFU--->B
        if(message['data']['userId'] != this.userId){
            console.log('onPublish:::其他人发布的Id:::' + message['data']['userId']);
            //使用发布者的userId创建Receiver
            this._onRtcCreateReceiver(message['data']['userId']);
        }

    }

    onUnpublish(message){
        console.log('退出用户:'+message['data']['pubid']);
        this._rtc.closeReceiver(message['data']['pubid']);
    }

    // 收到服务器发送消息，标识房间中的某个用户为主讲人
    mark(message){
        console.log('标记主讲人:' + message['data']['pubid']);
        this._rtc.markReceiver(message['data']['pubid']);
    }

    // 收到服务器发送消息，房间中主讲人标记结束
    endMark(message){
        console.log('主讲人结束:' + message['data']['pubid']);
        this._rtc.endMarkReceiver(message['data']['pubid']);
    }

    async _onRtcCreateReceiver(pubid) {
        try{
            let receiver = await this._rtc.createReciver(pubid);
           
            receiver.pc.onicecandidate = async (e) => {
                if(!receiver.senderOffer){
                    var offer = receiver.pc.localDescription;
                    receiver.senderOffer = true;
                    await this.subscribeFromServer(offer,pubid);
                }
            }
            //创建Offer
            let desc = await receiver.pc.createOffer();
            receiver.pc.setLocalDescription(desc);

        }catch(error){
            console.log('onRtcCreateReceiver error =>' + error);
        }
    }

    
    async subscribeFromServer(offer,pubid){
        let message = {
            'type':'subscribe',
            'data':{
                'jsep':offer,
                'pubid':pubid,
                'userName':this.userName,
                'userId':this.userId,
                'roomId':this.roomId,
            }
        };
        this.send(message);
    }


    onSubscribe(message){
        //使用发布者的Id获取Receiver
        var receiver = this._rtc.getReceivers(message['data']['pubid']);
        if(receiver){
            console.log('服务器应答Id:' + message['data']['pubid']);
            receiver.pc.setRemoteDescription(message['data']['jsep']);
        }else{
            console.log('receiver == null');
        }
    }

}