import DecorativeBorder from "@/components/DecorativeBorder";

export default function Home() {
  return (
    <div className="flex min-h-screen lg:px-5 flex-col  bg-[#0a0a0a] font-sans">
      <DecorativeBorder side="left" />
      <DecorativeBorder side="right" />

      <div className="px-4 sm:px-8 lg:px-16 min-h-screen flex lg:py-[1rem] flex-col w-full">
        <main className="w-full flex-1 lg:border-x lg:border-x-3 border-dashed border-white/50">
          <h1 className="font-sekuya font-[900] border-b-3 border-b-white/50 text-center text-[70px] sm:text-[120px] md:text-[160px] lg:text-[260px]">
            <span className="flex flex-col  leading-[1.3]"  >
              <span className="text-white/80">EmojiDB</span>
            </span>
          </h1>
        </main>
      </div>
    </div>
  );
}
