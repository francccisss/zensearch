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

export type DatabaseServiceDefinition = {
	exchange: {
		general: string
		crawler: string
	}
	routing_keys: {
		se_db_request: string

		es_db_check: string

		cr_db_indexing: string
		cr_db_enqueue: string
		cr_db_dequeue: string
		cr_db_getlen: string
	}
	queues: {

		se_db_request_queue: string
		es_db_check_queue: string

		cr_db_indexing_queue: string
		cr_db_enqueue_queue: string
		cr_db_dequeue_queue: string
		cr_db_getlen_queue: string
	}
}
