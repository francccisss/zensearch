package rabbitmq

// CRAWLER ROUTING KEYS
// queue used for when crawler is about to save webpages it crawled
const CRAWLER_DB_INDEXING_NOTIF_QUEUE = "crawler_db_notif_indexing_queue"

// a callback queue for notifying crawler that storing indexed webpages
// was successful or failure
const DB_CRAWLER_INDEXING_NOTIF_CBQ = "db_crawler_indexing_notif_cbq"

// queue used for consuming urls from express server
const EXPRESS_CRAWLER_QUEUE = "express_crawler_queue"

// a callback queue from express to crawler to notify express server
// about the state of the crawl
const CRAWLER_EXPRESS_CBQ = "crawler_express_cbq"
