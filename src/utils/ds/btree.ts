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

    //return null;
    if (node.children[0] === null) return null;
    return this.search_key(node.children[i], key);
  }

  search_for_insertion(node: BTreeNode, new_key: number): BTreeNode {
    let i = 0;

    while (i < node.keys.length && new_key > node.keys[i]) {
      i++;
    }
    if (node.children[0] == null) return node;
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

  insert_and_split(node: BTreeNode, new_key: number): BTreeNode | null {
    // if we found the bottom of the node
    const searched_node = this.search_for_insertion(node, new_key);
    // check if there is space to insert directly on the searched node
    const space = searched_node.keys.length !== this.order - 1 ? true : false;
    if (space) {
      // insert directly and sort;
      const inserted_node = this.insert(searched_node, new_key);
      return inserted_node;
    }
    // Do some mumbo jumbo here
    const new_keys = [...searched_node.keys, new_key].sort((a, b) => a - b);
    const median_index = Math.floor(new_keys.length / 2);
    const median = searched_node.keys[median_index];

    const left_node = new BTreeNode(this.order);
    const right_node = new BTreeNode(this.order);

    left_node.keys = new_keys.slice(0, median_index);
    right_node.keys = new_keys.slice(median_index + 1);

    console.log(searched_node);
    if (searched_node === this.root) {
      const new_root = new BTreeNode(this.order);
      new_root.keys = [median];
      new_root.children[0] = left_node;
      new_root.children[1] = right_node;
      this.root = new_root;
      return new_root;
    } else {
      const parent_node = this.find_parent(this.root, searched_node);
      console.log(parent_node);
      console.log(parent_node === searched_node);
      this.insert(parent_node as BTreeNode, median);
      parent_node!.children[parent_node!.children.indexOf(searched_node) + 1] =
        right_node;
      //return this.insert_and_split(parent_node!, median);
    }
    console.log("end");
    return null;
  }

  find_parent(root: BTreeNode, target_node: BTreeNode): BTreeNode | null {
    if (root === target_node) {
      return null; // Root node doesn't have a parent
    }

    const queue: BTreeNode[] = [root]; // Initialize BFS queue with the root node

    while (queue.length > 0) {
      const current_node = queue.shift()!;

      // Check if current_node has target_node as a child
      if (current_node.children.includes(target_node)) {
        return current_node;
      }

      // Enqueue all non-null children of current_node
      for (const child of current_node.children) {
        if (child !== null) {
          queue.push(child);
        }
      }
    }

    return null; // target_node was not found
  }
}

const btree = new BTree(4);

btree.insert_and_split(btree.root, 1);
btree.insert_and_split(btree.root, 4);
btree.insert_and_split(btree.root, 6);
btree.insert_and_split(btree.root, 10);
console.log(btree.root.keys);
console.log(btree.root.children);
