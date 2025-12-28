import { spawn } from 'child_process';
import path from 'path';
import readline from 'readline';
import fs from 'fs';
import https from 'https';
import os from 'os';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

class BinaryManager {
    constructor() {
        this.platform = os.platform();
        this.arch = os.arch();
        this.binDir = path.join(__dirname, 'bin');
        this.engineName = `emojidb-${this.platform}-${this.arch}${this.platform === 'win32' ? '.exe' : ''}`;
        this.enginePath = path.join(this.binDir, this.engineName);
    }

    async download(url, dest) {
        return new Promise((resolve, reject) => {
            const file = fs.createWriteStream(dest);
            const request = (requestUrl) => {
                https.get(requestUrl, (response) => {
                    // Handle All Redirects (301, 302, 307, 308)
                    if ([301, 302, 307, 308].includes(response.statusCode) && response.headers.location) {
                        request(response.headers.location);
                        return;
                    }

                    if (response.statusCode !== 200) {
                        fs.unlink(dest, () => { });
                        reject(new Error(`Server returned HTTP ${response.statusCode}. Please ensure a Release exists with the correctly named binary: ${this.engineName}`));
                        return;
                    }

                    response.pipe(file);
                    file.on('finish', () => {
                        file.close();
                        if (this.platform !== 'win32') {
                            fs.chmodSync(dest, 0o755);
                        }
                        resolve(dest);
                    });
                }).on('error', (err) => {
                    fs.unlink(dest, () => { });
                    reject(err);
                });
            };
            request(url);
        });
    }

    async ensureBinary() {
        if (fs.existsSync(this.enginePath)) return this.enginePath;

        console.log(`🚀 EmojiDB: Engine not found for ${this.platform}-${this.arch}. Attempting download...`);

        if (!fs.existsSync(this.binDir)) {
            fs.mkdirSync(this.binDir, { recursive: true });
        }

        const url = `https://github.com/ikwerre-dev/EmojiDB/releases/latest/download/${this.engineName}`;
        try {
            await this.download(url, this.enginePath);
            console.log('✅ EmojiDB: Engine ready.');
            return this.enginePath;
        } catch (err) {
            console.error(`\n❌ EmojiDB Setup Error: ${err.message}`);
            console.error(`💡 FIX: You MUST compile and upload '${this.engineName}' to your GitHub Release for this to work standalone.`);
            throw err;
        }
    }
}

class EmojiDB {
    constructor(options = {}) {
        this.manager = new BinaryManager();
        this.enginePath = options.enginePath || null;
        this.process = null;
        this.rl = null;
        this.pending = new Map();
    }

    async connect() {
        if (!this.enginePath) {
            this.enginePath = await this.manager.ensureBinary();
        }

        return new Promise((resolve, reject) => {
            this.process = spawn(this.enginePath);

            this.rl = readline.createInterface({
                input: this.process.stdout,
                terminal: false
            });

            this.rl.on('line', (line) => {
                try {
                    const res = JSON.parse(line);
                    const p = this.pending.get(res.id);
                    if (p) {
                        if (res.error) p.reject(new Error(res.error));
                        else p.resolve(res.data);
                        this.pending.delete(res.id);
                    }
                } catch (e) {
                    console.error('Failed to parse engine response:', e);
                }
            });

            this.process.stderr.on('data', (data) => {
                console.error(`Engine Error: ${data}`);
            });

            this.process.on('error', (err) => {
                reject(new Error(`Failed to start engine: ${err.message}`));
            });

            setTimeout(() => {
                resolve({ status: 'connected', pid: this.process.pid });
            }, 100);
        });
    }

    get status() {
        if (this.process && !this.process.killed) {
            return { status: 'connected', pid: this.process.pid };
        }
        return { status: 'disconnected' };
    }

    async send(method, params = {}) {
        const id = Math.random().toString(36).substring(7);
        return new Promise((resolve, reject) => {
            this.pending.set(id, { resolve, reject });
            const payload = JSON.stringify({ id, method, params });
            if (!this.process || this.process.killed) {
                return reject(new Error("Database not connected. Call db.connect() first."));
            }
            this.process.stdin.write(payload + '\n');
        });
    }

    async open(dbPath, key) {
        return this.send('open', { path: dbPath, key });
    }

    async defineSchema(table, fields) {
        return this.send('define_schema', { table, fields });
    }

    async insert(table, row) {
        return this.send('insert', { table, row });
    }

    async query(table, match = {}) {
        return this.send('query', { table, match });
    }

    async update(table, match, updateData) {
        return this.send('update', { table, match, update: updateData });
    }

    async delete(table, match) {
        return this.send('delete', { table, match });
    }

    async secure() {
        return this.send('secure');
    }

    async rekey(newKey, masterKey) {
        return this.send('rekey', { new_key: newKey, master_key: masterKey });
    }

    async close() {
        if (this.process && !this.process.killed) {
            await this.send('close');
            this.process.kill();
        }
    }
}

export default EmojiDB;
