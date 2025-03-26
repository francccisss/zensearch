// Message
export type IndexedWebpage = {
  Message: string;
  CrawlStatus: number;
  Webpage: {
    Header: header;
    Contents: string;
  };
  Title: string;
  URLSeed: string;
};
export const CRAWL_SUCCESS = 0;
export const CRAWL_FAIL = 1;
export type CrawlStatus = number;

type header = {
  Title: string;
  Url: string;
};
export type Webpage = {
  Contents: string;
  Title: string;
  Url: string;
};

export type Segment = {
  Header: {
    SequenceNum: number;
    SegmentLength: number;
    TotalLength: number;
  };
  Data: Buffer;
};

export type URLs = {
  Domain: string;
  Nodes: Array<string>;
};

export type Node = {
  ID: string;
  Url: string;
  QueueID: string;
};

export type FrontierQueue = {
  QueueID: string;
  Domain: string;
};
