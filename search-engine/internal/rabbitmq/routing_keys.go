package rabbitmq

// SEARCH ENGINE ROUTING KEYS
// queue for sending search query to search engine from express server
const EXPRESS_SENGINE_QUERY_QUEUE = "express_sengine_query_queue"

// a callback queue to consume from search engine to express server
// after ranking webpages
const SENGINE_EXPRESS_QUERY_CBQ = "sengine_express_query_cbq"

// routing key used by search engine service to request database for webpages.
const SENGINE_DB_REQUEST_QUEUE = "sengine_db_request_queue"

// routing key to reply back to the search engine service's callback queue.
const DB_SENGINE_REQUEST_CBQ = "db_sengine_request_cbq"
