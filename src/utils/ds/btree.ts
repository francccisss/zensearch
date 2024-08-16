import clean_links from "../cleanup_links/clean_links";

const sorted = [1, 2, 3, 4];
class BTreeNode {
  keys: number[];
  children: BTreeNode[];
  num_keys: number;
  constructor(order: number) {
    this.keys = [];
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
    if (node.children.length === 0 || node.children[i] === null) return null;
    return this.search_key(node.children[i], key);
  }

  search_for_insertion(node: BTreeNode, new_key: number): BTreeNode | null {
    let i = 0;
    console.log("Keys: ", node.keys);
    while (i < node.keys.length && new_key > node.keys[i]) {
      i++;
    }
    if (node.keys.includes(new_key)) {
      console.error("Duplicate key detected:", new_key);
      //node.keys = [...clean_links<number>(node.keys)];
      console.log(node.keys);
      return null; // or handle duplicates according to your requirements
    }
    if (node.children.length === 0 || node.children[i] === null) return node;

    // Check if the child at index `i` is null before proceeding
    return this.search_for_insertion(node.children[i], new_key);
  }

  private insert(node: BTreeNode, new_key: number) {
    let i = node.keys.length - 1;
    while (i >= 0 && node.keys[i] > new_key) {
      node.keys[i + 1] = node.keys[i];
      i--;
    }

    node.keys[i + 1] = new_key;
    return node;
  }

  insert_and_split(node: BTreeNode, new_key: number): BTreeNode | null {
    const searched_node = this.search_for_insertion(node, new_key);
    if (searched_node === null) return null;
    const space = searched_node.keys.length < this.order;
    if (space) {
      const inserted_node = this.insert(searched_node, new_key);
      return inserted_node;
    }
    const new_split = this.split(searched_node, new_key);
    if (new_split == null) return null;
    return new_split;
  }

  private split(node_to_split: BTreeNode, new_key: number): BTreeNode | null {
    // Do some mumbo jumbo here
    console.log("M = 4 - 1 = 3 keys only per node.");
    const new_keys = [...node_to_split.keys, new_key].sort((a, b) => a - b);
    const median_index = Math.floor(new_keys.length / 2);
    const median = new_keys[median_index];

    const left_node = new BTreeNode(this.order);
    const right_node = new BTreeNode(this.order);

    left_node.keys = new_keys.slice(0, median_index);
    right_node.keys = clean_links<number>(new_keys.slice(median_index + 1)); // something wrong with this
    // had to remove duplicate on right node
    console.log("Right node: ", right_node.keys);

    if (node_to_split.children[0] !== null) {
      const median_child_index = Math.floor(node_to_split.children.length / 2);
      left_node.children = node_to_split.children
        .slice(0, median_child_index + 1)
        .filter((child) => child !== null);
      right_node.children = node_to_split.children
        .slice(median_child_index + 1)
        .filter((child) => child !== null);
    }
    console.log({ median, new_key, added_new_key: node_to_split.keys });
    console.log({ left_node, right_node });
    if (node_to_split === this.root) {
      console.log("Current Node is Root", this.root.keys);
      const new_root = new BTreeNode(this.order);
      new_root.keys = [median];
      new_root.children[0] = left_node;
      new_root.children[1] = right_node;
      console.log("New Root: ", new_root.keys);
      this.root = new_root;
      return new_root;
    } else {
      const parent_node = this.find_parent(this.root, node_to_split);
      console.log("Current Node is not Root", this.root.keys);
      if (parent_node === null) return null;
      console.log("PARENT KEYS: ", parent_node?.keys);
      this.insert(parent_node, median);
      console.log("AFTER INSERT MEDIAN TO PARENT: ", parent_node.keys);
      const parent_index = parent_node.keys.indexOf(median);
      parent_node.children[parent_index] = left_node;
      parent_node.children[parent_index + 1] = right_node;
      if (parent_node.keys.length === this.order) {
        console.log("Split again: ", parent_node.keys);
        return this.split(parent_node, median);
      }
      return parent_node;
    }
    console.log("end");
  }

  find_parent(root: BTreeNode, target_node: BTreeNode): BTreeNode | null {
    if (root === target_node) {
      return null;
    }

    const queue: BTreeNode[] = [root];

    while (queue.length > 0) {
      const current_node = queue.shift()!;

      if (current_node.children.includes(target_node)) {
        return current_node;
      }

      for (const child of current_node.children) {
        if (child !== null) {
          queue.push(child);
        }
      }
    }

    return null;
  }
}

const btree = new BTree(5);
export default btree;
