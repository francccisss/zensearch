import clean_links from "../cleanup_links/clean_links";

const sorted = [1, 2, 3, 4];
class BTreeNode {
  keys: number[];
  children: BTreeNode[];
  num_keys: number;
  constructor(order: number) {
    this.keys = [];
    this.children = new Array(order).fill(null);
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
    console.log("Key length: ", node.keys.length);
    while (i < node.keys.length && new_key > node.keys[i]) {
      i++;
    }
    if (node.keys.includes(new_key)) {
      console.error("Duplicate key detected:", new_key);
      console.log(node.keys);
      return null; // or handle duplicates according to your requirements
    }
    if (node.children.length === 0 || node.children[i] === null) return node;

    return this.search_for_insertion(node.children[i], new_key);
  }

  private insert(node: BTreeNode, new_key: number) {
    node.keys = [...node.keys, new_key].sort((a, b) => a - b);
    return node;
  }

  insert_and_split(node: BTreeNode, new_key: number): BTreeNode | null {
    const searched_node = this.search_for_insertion(node, new_key);
    if (searched_node === null) return null;
    const space = searched_node.keys.length < this.order - 1;
    // without -1 it will only stop until it is equal to the order
    if (space) {
      const inserted_node = this.insert(searched_node, new_key);
      inserted_node.num_keys++;
      return inserted_node;
    }
    console.log("SPLIT DIPOTA");
    const new_split = this.split(searched_node, new_key);
    if (new_split == null) return null;
    return new_split;
  }

  private split(node_to_split: BTreeNode, new_key: number): BTreeNode | null {
    // Do some mumbo jumbo here
    console.log(
      "M = %d - 1 = %d keys only per node.",
      this.order,
      this.order - 1,
    );
    let new_keys;
    new_keys = [...node_to_split.keys, new_key].sort((a, b) => a - b);
    if (node_to_split.keys.includes(new_key)) {
      console.log("Duplicate");

      new_keys = node_to_split.keys;
    }
    console.log({ new_keys });
    const median_index = Math.floor(new_keys.length / 2);
    const median = new_keys[median_index];

    const left_node = new BTreeNode(this.order);
    const right_node = new BTreeNode(this.order);

    left_node.keys = new_keys.slice(0, median_index);
    right_node.keys = clean_links<number>(new_keys.slice(median_index + 1)); // something wrong with this
    // had to remove duplicate on right node

    if (node_to_split.children[0] !== null) {
      const median_child_index = Math.floor(node_to_split.children.length / 2);
      left_node.children = node_to_split.children
        .slice(0, median_child_index + 1)
        .filter((child) => child !== null);
      right_node.children = node_to_split.children
        .slice(median_child_index + 1)
        .filter((child) => child !== null);
    }
    console.log({ median, new_key, all_keys: node_to_split.keys });
    console.log({ left_node, right_node });
    if (node_to_split === this.root) {
      console.log("Current Node is Root", this.root.keys);
      const new_root = new BTreeNode(this.order);
      new_root.keys = [median];
      new_root.children[0] = left_node;
      new_root.children[1] = right_node;
      console.log("New Root: ", new_root.keys);
      console.log({ left_node, right_node });
      this.root = new_root;
      return new_root;
    } else {
      const parent_node = this.find_parent(this.root, node_to_split);
      if (parent_node === null) return null;
      console.log("Current Node is not Root", node_to_split.keys);
      console.log("PARENT KEYS: ", parent_node?.keys);
      this.insert(parent_node, median);
      console.log("AFTER INSERT MEDIAN TO PARENT: ", parent_node.keys);
      const parent_index = parent_node.keys.indexOf(median);
      parent_node.children[parent_index] = left_node;
      parent_node.children[parent_index + 1] = right_node;
      if (parent_node.keys.length === this.order) {
        console.log("Split again: ", parent_node.keys);
        //const median_index = Math.floor(parent_node.keys.length / 2);
        //const median = parent_node.keys[median_index];
        console.log(median);
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

const btree = new BTree(4);
for (let i = 0; i < 10; i++) {
  btree.insert_and_split(btree.root, i);
}
console.log(btree.root.children);
export default btree;
