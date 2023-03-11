import React from 'react';
import {HashRouter as Router,Route} from 'react-router-dom';
import SFUClient from './SFUClient';

class App extends React.Component{
    

    render(){
        return <Router>
            <div>
                <SFUClient/>
            </div>
        </Router>
    }

}

export default App;