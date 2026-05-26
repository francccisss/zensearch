package rabbitmq

// CRAWLER ROUTING KEYS

// Stores webpage to database
const CRAWLER_DB_INDEXING_QUEUE = "crawler.db.indexing.queue"
const DB_CRAWLER_INDEXING_CBQ = "db.crawler.indexing.cbq"

// Frontier Queue
const CRAWLER_DB_STOREURLS_FRONTIER_QUEUE = "crawler.db.storeurls.frontier.queue"
const CRAWLER_DB_CLEARURLS_QUEUE = "crawler.db.clearurls.frontier.queue"

const CRAWLER_DB_DEQUEUE_FRONTIER_QUEUE = "crawler.db.dequeue.frontier.queue"
const DB_CRAWLER_DEQUEUE_FRONTIER_CBQ = "db.crawler.dequeue.frontier.cbq"

// TODO: maybe instead of querying message broker
// instead create a map with root as key, and each node/url
// is stored in the array in the map's value which then
// can be used to pop and push items into it for checking len
// problem is "RACE CONDITION"
const CRAWLER_DB_LEN_FRONTIER_QUEUE = "crawler.db.len.frontier.queue"
const DB_CRAWLER_LEN_FRONTIER_CBQ = "db.crawler.len.frontier.cbq"

// queue used for consuming urls from express server
const EXPRESS_CRAWLER_QUEUE = "express.crawler.crawl.queue"
const CRAWLER_EXPRESS_CBQ = "crawler.express.crawl.cbq"
