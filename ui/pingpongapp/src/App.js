
import './App.css';
import React, {useState, useEffect } from 'react';
import { PingPongClient } from './proto/service_grpc_web_pb';
import { PingRequest } from './proto/service_pb';

import { EchoServiceClient } from './proto/echo_grpc_web_pb';
import { EchoRequest } from './proto/echo_pb';

import { PingPong2Client } from './proto/service2_grpc_web_pb';
import { Ping2Request } from './proto/service2_pb';

  // We create a client that connects to the api
var client = new PingPongClient("https://localhost:8080");
var clientEcho = new EchoServiceClient("https://localhost:8080");

var client2 = new PingPong2Client("https://localhost:8080");

function App() {
  // Create a const named status and a function called setStatus
  const [status, setStatus] = useState(false);
  const [echo, setEcho] = useState(false);
  const [status2, setStatus2] = useState(false);
  // sendPing is a function that will send a ping to the backend
  const sendPing = () => {
    var pingRequest = new PingRequest();
    var echoRequest = new EchoRequest();
    var ping2Request = new Ping2Request();
    // use the client to send our pingrequest, the function that is passed
    // as the third param is a callback. 
    client.ping(pingRequest, null, function(err, response) {
      // serialize the response to an object 
      var pong = response.toObject();
      // call setStatus to change the value of status
       setStatus(pong.ok);
    }); 

    clientEcho.echo(echoRequest, null, function(err, response) {
      // serialize the response to an object 
      var echo = response.toObject();
      // call setStatus to change the value of status
      setEcho(echo.ok);
    });

    client2.ping2(ping2Request, null, function(err, response) {
      // serialize the response to an object 
      var pong2 = response.toObject();
      // call setStatus to change the value of status
       setStatus2(pong2.ok);
    }); 

    console.log(client2, 111);
  }

  useEffect(() => {
    // Start a interval each 3 seconds which calls sendPing. 
    const interval = setInterval(() => sendPing(), 3000)
    return () => {
      // reset timer
      clearInterval(interval);
    }
  },[status]);
  
  // we will return the HTML. Since status is a bool
  // we need to + '' to convert it into a string
  return (
    <div className="App">
      <p>Status: {status + ''}</p>
      <div>///////////////</div>
      <p>Echo: {echo + ''}</p>
      <div>///////////////</div>
      <p>Status2: {status2 + ''}</p>
    </div>
  );


}


export default App;
