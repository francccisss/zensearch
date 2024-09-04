export type data_t = {
  webpages: Array<{
    header: { title: string; webpage_url: string };
    contents: string;
  }>;
  header: {
    title: string;
    url: string;
  };
};

export type webpage_t = {
  contents: string;
  title: string;
  webpage_url: string;
};
