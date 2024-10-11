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
  close: "close",
};

class WebsocketService {
  wss: WebSocketServer;
  isAlive: boolean = false;
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
    this.wss.on(EVENTS.connection, async (ws: WebSocket) => {
      console.log("connected");
      ws.on(EVENTS.message, async (data: Data) => {
        if (data.toString() === "pong") {
          // Client cant send a pong to the server so, server will have to handle that
          console.log("Client sent pong");
          ws.pong("ACK");
          return;
        }
        this.isAlive = true;
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
          //this.search_handler(data);
          console.log("Search");
        }
      });

      ws.on("close", () => {
        this.isAlive = false;
        console.log("Connection closed is alive is false");
      });
    });
  }

  /*
    This Callback function is executed right after the handler() has received a new search query
    object from the client, it calls the `send_search_query` to send.. a search query as the name
    implies to the search engine service for processing.
   */
  private async search_handler(data: Data) {
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

  /*
    A callback function that is called within the channel listeners for consuming
    messages from the message queues, it handles the delivery of acknoledgments to
    the rabbitmq message broker based on the connection of the client through
    the websocket connection.

    Need to create a reliable data transfer pipeline for sending acks from client
    that it received the message from the websocket server
    such that the websocket server will be able to send an acknoledgment back
    to the message broker that the message from the message queue
    has been successfully acknowledged


    I just realized theres no way for me to reconnect the client
    using the same ephemeral port that is stored as a client object
    in the server's memory once the client disconnects from the server
    so this will just be requeued indefinitely. Wow
  */
  async send_results_to_client(
    chan: Channel,
    msg: ConsumeMessage,
    message_type: string,
  ) {
    this.wss.clients.forEach((ws) => {
      let isAck = false;
      if (ws.OPEN && msg !== null) {
        console.log("Sending start");
        /* Handling Disconnected clients

         The requeuing of the message back to the message queue
         will call this function again.

         Need to somehow set a timer for on pong event, if the server has not received
         an acknowledgment from the alloted time, then we retransmit the message from message queue

         When nack is called on message queue channel, the message will be requeued.
        */

        // NOT LISTENING
        ws.on("pong", (message) => {
          const ack: "ACK" | "NACK" = message.toString() as "ACK" | "NACK";
          console.log("Server received pong");
          console.log(ack);
          if (ack === "ACK") {
            console.log("Client ACK");
            isAck = true;
            chan.ack(msg);
            return;
          }
          /*
           Client does not need to send a "NACK" message since we can rely on tcp for currupted bits
           handling, but instead we will only handle lossy channel between client and server to
           retransmit messages from the message queue when client suddenly disconnects.
          */
        });

        const responseData = JSON.stringify({
          data_buffer: msg.content,
          message_type,
        });

        // Send data to the client receiver
        ws.send(responseData, (err) => {
          if (err) {
            // This Nack will be labeled as a server | client error.
            // Still dont know how to handle this
            chan.nack(msg, true, false);
            console.error(err);
            throw new Error("ERROR: Unable to send message to client.");
          }
        });

        /*
         Start timer to determine if connection is closed and is unable
         to send an Ack to the websocket sender if client disconnects while
         sending this message.

         This is only for the current message to be labeled as nack and requeued
        */
        setTimeout(() => {
          if (isAck === false) {
            console.log("Timeout, retransmit message");
            chan.nack(msg, false, true);
          }
        }, 3 * 1000);
      }
    });
  }
}

export default WebsocketService;
