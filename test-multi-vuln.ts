// XSS
document.getElementById('app').innerHTML = userInput;
// Weak Crypto
const pass = md5("password");
const rand = Math.random();
// Injection
const query = "SELECT * FROM users WHERE id = " + id;
// Quality
// TODO: Fix this
