# search-engine
Building a small scale Distributed Search Engine using NodeJs for Crawling and indexing and Golang for calculating and ranking each webpages' relevancy to the user's search query.

## Why did I build this
I've been studying about operating systems through Modern Operating Systems 4th edition by Andrew S. Tanenbaum, where I learned about threads and processes and Tsoding daily's video about TF-IDF, so I was curious and had to think of a project where I could utilize both concepts.

## What does it do
Basically a simple distributed search engine, developers can crawl and index websites, process them and then store them in an sqlite database, I'm still planning if whether or not I should build this as a microservice or a monolithic project, the reason why is that I wanted to separate CRAWLERS/INDEXING, DATABASE CRUD OPERATIONS, SEARCH ENGINE functionality because NodeJs is a single threaded runtime environment for yavascript, so its much more efficient to use it I/O-Bound operations, so I might use Golang for querying and searching through huge datasets from indexed websites so It would be a good idea to use a more CPU-bound runtime and also make it much more easier to handle race conditions and threads.



# Tools and Dependencies

#### Frontend
[No Framework just vanilla yavascript](https://frontendmasters.com/blog/you-might-not-need-that-framework/)

#### Backend Server
[ExpressJs](http://expressjs.com/)
[NodeJS](https://nodejs.org/en)

#### Database Service
[ExpressJs](http://expressjs.com/)
[NodeJS](https://nodejs.org/en)
[Sqlite3 for Nodejs](https://www.npmjs.com/package/sqlite3)
[Sqlite](https://www.sqlite.org/index.html)

#### Crawler Service
[Go](https://go.dev/)
[Selenium](https://pkg.go.dev/github.com/tebeka/selenium)

#### Selenium Driver Dependencies (IMPORTANT)
[Chrome Driver Docs](https://developer.chrome.com/docs/chromedriver)
[Chrome Browser](https://www.google.com/chrome/)
[XFVB virtual frame buffer](https://www.x.org/releases/X11R7.6/doc/man/man1/Xvfb.1.xhtml)

#### Search Engine Service
[Go](https://go.dev/)
