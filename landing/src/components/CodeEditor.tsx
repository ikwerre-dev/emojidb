"use client";

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Copy } from 'lucide-react';

const codeExample = `import { EmojiDB } from '@ikwerre-dev/emojidb';

const db = new EmojiDB('mydb', 'secret-key');

await db.createTable('users', [
  { name: 'id', type: 'string', unique: true },
  { name: 'email', type: 'string', unique: true },
  { name: 'name', type: 'string' }
]);

await db.insert('users', {
  id: '1',
  email: 'user@example.com',
  name: 'John Doe'
});

const users = await db.query('users', {
  email: 'user@example.com'
});`;

export default function CodeEditor() {
    return (
            <div className="h-full  relative overflow-hidden">
                <button className="absolute top-4 right-4 p-2 border border-white/20 hover:bg-white/10 transition-colors rounded z-10">
                    <Copy size={16} className="text-white" />
                </button>

                <div className="h-full overflow-auto p-6 bg-black">
                    <SyntaxHighlighter
                        language="javascript"
                        style={vscDarkPlus}
                        customStyle={{
                            background: 'transparent',
                            padding: 0,
                            margin: 0,
                            fontSize: '14px',
                            lineHeight: '1.6',
                        }}
                        showLineNumbers={false}
                    >
                        {codeExample}
                    </SyntaxHighlighter>
                </div>
            </div>
     );
}
