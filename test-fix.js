const express = require("express");
const app = express();

// Hardcoded credential - should be fixed
const dbPassword = "root123";

// SQL Injection vulnerability
app.get("/user", (req, res) => {
    const query = `SELECT * FROM users WHERE id = ${req.query.id}`;
    db.query(query);
});

app.listen(3000);
