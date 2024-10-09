export default function () {
  let cookies = {};
  document.cookie.split("; ").forEach((cookie) => {
    const [key, value] = cookie.split("=");
    cookies[key] = value;
  });
  return cookies;
}
