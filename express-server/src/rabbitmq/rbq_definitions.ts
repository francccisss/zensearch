/*
 * Message queue convention
 * <source>_<destination>_<function (optional can be chained for more context at the cost of verbosity)>_<queue|cbq (callback queue)>
 *
 * since each endpoint does not implement any fan-out message to multiple
 * services, its easier to isolate the destination in this naming convention
 *
 * that's just for me.
 */

export type RabbitMQDefinitions = {
	exchange: {
		general: string
		crawler: string
	}
	routing_keys: {
		[key: string]: {
			[key: string]: string
		}
	}
	queues: {
		[key: string]: {
			[key: string]: string
		}
	}
}

export type ExpressServerDefinition = {
	exchange: {
		general: string
		crawler: string
	}
	routing_keys: {
		es_se_query: string
		es_db_check: string
		es_cr_request: string
	}
	queues: {

		es_se_query_queue: string
		es_se_query_cbq: string

		es_db_check_queue: string
		es_db_check_cbq: string

		es_cr_request_queue: string
		es_cr_request_cbq: string
	}
}


// CRAWLER ROUTING KEYS
const EXPRESS_CRAWLER_CRAWL_QUEUE = "express.crawler.crawl.queue"
const CRAWLER_EXPRESS_CRAWL_CBQ = "crawler.express.crawl.cbq"

// SEARCH ENGINE ROUTING KEYS
const EXPRESS_SENGINE_QUERY_QUEUE = "express.sengine.query.queue";
const SENGINE_EXPRESS_QUERY_CBQ = "sengine.express.query.cbq";

// EXPRESS ROUTING KEYS
const EXPRESS_DB_CHECK_QUEUE = "express.db.check.queue";
const DB_EXPRESS_CHECK_CBQ = "db.express.check.cbq";

export {
	EXPRESS_CRAWLER_CRAWL_QUEUE,
	CRAWLER_EXPRESS_CRAWL_CBQ,
	EXPRESS_DB_CHECK_QUEUE,
	EXPRESS_SENGINE_QUERY_QUEUE,
	DB_EXPRESS_CHECK_CBQ,
	SENGINE_EXPRESS_QUERY_CBQ,
};
