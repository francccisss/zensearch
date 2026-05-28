
import type { Channel, ChannelModel } from "amqplib";
import yaml from "js-yaml"
import fs from "fs"
import type { DatabaseServiceDefinition, RabbitMQDefinitions } from "./rbq_definitions.js"
import EventEmitter from "stream";
import path from "path";
import amqp from "amqplib";
import dbInterface from "../database.js";
import mysql from "mysql2/promise";
import type {
	DequeuedUrl,
	IndexedWebpage,
	URLs,
	Webpage,
} from "../utils/types.js";
import segmentSerializer from "../serializer/segment_serializer.js";


// search engine and database would need to consume highthroughput data
// create dedicated channels
//
// TODO: use worker threads for processing high throughput queues 
// from search engine request as well as crawler request
//
// Data segmentation request and processing and responding is for both search engine 
const cumulativeAckCount = 1000;
class RabbitMQClient {
	connection: null | ChannelModel = null;
	client: this = this;
	highThroughputChannel: Channel | null = null; // receiving search results
	eventsChannel: Channel | null = null;
	lowThroughputChannel: Channel | null = null; // general publishing channel
	eventEmitter: EventEmitter = new EventEmitter();
	definitions: DatabaseServiceDefinition


	constructor(def: DatabaseServiceDefinition) {
		this.definitions = def
	}

	async EstablishConnection(retryCount: number): Promise<RabbitMQClient> {
		if (retryCount > 0) {
			retryCount--;
			try {
				this.connection = await amqp.connect("amqp://localhost:5672");
				console.log(
					`Successfully connected to rabbitmq after ${retryCount} retries`,
				);
				return this;
			} catch (err) {
				await new Promise((resolve) => {
					const timeoutID = setTimeout(() => {
						resolve("Done blocking");
						clearTimeout(timeoutID);
					}, 2000);
				});
				return await this.EstablishConnection(retryCount);
			}
		}
		throw new Error("Shutting down dbInterface server after several retries");
	}


	// initializes the rabbitmq engine by and creating channels, asserting, binding queues and exchanges
	// NOTICE: by AMQP specifications, the commands are delegated to the channels SO, any commands
	// called by the channel does NOT mean that they are connected at all to the channel aside
	// from the channel creation, assertion, and bindings called by the distinct channels
	// are just used for making it clear that what channel will use for communicating with the
	// RabbitMQ Engine.
	//
	// We do not bind ephmeral/temporary queues
	async SetDefinitions() {
		try {
			if (this.connection == null) {
				throw new Error("ERROR: Connection interface is null.");
			}

			this.lowThroughputChannel = await this.connection.createChannel() // ouput channel

			// handling high throughput publishing to search engine from data segmentation 
			this.highThroughputChannel = await this.connection.createChannel() // ouput channel

			this.highThroughputChannel!.prefetch(cumulativeAckCount, false);
			// incoming data can be from, search engine query, frontier queue requests from crawler
			// as well has webpage indexing.
			// and crawl list check request from express server
			this.eventsChannel = await this.connection.createChannel(); // consuming channel

			await this.eventsChannel.assertQueue(this.definitions.queues.se_db_request_queue, { durable: true, exclusive: false })
			await this.eventsChannel.assertQueue(this.definitions.queues.es_db_check_queue, { durable: true, exclusive: false })

			await this.eventsChannel.assertQueue(this.definitions.queues.cr_db_getlen_queue, { durable: true, exclusive: false })
			await this.eventsChannel.assertQueue(this.definitions.queues.cr_db_indexing_queue, { durable: true, exclusive: false })
			await this.eventsChannel.assertQueue(this.definitions.queues.cr_db_enqueue_queue, { durable: true, exclusive: false })
			await this.eventsChannel.assertQueue(this.definitions.queues.cr_db_dequeue_queue, { durable: true, exclusive: false })


			try {
				await this.highThroughputChannel.assertExchange(this.definitions.exchange.general, "direct", { durable: true })
				console.log("General Exchange ok")
			} catch (err: any) {
				throw new Error(err)
			}
			try {
				await this.highThroughputChannel.assertExchange(this.definitions.exchange.crawler, "direct", { durable: true })
				console.log("Crawler Exchange ok")
			} catch (err: any) {
				throw new Error(err)
			}

			await this.eventsChannel.bindQueue(this.definitions.queues.se_db_request_queue, this.definitions.exchange.general, this.definitions.routing_keys.se_db_request)
			await this.eventsChannel.bindQueue(this.definitions.queues.es_db_check_queue, this.definitions.exchange.general, this.definitions.routing_keys.es_db_check)

			await this.eventsChannel.bindQueue(this.definitions.queues.cr_db_getlen_queue, this.definitions.exchange.crawler, this.definitions.routing_keys.cr_db_getlen)
			await this.eventsChannel.bindQueue(this.definitions.queues.cr_db_indexing_queue, this.definitions.exchange.crawler, this.definitions.routing_keys.cr_db_indexing)
			await this.eventsChannel.bindQueue(this.definitions.queues.cr_db_enqueue_queue, this.definitions.exchange.crawler, this.definitions.routing_keys.cr_db_enqueue)
			await this.eventsChannel.bindQueue(this.definitions.queues.cr_db_dequeue_queue, this.definitions.exchange.crawler, this.definitions.routing_keys.cr_db_dequeue)
		} catch (err) {
			const error = err as Error;
			console.error(
				"ERROR: Something went wrong while creating channels.",
			);
			console.error(err);
			throw new Error(error.message);
		}
	}

