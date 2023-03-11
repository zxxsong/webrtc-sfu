import { EventEmitter } from 'events';

const ices = 'stun:stun.stunprotocol.org:3478'

export default class RTC extends EventEmitter {
    constructor(){
        super();
        this._sender = {};
        this._receivers = new Map();
    }

    get sender (){
        return this._sender;
    }

    getReceivers(pubid){
        return this._receivers.get(pubid);
    }

    async createSender(pubid){
        let sender = {
            offerSent:false,
            pc:null,
        };
        sender.pc = new RTCPeerConnection({ iceServers :[{urls:ices}]});
        let stream = await navigator.mediaDevices.getUserMedia({video:true,audio:true});
        sender.pc.addStream(stream);
        this.emit('localstream',pubid,stream);
        this._sender = sender;
        return sender;
    }

    createReciver(pubid){
        try{
            let receiver = {
                offerSent:false,
                pc:null,
                id:pubid,
                streams:[]
            };
            var pc = new RTCPeerConnection({ iceServers :[{urls:ices}]});
            pc.onicecandidate = e => {
                console.log('receiver.pc.onicecandidate =>' + e.candidate);
            }

            pc.addTransceiver('audio',{'direction':'recvonly'});
            pc.addTransceiver('video',{'direction':'recvonly'});

            pc.onaddstream = (e) => {
                var stream = e.stream;
                console.log('receiver.pc.onaddstream',stream.id);
                var receiver = this._receivers.get(pubid);
                receiver.streams.push(stream);
                this.emit('addstream',pubid,stream);
            }

            pc.onremovestream = (e) => {
                var stream = e.stream;
                console.log('receiver.pc.onremovestream',stream.id);
                this.emit('removestream',pubid,stream);
            }

            // // 增加主讲人标注事件
            // pc.onmarkstream = (e) => {
            //     var stream = e.stream;
            //     console.log('recevier.pc.onmarkstream', stream.id);
            //     this.emit('markstream', pubid, stream);
            // }

            receiver.pc = pc;
            console.log("createReceiver::id",pubid);
            this._receivers.set(pubid,receiver);
            return receiver;


        }catch(e){
            console.log(e);
            throw e;
        }
    }

    closeReceiver(pubid){
        var receiver = this._receivers.get(pubid);
        if(receiver){
            receiver.streams.forEach(stream => {
                this.emit('removestream',pubid,stream)
            })
            receiver.pc.close();
            this._receivers.delete(pubid);
        }
    }

    markReceiver(pubid) {
        var receiver = this._receivers.get(pubid);
        // 这个地方没有给自己pubid，发送消息，因为这里recevier存储的都是其他用户的流
        if(receiver){
            receiver.streams.forEach(stream => {
                this.emit('markstream', pubid, stream);
            })
        }
    }

    endMarkReceiver(pubid) {
        var receiver = this._receivers.get(pubid);
        // 这个地方没有给自己pubid，发送消息，因为这里recevier存储的都是其他用户的流
        if(receiver){
            receiver.streams.forEach(stream => {
                this.emit('endmarkstream', pubid, stream);
            })
        }
    }

    
}