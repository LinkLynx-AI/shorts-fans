import { expect, test } from "@playwright/test";

test("consumer shell keeps shorts as the default entry and can move to home", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("heading", { name: "縦型 feed をこのアプリの主導線として固定する。" })).toBeVisible();

  await page.getByRole("link", { name: /home discovery hub/i }).click();

  await expect(page).toHaveURL(/\/home$/);
  await expect(page.getByRole("heading", { name: "好みの creator を掘るための入口を先に固定する。" })).toBeVisible();

  await page.getByRole("link", { name: "shorts を開く" }).click();

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("heading", { name: "縦型 feed をこのアプリの主導線として固定する。" })).toBeVisible();
});

test("subscriptions と profile route も shell から辿れる", async ({ page }) => {
  await page.goto("/subscriptions");
  await expect(page.getByRole("heading", { name: "購読後の継続視聴面を独立 route として先に確保する。" })).toBeVisible();

  await page.getByRole("link", { name: /profile account/i }).click();
  await expect(page).toHaveURL(/\/profile$/);
  await expect(page.getByRole("heading", { name: "profile まわりを consumer shell の一部として先に区切る。" })).toBeVisible();
});

test("non-consumer routes と 404 が正しく描画される", async ({ page }) => {
  await page.goto("/creator/atelier-rin");
  await expect(page.getByRole("heading", { name: "Atelier Rin の公開面と限定導線を切り分ける。" })).toBeVisible();

  await page.goto("/admin");
  await expect(page.getByRole("heading", { name: "consumer shell とは分離した admin route。" })).toBeVisible();

  await page.goto("/missing-route");
  await expect(page.getByRole("heading", { name: "指定された route はまだ用意されていません。" })).toBeVisible();
});
