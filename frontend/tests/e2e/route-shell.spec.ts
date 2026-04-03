import { expect, test } from "@playwright/test";

test("fan shell routes render and navigation works", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("link", { name: "おすすめ" })).toHaveAttribute("aria-current", "page");
  await expect(page.getByRole("link", { name: /Unlock/i })).toBeVisible();
  await expect(page.getByText("Mina Rei")).toBeVisible();

  await page.getByRole("link", { name: "検索" }).click();
  await expect(page).toHaveURL(/\/search$/);
  await page.getByRole("searchbox").fill("mina");
  await expect(page.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute("href", "/creators/mina?from=search&q=mina");
  await page.getByRole("link", { name: /Mina Rei/i }).click();
  await expect(page).toHaveURL(/\/creators\/mina\?from=search&q=mina$/);
  await expect(page.getByRole("heading", { name: /Mina Rei creator profile/i })).toHaveCount(1);
  await expect(page.getByRole("button", { name: "Following" })).toBeVisible();
  await expect(page.getByText("Unlock")).toHaveCount(0);
  await page.getByRole("link", { name: /Back/i }).click();
  await expect(page).toHaveURL(/\/search\?q=mina$/);
  await expect(page.getByRole("searchbox")).toHaveValue("mina");

  await page.goto("/");
  await page.getByRole("link", { name: /Mina Rei/i }).click();
  await expect(page).toHaveURL(/\/creators\/mina\?from=feed&tab=recommended$/);
  await page.getByRole("link", { name: /Mina Rei quiet rooftop preview/i }).click();
  await expect(page).toHaveURL(/\/shorts\/rooftop\?creatorId=mina&from=creator&profileFrom=feed&profileTab=recommended$/);
  await page.getByRole("link", { name: /Back/i }).click();
  await expect(page).toHaveURL(/\/creators\/mina\?from=feed&tab=recommended$/);

  await page.getByRole("link", { name: "マイ" }).click();
  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("heading", { name: "Fan hub structure" })).toBeVisible();

  await page.goto("/shorts/rooftop");
  await expect(page.getByRole("link", { name: /Back/i })).toBeVisible();
  await expect(page.getByText("quiet rooftop preview.")).toBeVisible();

  await page.goto("/creators/sora");
  await expect(page.getByText("まだ公開中の short はありません。")).toBeVisible();

  await page.goto("/creators/mina?from=twitter&tab=other");
  await expect(page).toHaveURL(/\/creators\/mina\?from=twitter&tab=other$/);
  await expect(page.getByRole("heading", { name: /Mina Rei creator profile/i })).toHaveCount(1);
  await expect(page.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");

  await page.goto("/shorts/rooftop?from=share&profileFrom=other");
  await expect(page).toHaveURL(/\/shorts\/rooftop\?from=share&profileFrom=other$/);
  await expect(page.getByText("quiet rooftop preview.")).toBeVisible();
  await expect(page.getByRole("link", { name: /Back/i })).toHaveAttribute("href", "/");
});

test("undefined routes fall back to the shared not-found page", async ({ page }) => {
  await page.goto("/missing-route");
  await expect(page.getByRole("heading", { name: "指定された surface はまだ用意されていません。" })).toBeVisible();
  await page.getByRole("link", { name: "feed に戻る" }).click();
  await expect(page).toHaveURL(/\/$/);
});
