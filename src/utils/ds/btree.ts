const sorted = [1, 2, 3, 4];
class BTreeNode {
  keys: number[];
  children: BTreeNode[];
  num_keys: number;
  constructor(order: number) {
    this.keys = [];
    //this.keys = new Array(order).fill(null);
    this.children = new Array(order + 1).fill(null);
    this.num_keys = 0;
  }
}

class BTree {
  root: BTreeNode;
  order: number;

  constructor(order: number) {
    this.root = new BTreeNode(order);
    this.order = order;
  }
  search_key(node: BTreeNode, key: number): BTreeNode | null {
    let i = 0;
    while (i < node.keys.length && key > node.keys[i]) {
      i++;
    }
    if (i < node.keys.length && key === node.keys[i]) {
      return node;
    }

    //return null;
    if (node.children.length === 0) {
      // problem because children is not empty but just filled with null values
      return null;
    }
    return this.search_key(node.children[i], key);
  }

  search_for_insertion(node: BTreeNode, new_key: number): BTreeNode {
    let i = 0;
    while (i < node.keys.length && new_key > node.keys[i]) {
      i++;
    }
    if (node.children.length === 0) {
      node.num_keys++; // increment numkeys when searching to keep track of how many is being added to the node keys
      return node;
    }
    // if there are no more keys to compare to and is greater than all of the keys, i == to the maximum key
    // if new_key is < all of the keys, then i = to the range where new_key fits
    return this.search_for_insertion(node.children[i], new_key);
  }

  private insert(node: BTreeNode, new_key: number) {
    // INSERTION SORT
    let i = node.keys.length - 1;
    while (i >= 0 && node.keys[i] > new_key) {
      node.keys[i + 1] = node.keys[i];
      i--;
    }
    node.keys[i + 1] = new_key;
    return node;
  }

  insert_and_split(node: BTreeNode, new_key: number) {
    // if we found the bottom of the node
    const searched_node = this.search_for_insertion(node, new_key);
    // check if there is space to insert directly on the searched node
    const space = searched_node.num_keys < this.order ? true : false;
    if (space) {
      // insert directly and sort;
      const inserted_node = this.insert(searched_node, new_key);
      return inserted_node;
    }
    // Do some mumbo jumbo here
    const new_keys = [...searched_node.keys, new_key].sort((a, b) => a - b);
    const median_index = Math.floor(new_keys.keys.length / 2);
    const median = searched_node.keys[median_index];

    const left_node = new BTreeNode(this.order);
    const right_node = new BTreeNode(this.order);

    for (let i = 0; i < median_index; i++) {
      left_node.keys.push(searched_node.keys[i]);
    }
    for (let i = median_index + 1; i < new_keys.length; i++) {
      right_node.keys.push(searched_node.keys[i]);
    }
    left_node.num_keys = left_node.keys.length;
    right_node.num_keys = right_node.keys.length;

    if (searched_node === this.root) {
      const new_root = new BTreeNode(this.order);
      new_root.keys = [median];
      new_root.num_keys = new_root.keys.length;
      new_root.children[0] = left_node;
      new_root.children[1] = right_node;
      return new_root;
    }
  }
}

const btree = new BTree(3);

console.log(btree.search_key(btree.root, 2));
