// This is a test file for Tharos
const API_KEY = "sk-1234567890abcdef1234567890abcdefguyyh"; // Should be blocked

function unsafe() {
    eval("console.log('dangerous')"); // Should be blocked
}

// TODO: Refactor this logic later
function smells() {
    return true;
}
