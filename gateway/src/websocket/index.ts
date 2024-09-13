import { WebSocketServer, WebSocket } from "ws";
import http, { Server, ServerOptions } from "http";

import os from "os";

const EVENTS = {
  search: "search",
  connection: "connection",
};

function handler(wss: WebSocketServer) {
  wss.on(EVENTS.connection, (socket: WebSocket) => {
    console.log("connected");
  });
}

export default { handler };
