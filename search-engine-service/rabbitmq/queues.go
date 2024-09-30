package rabbitmq

// Routing key used by the websocket server
const PUBLISH_QUEUE = "search_poll_queue"

// Routing key used by the express server to send
// a search query to the search engine service
const SEARCH_QUEUE = "search_queue"

const DB_QUERY_QUEUE = "db_query_sengine"
const DB_RESPONSE_QUEUE = "db_cbq_sengine"
