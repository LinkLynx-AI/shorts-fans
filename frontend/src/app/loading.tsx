export default function Loading() {
  return (
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] flex-col overflow-hidden bg-white">
      <div className="flex-1 bg-[linear-gradient(180deg,#94e0ff_0%,#2a648f_56%,#07131d_100%)] px-4 pb-[76px] pt-6">
        <div className="mx-auto h-8 w-44 animate-pulse rounded-full bg-white/22" />
        <div className="mx-auto mt-10 h-[420px] w-full max-w-[274px] animate-pulse rounded-[34px] border border-white/24 bg-white/18" />
        <div className="mt-8 h-12 animate-pulse rounded-full bg-white/82" />
        <div className="mt-5 space-y-3">
          <div className="flex items-center gap-3">
            <div className="size-12 animate-pulse rounded-[20px] bg-white/28" />
            <div className="space-y-2">
              <div className="h-4 w-28 animate-pulse rounded-full bg-white/22" />
              <div className="h-3 w-20 animate-pulse rounded-full bg-white/18" />
            </div>
          </div>
          <div className="h-4 w-60 animate-pulse rounded-full bg-white/18" />
          <div className="h-4 w-48 animate-pulse rounded-full bg-white/18" />
        </div>
      </div>
    </main>
  );
}
