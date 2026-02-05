// Safe Secrets (Placeholders or Env Vars)
const apiKey = process.env.API_KEY;
const dbPass = "unset"; // Too short to trigger entropy/length rule

const config = {
    token: getSecureToken(),
    mode: "development"
};

// Not a secret, just a long string
const description = "This is a very long description that might have some entropy but it is clearly just text and not a hex string or a key.";
