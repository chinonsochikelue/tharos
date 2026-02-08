const http = require('http');

const data = JSON.stringify({
    code: `
    const admin_key = "THAROS_MOCK_SECRET_8x9y2z_HIGH_ENTROPY_STRING";
    eval("console.log(admin_key)");
  `,
    ai: false,
    fix: false
});

const options = {
    hostname: 'localhost',
    port: 3000,
    path: '/api/analyze',
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Content-Length': data.length
    }
};

console.log('🧪 Testing Playground API at http://localhost:3000/api/analyze...');

const req = http.request(options, (res) => {
    let body = '';
    res.on('data', (chunk) => body += chunk);
    res.on('end', () => {
        console.log(`📡 Status: ${res.statusCode}`);
        try {
            const json = JSON.parse(body);
            console.log('✅ Response Received:');
            console.log(JSON.stringify(json, null, 2));

            if (json.findings && json.findings.length > 0) {
                console.log(`\n🎉 Success! Found ${json.findings.length} issues.`);
            } else if (json.error) {
                console.error(`\n❌ Error from API: ${json.error}`);
                if (json.details) console.error(`   Details: ${json.details}`);
                if (json.path) console.error(`   Path looked at: ${json.path}`);
            } else {
                console.log('\n⚠️  No findings, but no error either. Check code snippet.');
            }
        } catch (e) {
            console.error('\n❌ Failed to parse response body as JSON:');
            console.error(body);
        }
    });
});

req.on('error', (e) => {
    console.error(`\n❌ Request Error: ${e.message}`);
    console.log('   Is the dev server running on port 3000?');
});

req.write(data);
req.end();
