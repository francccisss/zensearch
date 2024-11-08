// Message
export type Message = {
  Message: string;
  CrawlStatus: number;
  Webpages: Array<{
    Header: header;
    Contents: string;
  }>;
  Title: string;
  Url: string;
};

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
