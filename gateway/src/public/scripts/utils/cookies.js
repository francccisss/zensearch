function extractCookies() {
  let cookies = {};
  document.cookie = "message_type=; Max-Age=0; path=/";
  console.log(document.cookie);
  document.cookie.split("; ").forEach((cookie) => {
    const [key, value] = cookie.split("=");
    cookies[key] = value;
  });
  return cookies;
}

function clearAllCookies() {
  const cookies = document.cookies.split("; ");
  for (let cookie of cookies) {
    console.log(cookie);
  }
}

export default { extractCookies, clearAllCookies };
