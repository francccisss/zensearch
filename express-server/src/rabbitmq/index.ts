import type { Channel, Connection, ConsumeMessage } from "amqplib";
import yaml from "js-yaml"
import fs from "fs"
import type { ExpressServerDefinition, RabbitMQDefinitions } from "./rbq_definitions.js"
import amqp from "amqplib";
import EventEmitter from "stream";
import CircularBuffer from "../segments/circular_buffer.js";
import type { CrawlMessageStatus } from "../types/index.js";
import path from "path";

// TODO ADD LOGS TO RECEIVED AND PROCESSED SEGMENTS
class RabbitMQClient {
	connection: null | Connection = null;
	client: this = this;
	publishChannel: Channel | null = null;
	eventsChannel: Channel | null = null;
	crawlChannel: Channel | null = null;
	searchChannel: Channel | null = null;
	eventEmitter: EventEmitter = new EventEmitter();
	circleBuffer: CircularBuffer = new CircularBuffer(100);
	definitions: ExpressServerDefinition


	constructor(def: ExpressServerDefinition) {
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
				console.error("Retrying Web server service connection");
				await new Promise((resolve) => {
					const timeoutID = setTimeout(() => {
						resolve("Done blocking");
						clearTimeout(timeoutID);
					}, 2000);
				});
				return await this.EstablishConnection(retryCount);
			}
		}
		throw new Error("Shutting down web server after several retries");
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
			this.publishChannel = await this.connection.createChannel() as Channel
			// consumer channels
			this.eventsChannel = await this.connection.createChannel() as Channel;
			// SearchChannel consumes from queue with heavier workload
			this.searchChannel = await this.connection.createChannel() as Channel;
			this.crawlChannel = await this.connection.createChannel() as Channel;

			// EXCHANGE ASSERTION
			await this.publishChannel.assertExchange(this.definitions.exchange.general, "direct", { durable: true })

			// QUEUE ASSERTIONS

			// DB,EXPRESS & SEARCH SPECIFIC TASK
			await this.eventsChannel.assertQueue(this.definitions.queues.es_db_check_queue)
			await this.eventsChannel.assertQueue(this.definitions.queues.es_db_check_cbq,
				{
					exclusive: true,
					durable: false,
				}
			)

			await this.eventsChannel.assertQueue(this.definitions.queues.es_cr_request_queue)
			await this.eventsChannel.assertQueue(this.definitions.queues.es_cr_request_cbq,
				{
					exclusive: true,
					durable: false,
				}
			)
			// SEARCH SPECIFIC TASK
			await this.searchChannel.assertQueue(this.definitions.queues.es_se_query_queue)
			await this.searchChannel.assertQueue(this.definitions.queues.es_se_query_cbq,
				{
					exclusive: true,
					durable: false,
				}
			)

			// QUEUE BINDING
			await this.publishChannel.bindQueue(this.definitions.queues.es_se_query_queue, this.definitions.exchange.general, this.definitions.routing_keys.es_se_query)
			await this.publishChannel.bindQueue(this.definitions.queues.es_cr_request_queue, this.definitions.exchange.general, this.definitions.routing_keys.es_cr_request)
			await this.publishChannel.bindQueue(this.definitions.queues.es_db_check_queue, this.definitions.exchange.general, this.definitions.routing_keys.es_db_check)

		} catch (err) {
			const error = err as Error;
			console.error(
				"ERROR: Something went wrong while creating search channels.",
			);
			console.error(err);
			throw new Error(error.message);
		}
	}

	/*
	 This Function sends a new search query to the SEARCH ENGINE SERVICE.
	 returns a bool to check if it has been sent
	*/
	async SendSearchQuery(q: string): Promise<boolean> {
		if (this.publishChannel == null) {
			throw new Error("ERROR: Search Channel is null");
		}
		return this.publishChannel.publish(this.definitions.exchange.general,
			this.definitions.routing_keys.es_se_query,
			Buffer.from(q),
			{
				replyTo: this.definitions.queues.es_se_query_cbq,
			},
		);
	}

	// cb handles data ack
	async CrawlChannelHandler(
		cb: (chan: Channel, data: ConsumeMessage, message_type: string) => void,
	) {
		if (this.crawlChannel == null)
			throw new Error("ERROR: Crawl Channel is null.");

		try {
			this.crawlChannel.consume(this.definitions.queues.es_cr_request_cbq, async (msg) => {
				if (msg === null) throw new Error("No Response");
				console.log("LOG: Message received from crawler");
				if (this.crawlChannel == null) {
					throw new Error("ERROR: Crawl Channel is null.");
				}
				const deserializedMessage = msg.content.toString();
				const crawlMessage: CrawlMessageStatus =
					JSON.parse(deserializedMessage);
				console.log(crawlMessage);
				//this.crawlChannel.ack(msg, true);
				cb(this.crawlChannel, msg, "crawling");
			});
		} catch (err) {
			const error = err as Error;
			console.log("LOG: Something went wrong with the channel listeners");
			console.error(error.name);
			console.error(error.message);
		}
	}

	async SearchChannelHandler() {
		try {
			this.searchChannel!.consume(
				this.definitions.queues.es_se_query_cbq,
				(data: ConsumeMessage | null) => {
					if (data === null) {
						this.eventsChannel!.close();
						console.error("Msg does not exist");
						this.eventEmitter.emit("segmentError", {
							data: null,
							err: new Error(
								"Something went wrong while listening to segments",
							),
						});
						return;
					}
					this.eventEmitter.emit("newSegment", { data, err: null });
				},
				{ noAck: false },
			);
		} catch (err) {
			console.error(err);
		}
	}

	async AddSegmentsToQueue() {
		this.eventEmitter.on("newSegment", (segment) => {
			this.circleBuffer.write(segment);
		});
	}

	/*
	 Everytime a new segment arrives from the search engine the promise is resolved
	 within the callback function of the event listener, this is synchronous
	 in terms of data being perserved within the segmentGenerator.
	*/

	async *SegmentGenerator() {
		while (true) {
			if (this.circleBuffer.inUseSize() > 0) {
				yield this.circleBuffer.read();
			} else {
				await new Promise((resolve) => setImmediate(resolve));
			}
		}
	}

	// Crawler Expects an Object to be unmarshalled where the array of websites are inside a Docs property.
	async Crawl(
		websites: Buffer,
		job: { queue: string; id: string },
	): Promise<boolean> {
		if (this.connection === null)
			throw new Error("ERROR: TCP Connection lost.");
		try {
			const success = this.publishChannel!.publish(this.definitions.exchange.general, this.definitions.routing_keys.es_cr_request, websites, {
				replyTo: this.definitions.queues.es_cr_request_cbq,
				correlationId: job.id,
			});
			if (!success) {
				throw new Error("ERROR: Unable to send to job to crawler service");
			}
			return true;
		} catch (err) {
			const error = err as Error;
			console.error("ERROR: Something went wrong while starting the crawl.");
			console.error(error.message);
			return false;
		}
	}

	/*
	    Sends a message to the database service to check and see if the DOCS
	    or list of websites the users want to crawl already exists in the database.
	  */
	async DBCheckHandler(
		encoded_list: Uint8Array
	): Promise<{
		unindexed: Array<string>
	}> {
		this.publishChannel!.publish(this.definitions.exchange.general, this.definitions.routing_keys.es_db_check, Buffer.from(encoded_list), {
			replyTo: this.definitions.queues.es_db_check_cbq,
		});

		return new Promise(async (resolve, reject) => {
			await this.eventsChannel!.consume(this.definitions.queues.es_db_check_cbq, async (data) => {
				if (data === null) {
					throw new Error("ERROR: Data received is null.");
				}
				try {
					const parseList: { Docs: Array<string> } = JSON.parse(
						data.content.toString(),
					);
					console.log(parseList)
					this.eventsChannel!.ack(data);
					resolve({ unindexed: parseList.Docs })
				} catch (err) {
					this.eventsChannel!.nack(data, false, false);
					reject(err)
				}
			});

		})

	}

}


const yamlfile = fs.readFileSync(path.resolve(import.meta.dirname, "../../../rabbitmq.yml"), "utf8")
const doc = yaml.load(yamlfile, {}) as unknown as RabbitMQDefinitions
const expressDef: ExpressServerDefinition = {
	exchange: {
		...doc.exchange
	}
	,
	queues: { ...doc.queues["express_server_queues"] as any },
	routing_keys: { ...doc.routing_keys["express_server_keys"] as any },

}
console.log(expressDef.exchange)
console.log(expressDef.queues)
console.log(expressDef.routing_keys)
console.log("Starting express server");

const rabbitClient = new RabbitMQClient(expressDef);
export default rabbitClient
