const fs = require('fs');
const path = require('path');

const base = 'D:\\go_work\\simple-admin\\travel\\travel-rpc\\desc';

// The merged proto with ALL services is at desc/travel/travel.proto
const mergedPath = path.join(base, 'travel', 'travel.proto');
const content = fs.readFileSync(mergedPath, 'utf8');

// Write to desc/travel.proto
const outPath = path.join(base, 'travel.proto');
fs.writeFileSync(outPath, content, 'utf8');
console.log('Written desc/travel.proto:', fs.statSync(outPath).size, 'bytes');

// Verify
const verify = fs.readFileSync(outPath, 'utf8');
const serviceCount = (verify.match(/service\s+\w+/g) || []).length;
const messageCount = (verify.match(/message\s+\w+/g) || []).length;
console.log(`Services: ${serviceCount}, Messages: ${messageCount}`);
console.log('Has "proto3":', verify.includes('"proto3"'));

// Remove subdirectories
const subdirs = ['travel', 'catalog', 'inventory', 'order', 'payment', 'management', 'user_auth'];
for (const dir of subdirs) {
    const dirPath = path.join(base, dir);
    if (fs.existsSync(dirPath) && fs.statSync(dirPath).isDirectory()) {
        fs.rmSync(dirPath, { recursive: true, force: true });
        console.log('Removed dir:', dir);
    }
}

// Remove temp files from project root
const rootDir = path.join(base, '..');
for (const f of ['_fix_proto.js', '_fix_proto.py', '_test_gen.js', '_merge_proto.js']) {
    const fp = path.join(rootDir, f);
    if (fs.existsSync(fp)) {
        fs.unlinkSync(fp);
        console.log('Removed temp:', f);
    }
}

// Also remove desc/travel.proto.tmp if exists
const tmpFile = path.join(base, 'travel.proto.tmp');
if (fs.existsSync(tmpFile)) {
    fs.unlinkSync(tmpFile);
    console.log('Removed temp: travel.proto.tmp');
}

console.log('\nFinal desc/ contents:');
for (const f of fs.readdirSync(base)) {
    const stat = fs.statSync(path.join(base, f));
    console.log(`  ${f} (${stat.size} bytes)`);
}
