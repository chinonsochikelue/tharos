// security-ignore
const secret = "sk_live_this_is_a_test";

const password = "root123"; // Should be critical

const id = req.body.id; // identifier named password - should ignore or info if I had that rule
const query = `SELECT * FROM users WHERE id = ${id}`; // SQLi critical

// This route should be high
// app.get("/admin", (req, res) => { ... });
const route = "/debug";
