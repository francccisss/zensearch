// Message
export const CRAWL_SUCCESS = 0;
export const CRAWL_FAIL = 1;
export type CrawlStatus = number;

type header = {
  Title: string;
  Url: string;
};

export type IndexedWebpage = {
  Message: string;
  CrawlStatus: number;
  Webpage: {
    Header: header;
    Contents: string;
  };
  URLSeed: string;
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

// Corresponds to SQL schema
export type Node = {
  id: number;
  url: string;
  queue_id: string;
  status: string;
};
export type FrontierQueue = {
  id: string;
  domain: string;
};
export type VisitedNode = {
  id: string;
  node_id: number;
  queue_id: number;
};

export type DequeuedUrl = {
  Url: string;
  RemainingInQueue: number;
};
