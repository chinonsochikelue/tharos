// XSS Vulnerabilities
const content = "<h1>Hacked</h1>";
element.innerHTML = content; // Should trigger XSS warning
element.outerHTML = content; // Should trigger XSS warning
document.write(userInput);   // Should trigger XSS warning

// React XSS
const dangerousComponent = () => {
    return <div dangerouslySetInnerHTML={{ __html: userInput }} />; // Should trigger XSS warning
};

// Weak Cryptography
const randomToken = Math.random(); // Should trigger Weak Crypto warning
const hash = md5(password);        // Should trigger Weak Crypto warning
const badHash = sha1(data);        // Should trigger Weak Crypto warning

// SQL Injection
const query = "SELECT * FROM users WHERE id = " + userInput; // Should trigger SQLi warning
const update = "UPDATE users SET name = '" + name + "'";     // Should trigger SQLi warning

// Hardcoded Secrets (Existing check)
const apiKey = "sk-1234567890abcdef1234567890abcdef";

// TODOs (Existing check)
// TODO: Fix this security hole
