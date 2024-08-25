type node_t = {
  url: string;
  neighbors: Array<node_t>;
};
const rootNode: node_t = {
  url: "http://example.com/root",
  neighbors: [
    {
      url: "http://example.com/neighbor1",
      neighbors: [
        {
          url: "http://example.com/neighbor1-1",
          neighbors: [],
        },
        {
          url: "http://example.com/neighbor1-2",
          neighbors: [],
        },
      ],
    },
    {
      url: "http://example.com/neighbor2",
      neighbors: [
        {
          url: "http://example.com/neighbor2-1",
          neighbors: [],
        },
        {
          url: "http://example.com/neighbor2-2",
          neighbors: [],
        },
      ],
    },
    {
      url: "http://example.com/neighbor3",
      neighbors: [
        {
          url: "http://example.com/neighbor3-1",
          neighbors: [],
        },
        {
          url: "http://example.com/neighbor3-2",
          neighbors: [],
        },
      ],
    },
  ],
};
test("Traverse with Throttling", async () => {
  const visited_stack = new Set();
  const quantum = 50;
  let start: number = 0; // will be set by the time it was created
  let end: number = 0;
  //   let time_to_next_goto: number = 100; // will be set by the time it was created
  async function index_page() {
    return "Page Indexed";
  }
  async function goto(current_page: string, quantum: number) {
    `Visiting ${current_page}`;
    return new Promise((resolve) => {
      setTimeout(() => {
        console.log(`Timeout quantum ${quantum}`);
        resolve(quantum);
      }, quantum);
    });
  }

  async function traverse_pages(link: node_t) {
    visited_stack.add(link.url);
    console.log({ start, end }, "ms");
    if (start - end < quantum) {
      console.log({ difference: Math.abs(start - end) }, "ms");
      await goto(link.url, 100);
    }
    await goto(link.url, quantum);
    start = new Date().getMilliseconds();
    await index_page();
    const neighbors = link.neighbors;
    if (neighbors === undefined || neighbors.length === 0) {
      console.log("LOG: No neighbors.");
      return "Hello";
    }
    for (let current_neighbor of neighbors) {
      end = new Date().getMilliseconds();
      if (visited_stack.has(current_neighbor.url)) {
        console.log("Already Visited: " + link.url);
        continue;
      }
      await traverse_pages(current_neighbor);
    }
    return "End";
  }
  expect(await traverse_pages(rootNode)).toBe("End");
});
