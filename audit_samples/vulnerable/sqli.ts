// Vulnerable SQL Injection
const userId = "123";
const query1 = "SELECT * FROM users WHERE id = " + userId;
const query2 = `SELECT * FROM users WHERE id = ${userId}`;

function getUser(id: string) {
    const sql = "SELECT * FROM accounts WHERE acc_id = " + id;
    return db.execute(sql);
}
