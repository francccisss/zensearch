export type data_t = {
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
export type webpage_t = {
  Contents: string;
  Title: string;
  Url: string;
};
