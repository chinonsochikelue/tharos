const express = require("express");
const app = express();

// Test CORS detection
app.use((req, res, next) => {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("X-Content-Type-Options", "none");
    next();
});

app.listen(3000);
