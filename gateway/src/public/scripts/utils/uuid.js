export default function () {
  var ret = "",
    value;
  for (var i = 0; i < 32; i++) {
    value = (random() * 16) | 0;
    // Insert the hypens
    if (i > 4 && i < 21 && !(i % 4)) {
      ret += "-";
    }
    // Add the next random character
    ret += (i === 12 ? 4 : i === 16 ? (value & 3) | 8 : value).toString(16);
  }
  return ret;
}

export function random() {
  return Math.random();
}
