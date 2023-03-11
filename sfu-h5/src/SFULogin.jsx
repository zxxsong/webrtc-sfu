import React from 'react'

import { Form,Input,Button} from 'antd';

class SFULogin extends React.Component{

    constructor(props){
        super(props);
    }

    handleSubmit = (values) => {
        console.log(values);
        this.props.loginHandler(values.userName,values.roomId);
    };

    render(){
        return(
          <Form onFinish={this.handleSubmit} className="login-form">
              <Form.Item name="userName">
                  <Input placeholder="请输入用户名"/>
              </Form.Item>
              <Form.Item name="roomId">
                  <Input placeholder="请输入房间号"/>
              </Form.Item>
              <Form.Item>
                  <Button type="primary" htmlType="submit" className="login-join-button">
                      登录
                  </Button>
              </Form.Item>
          </Form>  
        );
    }

}

export default SFULogin;