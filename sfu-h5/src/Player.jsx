class Player{
    constructor(opt){
        this._create(opt);
    }

    _create({id,stream,parent}){
        let video = document.createElement('video');
        video.class = 'player';
        video.style = 'width:320px;height:240px;';
        video.autoplay = true;
        video.playsinline = true;
        video.controls = true;
        video.muted = true;
        video.srcObject = stream;
        video.id = `stream${id}`;
        this.video = video;
        
        //增加视频流id
        let name = document.createElement('name');
        name.style.fontSize = '20px';
        name.style.color = 'black';
        var name_content = document.createTextNode("ID:" + id);
        name.appendChild(name_content);
        this.name = name;

        let parentElement = document.getElementById(parent);
        parentElement.appendChild(video);
        parentElement.appendChild(name);
        this.parentElement = parentElement;
    }

    destory(){
        this.video.pause();
        this.parentElement.removeChild(this.video);
        this.parentElement.removeChild(this.name);
    }
}

export default Player;