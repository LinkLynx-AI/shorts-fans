export default function RootPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-12 sm:px-10">
      <section className="w-full rounded-[2rem] border border-white/80 bg-white/76 p-8 shadow-[0_24px_80px_rgba(87,38,8,0.14)] backdrop-blur-lg sm:p-10">
        <p className="text-xs font-semibold uppercase tracking-[0.26em] text-accent">frontend foundation</p>
        <h1 className="mt-5 max-w-3xl text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
          ページ構造の仮置きは外し、UI 基盤だけを残しています。
        </h1>
        <p className="mt-4 max-w-3xl text-sm leading-7 text-muted sm:text-base">
          `home`、`subscriptions`、`profile`、`creator` などの route 前提は一旦削除しました。
          この状態から、必要な画面が固まってから App Router 配下を組み直せます。
        </p>
      </section>
    </main>
  );
}
