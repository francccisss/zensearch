package rabbitmq

// SEARCH ENGINE ROUTING KEYS
// queue for sending search query to search engine from express server
const EXPRESS_SENGINE_QUERY_QUEUE = "express.sengine.query.queue"

// a callback queue to consume from search engine to express server
// after ranking webpages
const SENGINE_EXPRESS_QUERY_CBQ = "sengine.express.query.cbq"

// routing key used by search engine service to request database for webpages.
const SENGINE_DB_REQUEST_QUEUE = "sengine.db.request.queue"

// routing key to reply back to the search engine service's callback queue.
const DB_SENGINE_REQUEST_CBQ = "db.sengine.request.cbq"
