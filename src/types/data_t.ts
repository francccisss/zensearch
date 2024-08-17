
export type data_t = {
  webpage_contents: Array<{
    header: { title: string; page_url: string };
    contents: string;
  }>;
  header: {
    title: string;
  };
};