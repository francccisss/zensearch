# search-engine
Building a small scale search engine using NodeJs for Crawling and indexing.

### Why did I build this
I've been studying about operating systems through Modern Operating Systems 4th edition by Andrew S. Tanenbaum, where I learned about threads and processes and Tsoding daily's video about TF-IDF, so I was curious and had to think of a project where I could utilize both concepts.

### What does it do
Basically a simple distributed search engine, developers can crawl and index websites, process them and then store them in an sqlite database, I'm still planning if whether or not I should build this as a microservice or a monolithic project, the reason why is that I wanted to separate CRAWLERS/INDEXING, DATABASE CRUD OPERATIONS, SEARCH ENGINE functionality because NodeJs is a single threaded runtime environment for yavascript, so its much more efficient to use it I/O-Bound operations, so I might use Golang for querying and searching through huge datasets from indexed websites so It would be a good idea to use a more CPU-bound runtime and also make it much more easier to handle race conditions and threads.