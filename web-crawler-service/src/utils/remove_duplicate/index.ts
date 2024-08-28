export default function <T>(links: Array<T>): Array<T> {
  let tmp: Array<T> = [];
  outerLoop: for (let i = 0; i < links.length; i++) {
    if (tmp.length === 0) {
      tmp.push(links[i]);
      continue;
    }
    let exists = false;
    for (let j = 0; j < tmp.length; j++) {
      if (links[i] === tmp[j]) {
        exists = true;
        break;
      }
    }
    if (!exists) tmp.push(links[i]);
  }
  return tmp;
}
