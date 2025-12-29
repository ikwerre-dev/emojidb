"use client";

import { Code, Copy, Github } from "lucide-react";

interface HeroSectionProps {
    onCopy: () => void;
}

export default function HeroSection({ onCopy }: HeroSectionProps) {
    return (
        <main className="w-full">
            <h1 className="font-sekuya py-2 md:py-0 lg:border-x-[2.3px] lg:border-dashed border-white/90 font-[900] text-center text-[70px] sm:text-[120px] md:text-[160px] lg:text-[260px]">
                <span className="flex flex-col leading-[1.3]">
                    <span className="text-white/80">EmojiDB</span>
                </span>
            </h1>
            <div className="grid grid-cols-1 md:grid-cols-6 border-2 border-x-white/90">
                <a
                    href="https://github.com/ikwerre-dev/emojidb"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="border-b md:border-b-0 md:border-r-2 items-center col-span-1 md:col-span-2 gap-2 flex cursor-pointer bg-white/5 hover:bg-white/50 transition-all duration-300 p-3 md:p-5"
                >
                    <Github size={15} className="md:size-5" />
                    <p className="text-sm md:text-base">Star this shit project on Github</p>
                </a>

                <div className="flex gap-2 items-center md:gap-5 col-span-1 md:col-span-4  md:px-10 text-sm md:text-base font-sekuya  md:justify-end p-3 md:p-5">
                    <Code size={15} className="md:hidden" />

                    <p className="truncate">npm install @ikwerre-dev/emojidb</p>
                    <button onClick={onCopy} className="hover:opacity-70 transition-opacity">
                        <Copy size={15} className="md:size-[18px] flex-shrink-0" />
                    </button>
                </div>
            </div>
        </main>
    );
}
