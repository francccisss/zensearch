export type wsheader_t = {
  type: "search" | "crawl";
};

export type CrawlMessageStatus = {
  IsSuccess: boolean;
  Message: string;
  URLSeed: string;
};
