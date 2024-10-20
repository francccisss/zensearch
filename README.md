# zensearch
Building a small scale Distributed Search Engine using NodeJs for Crawling and indexing and Golang for calculating and ranking each webpages' relevancy to the user's search query.

## Why did I build this
I've been studying about operating systems through Modern Operating Systems 4th edition by Andrew S. Tanenbaum, where I learned about threads and processes and Tsoding daily's video about TF-IDF, so I was curious and had to think of a project where I could utilize both concepts.

## What does it do
A distributed search engine where user's are able to control what they can search, they can manually crawl specific websites of their liking and based on what they want to work with everyday. User's can crawl the web with a click of a button, and while crawling they can continue using the search feature to query existing webpages in their database, crawling might take some time because of security reasons and network throttling mechanism used by different website authors.   

## Concepts
**TF-IDF**: "In information retrieval, tf–idf (also TF*IDF, TFIDF, TF–IDF, or Tf–idf), short for term frequency–inverse document frequency, is a measure of importance of a word to a document in a collection or corpus, adjusted for the fact that some words appear more frequently in general.Like the bag-of-words model, it models a document as a multiset of words, without word order. It is a refinement over the simple bag-of-words model, by allowing the weight of words to depend on the rest of the corpus." [source](https://en.wikipedia.org/wiki/Tf%E2%80%93idf).

- Despite it being fast to determine the ranking of each document relative to the user's search query, it disregards the context of the query as long as a document matches the terms in the search query eg: "dog bites man" or "man bites dog" does not matter in the context of this model.

**Document Length Normalization**: Is a technique which mitigates the length bias of a document whose context might not be relevant to the query of a user but because a document's length is greater than all other documents that contains the user's query, the *term frequency* of a longer document would be much more higher compared to all other documents if we disregard the concentration of a lengthy document to the term. 

- The Document length normalization mitigates the length of a long document by dividing: currentDocLength / avgDocLength and controlled by `b` which controls the normalization of the document to determine the concentration of the term in that document. if the term frequency is proportionate to the document length then that means the current document is relevant to the query, else if the document is longer than average and is not proportionate to the term frequency then it is most likely no the main focus of the document.

**Beyond TF-IDF using BM25**: (Needs documentation.)


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
