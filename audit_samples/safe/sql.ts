// Safe SQL (Parameterized)
const userId = "123";
const query = "SELECT * FROM users WHERE id = ?";

function getUser(id: string) {
    return db.execute("SELECT * FROM accounts WHERE acc_id = ?", [id]);
}

// Safe string concatenation (not a query)
const welcome = "Hello " + name;
