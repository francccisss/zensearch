import { WebSocketServer, WebSocket, RawData, Data } from "ws";
import http, { Server, ServerOptions } from "http";
import os from "os";
import rabbitmq from "../rabbitmq";

const EVENTS = {
  message: "message",
  connection: "connection",
};

function handler(wss: WebSocketServer) {
  wss.on(EVENTS.connection, (client_ws: WebSocket) => {
    console.log("connected");
    client_ws.on(EVENTS.message, (data: Data) => {
      message_handler(data, client_ws);
    });
  });
}

async function message_handler(data: Data, client_ws: WebSocket) {
  // job cookies already set, beforehand

  const connection = await rabbitmq.connect();
  const message = JSON.parse(data.toString());
  console.log(message);
  try {
    if (connection === null) throw new Error("TCP Connection lost.");
    //await rabbitmq.search_job({ q: message.q, job_id: message.id }, connection);
    console.log("Message received from client: %s", message.toString());
  } catch (err) {
    const error = err as Error;
    console.log(
      "LOG: Something went wrong while processing message from websocket.",
    );
    console.error(error.message);
  }
}

export default { handler };
