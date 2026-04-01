import { expect, test } from "@playwright/test";

test("root page renders without predefined route structure", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("heading", { name: "ページ構造の仮置きは外し、UI 基盤だけを残しています。" })).toBeVisible();
});

test("undefined routes fall back to the shared not-found page", async ({ page }) => {
  await page.goto("/missing-route");
  await expect(page.getByRole("heading", { name: "指定された route はまだ用意されていません。" })).toBeVisible();
  await page.getByRole("link", { name: "トップへ戻る" }).click();
  await expect(page).toHaveURL(/\/$/);
});
