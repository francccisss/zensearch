package rabbitmq

// CRAWLER ROUTING KEYS
// queue used for when crawler is about to save webpages it crawled
const CRAWLER_DB_INDEXING_QUEUE = "crawler_db_indexing_queue"

const CRAWLER_DB_DEQUEUE_URL_QUEUE = "crawler_db_dequeue_url_queue"
const DB_CRAWLER_DEQUEUE_URL_CBQ = "db_crawler_dequeue_url_cbq"

// a callback queue for notifying crawler that storing indexed webpages
// was successful or failure
const DB_CRAWLER_INDEXING_CBQ = "db_crawler_indexing_cbq"

// queue used for consuming urls from express server
const EXPRESS_CRAWLER_QUEUE = "express_crawler_queue"

// a callback queue from express to crawler to notify express server
// about the state of the crawl
const CRAWLER_EXPRESS_CBQ = "crawler_express_cbq"