	SearchEngineHandler(
		pool: mysql.Pool,
	) {

		/*
		  Consumes messages sent by the search engine services to query the dbInterface for
		  the webpages to be ranked and sent to the client.
	      
		  a callback queue is assigned once we consume and query webpages so that we can send
		  it back right after all those process are done.
		*/
		console.log(`database consume queue: ${this.definitions.queues.se_db_request_queue}`)
		this.eventsChannel!.consume(this.definitions.queues.se_db_request_queue, async (data) => {
			if (data === null) throw new Error("No data was pushed.");
			console.log("Received query from search engine")
			try {
				const dataQuery: Webpage[] = await dbInterface.queryWebpages(pool);

				const blen = Buffer.from(JSON.stringify(dataQuery))

				this.eventsChannel!.ack(data);
				const MSS = 100000;

				segmentSerializer.createSegments(
					Buffer.from(JSON.stringify(dataQuery)),
					MSS,
					async (newSegment: Buffer) => {
						this.highThroughputChannel!.publish("",
							data.properties.replyTo, // respond back to this queue from search engine
							newSegment,
							{
								correlationId: data.properties.correlationId,
							},
						);
					},
				);

				console.log({ searchEngineMessage: data.content.toString(), BuffLen: Math.trunc(blen.byteLength / MSS) + 1 });
			} catch (err) {
				const error = err as Error;
				console.error(error.message);
				console.error(error.stack);
				this.eventsChannel!.nack(data, false, false);
			}
		});
	}
	DBCheckHandler(pool: mysql.Pool) {

		/*
		  The `ES_DB_CHECK_QUEUE` routing key used to consumes the messages from the express server
		  for new crawl tasks, its responsibility is to check send
		  the array of crawl tasks from the client to the dbInterface service
		  and query the dbInterface to see if the list of crawl tasks exists
		  on the dbInterface already by checking the indexed_sites TABLE,
		  if it does it returns back the array filtering out the ones that
		  have already been crawled and indexed into the dbInterface, returning
		  the remaining items as uncrawled to be processed by the crawler service.
		 */

		this.eventsChannel!.consume(this.definitions.queues.es_db_check_queue, async (data) => {
			if (data == null) throw new Error("No data was pushed.");
			try {
				console.log(
					"NOTIF: DB service received crawl list to check existing indexed websites.",
				);
				const crawlList: { Docs: Array<string> } = JSON.parse(
					data.content.toString(),
				);
				const unindexedWebsites = await dbInterface.checkIndexedWebpage(
					pool,
					crawlList.Docs,
				);

				console.log(unindexedWebsites)
				const encoder = new TextEncoder();
				const encodedDocs = encoder.encode(
					JSON.stringify({ Docs: unindexedWebsites }),
				).buffer;

				this.eventsChannel!.ack(data);
				const is_sent = this.lowThroughputChannel!.publish("",
					data.properties.replyTo,
					Buffer.from(encodedDocs),

				);
				if (!is_sent) {
					console.error("ERROR: Unable to send back message.");
					return
				}
				console.log("Sent crawlist")
			} catch (err) {
				const error = err as Error;
				this.eventsChannel!.nack(data, false, false);
				console.error(error.message);
				console.error(err);
			}
		});

	}


