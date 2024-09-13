import { WebSocketServer, WebSocket, RawData, Data } from "ws";
import http, { Server, ServerOptions } from "http";
import os from "os";
import rabbitmq from "../rabbitmq";
import { Channel, ConsumeMessage } from "amqplib";
import { SEARCH_QUEUE, SEARCH_QUEUE_CB } from "../rabbitmq/queues";

const EVENTS = {
  message: "message",
  connection: "connection",
};

class WebsocketService {
  wss: WebSocketServer;
  constructor(WS: WebSocketServer) {
    this.wss = WS;
  }

  async handler() {
    this.wss.on(EVENTS.connection, (client_ws: WebSocket) => {
      console.log("connected");
      client_ws.on(EVENTS.message, (data: Data) => {
        this.message_handler(data);
      });
    });
  }

  private async message_handler(data: Data) {
    const message: { q: string; job_id: string } = JSON.parse(data.toString());
    console.log("Messaged");
    try {
      await rabbitmq.search_job({ q: message.q, job_id: message.job_id });
      console.log("Message received from client: %s", message.toString());
    } catch (err) {
      const error = err as Error;
      console.log(
        "LOG: Something went wrong while processing message from websocket.",
      );
      console.error(error.message);
    }
  }

  async send_results_to_client(data: ConsumeMessage | null) {
    this.wss.clients.forEach((ws) => {
      if (ws.OPEN && data !== null) {
        console.log(data);
        ws.send(data?.content, (err) => {
          if (err) throw new Error("Unable to send search results to client.");
        });
      }
    });
  }
}

export default WebsocketService;
