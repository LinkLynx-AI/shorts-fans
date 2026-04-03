import { expect, test } from "@playwright/test";

test("fan shell routes render and navigation works", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("link", { name: "おすすめ" })).toHaveAttribute("aria-current", "page");
  await expect(page.getByRole("link", { name: /Unlock/i })).toBeVisible();
  await expect(page.getByText("Mina Rei")).toBeVisible();

  await page.getByRole("link", { name: /Search/i }).click();
  await expect(page).toHaveURL(/\/search$/);
  await expect(page.getByRole("heading", { name: "Creator search structure" })).toBeVisible();

  await page.getByRole("link", { name: /My/i }).click();
  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("heading", { name: "Fan hub structure" })).toBeVisible();

  await page.goto("/shorts/rooftop");
  await expect(page.getByRole("link", { name: /Back/i })).toBeVisible();
  await expect(page.getByText("quiet rooftop preview.")).toBeVisible();

  await page.goto("/creators/mina");
  await expect(page.getByRole("heading", { name: "Creator profile structure" })).toBeVisible();
});

test("undefined routes fall back to the shared not-found page", async ({ page }) => {
  await page.goto("/missing-route");
  await expect(page.getByRole("heading", { name: "指定された surface はまだ用意されていません。" })).toBeVisible();
  await page.getByRole("link", { name: "feed に戻る" }).click();
  await expect(page).toHaveURL(/\/$/);
});