	CrawlerHandler(
		pool: mysql.Pool,
	) {

		// Crawler indexing handler
		this.eventsChannel!.consume(this.definitions.queues.cr_db_indexing_queue, async (data) => {
			// who is going to catch this error? aaaaaa
			if (data === null) throw new Error("No data was pushed.");
			const decoder = new TextDecoder();
			const decodedData = decoder.decode(data.content as unknown as ArrayBuffer);
			const deserializeData: IndexedWebpage = JSON.parse(decodedData);
			try {
				this.eventsChannel!.ack(data);
				await dbInterface.saveWebpage(pool, deserializeData);
				console.log("Storing data");

				// todo: use replyto queue to notify the express server of what is going on
				// currently this there is no consumer for this queue
				this.lowThroughputChannel!.publish("",
					data.properties.replyTo,
					Buffer.from(
						JSON.stringify({
							isSuccess: true,
							Message: "Successfully stored webpages",
							URLSeed: deserializeData.URLSeed,
						}),
					),
				);
			} catch (err) {
				const error = err as Error;
				console.error("ERROR: %s", error);
				console.log("Sending back response to crawler");
				console.log(deserializeData);
				// TODO: USE REPLYTO Queue to notify the express server of what is going on
				// currently this there is no consumer for this queue
				this.lowThroughputChannel!.publish("",
					data.properties.replyTo,
					Buffer.from(
						JSON.stringify({
							IsSuccess: false,
							Message: "Unable to store indexed webpages to sqlite",
							URLSeed: deserializeData.URLSeed,
						}),
					),
				);
				throw Error(error.message);
			}
		});


		// Crawler frontier queues handlers
		this.eventsChannel!.consume(this.definitions.queues.cr_db_dequeue_queue,
			async (msg: amqp.ConsumeMessage | null) => {

				if (msg == null) {
					throw new Error("Message received is null");
				}

				try {
					this.eventsChannel!.ack(msg);
					console.log("dbInterface TEST: DEQUEUEING");
					const domain = msg.content.toString();
					const { length, url, inProgressNode, message } =
						await dbInterface.dequeueURL(pool, domain);

					const dequeuedUrl: DequeuedUrl = { RemainingInQueue: length, Url: url };
					const msgBuffer = Buffer.from(JSON.stringify(dequeuedUrl));

					const sent = this.lowThroughputChannel!.publish("",
						msg.properties.replyTo,
						msgBuffer,
					);
					if (!sent) {
						throw new Error("Error: Unable to send a dequeueded URL");
					}

					console.log("TEST dbInterface: DEQUEUED NODE ", dequeuedUrl);
					console.log("TEST dbInterface: DEQUEUE MESSAGE: %s", message);
					console.log(inProgressNode);
					// node can be null if queue is empty
					if (inProgressNode !== null) {
						await dbInterface.setNodeToVisited(pool, inProgressNode);
						console.log("Node updated to visited, remove in_progress node.");
					}

					if (inProgressNode == null && length == 0) {
						console.log("dbInterface TEST: REMOVED QUEUE");
						await dbInterface.removeQueue(pool, domain);
					}
				} catch (e: any) {
					console.error(e);
					this.eventsChannel!.nack(msg)
					throw new Error(e);
				}
			},
		);
		// doensnt really send any message back to crawler enqueue cbq
		this.eventsChannel!.consume(
			this.definitions.queues.cr_db_enqueue_queue,
			async (msg: amqp.ConsumeMessage | null) => {
				if (msg == null) {
					throw new Error("Message received is null");
				}
				try {
					const URLs: URLs = JSON.parse(msg.content.toString());
					console.log("dbInterface TEST: URLS ", URLs);
					await dbInterface.enqueueUrls(pool, URLs);
					this.eventsChannel!.ack(msg);
				} catch (e: any) {
					console.error(e);
					this.eventsChannel!.nack(msg)
					throw new Error(e);
				}
			},
		);

		this.eventsChannel!.consume(
			this.definitions.queues.cr_db_getlen_queue,
			async (msg: amqp.ConsumeMessage | null) => {
				if (msg == null) {
					throw new Error("Message received is null");
				}
				try {
					this.eventsChannel!.ack(msg);
					const hostname = msg.content.toString();
					const queueLen = await dbInterface.getCurrentQueueLen(pool, hostname);

					const queueLenBuf = Buffer.alloc(4);
					queueLenBuf.writeIntLE(queueLen, 0, 4);
					console.log("dbInterface TEST: QUEUE LEN BUFFER ", queueLenBuf);
					console.log("dbInterface TEST: QUEUE='%s' LENGTH ", hostname, queueLen);

					this.lowThroughputChannel!.publish("", msg.properties.replyTo, queueLenBuf);
				} catch (e: any) {
					console.error(e);
					this.eventsChannel!.nack(msg)
					throw new Error(e);
				}
			},
		);
	}

}


const yamlfile = fs.readFileSync(path.join(import.meta.dirname, "../../../rabbitmq.yml"), "utf8")
const doc = yaml.load(yamlfile, {}) as unknown as RabbitMQDefinitions
const expressDef: DatabaseServiceDefinition = {
	exchange: {
		...doc.exchange
	}
	,
	queues: { ...doc.queues["express_server_queues"] as any, ...doc.queues["search_engine_queues"] as any, ...doc.queues["crawler_queues"] as any },
	routing_keys: { ...doc.routing_keys["express_server_keys"] as any },

}
console.log("Starting express server");
console.log(`Express Definitions: ${JSON.stringify(expressDef)}`)

const rabbitClient = new RabbitMQClient(expressDef);
export default rabbitClient

