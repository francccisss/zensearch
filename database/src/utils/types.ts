// Message
export type IndexedWebpages = {
  Message: string;
  CrawlStatus: number;
  Webpages: Array<{
    Header: header;
    Contents: string;
  }>;
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
  Nodes: Array<Node>;
};

export type Node = {
  ID: string;
  Url: string;
  QueueID: string;
};

export type Queue = {
  QueueID: string;
  Domain: string;
};
