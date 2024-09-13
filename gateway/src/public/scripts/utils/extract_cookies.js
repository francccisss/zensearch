export default function () {
  let cookies = {};
  document.cookie.split("; ").forEach((c) => {
    const [key, value] = c.split("=");
    cookies[`${key}`] = value;
  });
  console.log(cookies);
  return cookies;
}
