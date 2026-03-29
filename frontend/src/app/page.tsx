import styles from "./page.module.css";

export default function HomePage() {
  return (
    <main className={styles.page}>
      <section className={styles.panel} aria-labelledby="app-home-title">
        <p className={styles.eyebrow}>frontend foundation</p>
        <h1 id="app-home-title" className={styles.title}>
          Next.js App Router is ready.
        </h1>
        <p className={styles.description}>
          `frontend/` は Next.js App Router を前提に初期化済みです。
          ここから route、layout、feature 単位の実装を積み上げられます。
        </p>
        <ul className={styles.list}>
          <li>TypeScript + ESLint を有効化</li>
          <li>`src/app` を root にした App Router 構成</li>
          <li>`@/*` import alias を設定</li>
          <li>`pnpm lint` と `pnpm typecheck` を実行可能に調整</li>
        </ul>
      </section>
    </main>
  );
}
