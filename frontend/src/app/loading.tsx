export default function Loading() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-8 px-6 py-8 sm:px-10">
      <div className="h-12 w-52 animate-pulse rounded-full bg-white/50" />
      <div className="grid gap-5 lg:grid-cols-[1.4fr_0.9fr]">
        <section className="rounded-[2rem] border border-white/70 bg-white/60 p-8 shadow-[0_24px_80px_rgba(87,38,8,0.12)] backdrop-blur">
          <div className="h-4 w-24 animate-pulse rounded-full bg-accent/20" />
          <div className="mt-6 h-10 w-2/3 animate-pulse rounded-full bg-stone-300/50" />
          <div className="mt-4 h-4 w-full animate-pulse rounded-full bg-stone-300/50" />
          <div className="mt-3 h-4 w-5/6 animate-pulse rounded-full bg-stone-300/50" />
          <div className="mt-8 grid gap-3 sm:grid-cols-3">
            {Array.from({ length: 3 }).map((_, index) => (
              <div
                key={index}
                className="h-24 animate-pulse rounded-[1.5rem] bg-[linear-gradient(135deg,rgba(255,255,255,0.95),rgba(255,237,213,0.75))]"
              />
            ))}
          </div>
        </section>
        <aside className="rounded-[2rem] border border-white/70 bg-[#1f1713] p-6 shadow-[0_24px_80px_rgba(87,38,8,0.16)]">
          <div className="h-4 w-28 animate-pulse rounded-full bg-white/15" />
          <div className="mt-5 grid gap-3">
            {Array.from({ length: 3 }).map((_, index) => (
              <div key={index} className="h-16 animate-pulse rounded-[1.25rem] bg-white/10" />
            ))}
          </div>
        </aside>
      </div>
    </main>
  );
}
