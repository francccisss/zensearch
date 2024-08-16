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
    if (node.children[0] === null) return null;
    return this.search_key(node.children[i], key);
  }

  search_for_insertion(node: BTreeNode, new_key: number): BTreeNode {
    let i = 0;
    while (i < node.keys.length && new_key > node.keys[i]) {
      i++;
    }
    if (node.children[0] == null) {
      return node;
    }
    // Check if the child at index `i` is null before proceeding
    if (node.children[i] === null) {
      throw new Error("Child node cannot be null");
    }

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
    const space = searched_node.keys.length !== this.order - 1 ? true : false;
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
    const new_keys = [...node_to_split.keys, new_key].sort((a, b) => a - b);
    const median_index = Math.floor(new_keys.length / 2);
    const median = new_keys[median_index];

    const left_node = new BTreeNode(this.order);
    const right_node = new BTreeNode(this.order);

    left_node.keys = new_keys.slice(0, median_index);
    right_node.keys = new_keys.slice(median_index + 1);

    if (node_to_split.children[0] !== null) {
      const median_child_index = Math.floor(node_to_split.children.length / 2);
      left_node.children = node_to_split.children.slice(
        0,
        median_child_index + 1,
      );
      right_node.children = node_to_split.children.slice(
        median_child_index + 1,
      );
    }
    if (node_to_split === this.root) {
      const new_root = new BTreeNode(this.order);
      new_root.keys = [median];
      new_root.children[0] = left_node;
      new_root.children[1] = right_node;
      this.root = new_root;
      return new_root;
    } else {
      const parent_node = this.find_parent(this.root, node_to_split);
      if (parent_node === null) return null;
      this.insert(parent_node, median);
      const parent_index = parent_node.keys.indexOf(median);
      parent_node.children[parent_index] = left_node;
      parent_node.children[parent_index + 1] = right_node;
      parent_node.children.splice(parent_index + 1, 0, right_node);
      //Check if the parent needs to be split and recurse upwards
      if (parent_node.keys.length === this.order) {
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

btree.insert_and_split(btree.root, 1);
btree.insert_and_split(btree.root, 4);
btree.insert_and_split(btree.root, 6);
btree.insert_and_split(btree.root, 10);
btree.insert_and_split(btree.root, 11);
btree.insert_and_split(btree.root, 12);
btree.insert_and_split(btree.root, 5);
btree.insert_and_split(btree.root, 7);
btree.insert_and_split(btree.root, 20);
btree.insert_and_split(btree.root, 15);
btree.insert_and_split(btree.root, 24);
btree.insert_and_split(btree.root, 18);
btree.insert_and_split(btree.root, 25);
btree.insert_and_split(btree.root, 26);
btree.insert_and_split(btree.root, 100);
btree.insert_and_split(btree.root, 30);
btree.insert_and_split(btree.root, 32);
btree.insert_and_split(btree.root, 13);
console.log(btree.root);
btree.root.children.forEach((child) => {
  console.log(child);
});
