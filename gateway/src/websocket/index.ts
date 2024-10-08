import { WebSocketServer, WebSocket, RawData, Data } from "ws";
import http, { Server, ServerOptions } from "http";
import os from "os";
import rabbitmq from "../rabbitmq";
import { Channel, ConsumeMessage } from "amqplib";
import {
  CRAWL_QUEUE,
  SEARCH_QUEUE,
  SEARCH_QUEUE_CB,
} from "../rabbitmq/routing_keys";

const EVENTS = {
  message: "message",
  connection: "connection",
};

class WebsocketService {
  wss: WebSocketServer;
  constructor(WS: WebSocketServer) {
    this.wss = WS;
  }

  /*
    Listens to new tcp connections using websocket protocol and receives/listens
    to new client messages that are pushed to the websocket server.
  */

  // Handler needs to mutliplex incoming messages from the client
  // if it a crawl request or a search request.

  async handler() {
    this.wss.on(EVENTS.connection, async (client_ws: WebSocket) => {
      console.log("connected");
      client_ws.on(EVENTS.message, async (data: Data) => {
        console.log("Message received");

        const decode_buffer: {
          message_type: "crawling" | "searching";
          meta: { job_id: string };
          unindexed_list?: Array<string>;
          [key: string]: any;
        } = JSON.parse(data.toString());

        if (decode_buffer.message_type === "crawling") {
          const serialize_list = Buffer.from(
            JSON.stringify({ Docs: decode_buffer.unindexed_list! }),
          );
          client_ws.send(JSON.stringify(decode_buffer));
          const success = await rabbitmq.client.crawl(serialize_list, {
            queue: CRAWL_QUEUE,
            id: decode_buffer.meta.job_id,
          });
          if (!success) {
            throw new Error(
              "Unable to send crawl list to web crawler service.",
            );
          }
        }
        if (decode_buffer.message_type === "searching") {
          //this.message_handler(data);
          console.log("Search");
        }
      });
    });
  }

  /*
    This Callback function is executed right after the handler() has received a new search query
    object from the client, it calls the `send_search_query` to send.. a search query as the name
    implies to the search engine service for processing.
   */
  private async message_handler(data: Data) {
    const { q, job_id }: { q: string; job_id: string } = JSON.parse(
      data.toString(),
    );
    try {
      const is_sent = await rabbitmq.client.send_search_query({
        q,
        job_id,
      });
      if (!is_sent) {
        throw new Error("ERROR: Unable to send search query.");
      }
      console.log(
        "NOTIF: Message received from client: { search query: %s, job id: %s",
        q,
        job_id,
      );
      console.log("NOTIF: Search query sent to the search engine service.");
    } catch (err) {
      // TODO need to send an error back to the user.
      const error = err as Error;
      console.log(
        "LOG: Something went wrong while processing message from websocket.",
      );
      console.error(error.message);
    }
  }

  // send something back to clients if an error occured maybe
  async send_results_to_client(data: ConsumeMessage | null) {
    this.wss.clients.forEach((ws) => {
      if (ws.OPEN && data !== null) {
        const webpages = data.content.toString();
        ws.send(webpages, (err) => {
          if (err) throw new Error("Unable to send search results to client.");
        });
      }
    });
  }
}

export default WebsocketService;
