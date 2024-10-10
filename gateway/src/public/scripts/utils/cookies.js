function extractCookies() {
  let cookies = {};
  document.cookie.split("; ").forEach((cookie) => {
    const [key, value] = cookie.split("=");
    cookies[key] = value;
  });
  return cookies;
}

function clearAllCookies() {
  const cookies = document.cookie.split("; ");
  for (let cookie of cookies) {
    console.log(cookie);
    const name = cookie.split("=")[0];
    document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/";
  }
  console.log(document.cookie);
}

export default { extractCookies, clearAllCookies };
