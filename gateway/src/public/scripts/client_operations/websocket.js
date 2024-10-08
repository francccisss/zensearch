const ws = new WebSocket("ws://192.168.1.20:8080");

ws.addEventListener("open", (msg) => {
  console.log("Connected to websocket server");
});

ws.addEventListener("message", (event) => {
  console.log(event.data);
});

export default ws;
