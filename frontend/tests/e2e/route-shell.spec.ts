import { expect, test } from "@playwright/test";

test("fan shell routes render and unlock flow works", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole("link", { name: "おすすめ" })).toHaveAttribute("aria-current", "page");
  await expect(page.getByRole("button", { name: /Unlock/i })).toBeVisible();
  await expect(page.getByText("Mina Rei")).toBeVisible();

  await page.getByRole("button", { name: /Unlock/i }).click();
  const paywall = page.getByRole("dialog", { name: "quiet rooftop preview の続きを見る" });
  await expect(paywall).toBeVisible();
  await expect(paywall.getByRole("button", { name: "Unlock ¥1,800 | 8分" })).toBeDisabled();
  await page.getByLabel("18歳以上であり、年齢確認に同意する").check();
  await page.getByLabel("利用規約とポリシーに同意し、確認面なしで main 再生へ進む").check();
  await paywall.getByRole("button", { name: "Unlock ¥1,800 | 8分" }).click();
  await expect(page).toHaveURL(/\/mains\/main_mina_quiet_rooftop\?fromShortId=rooftop&grant=/);
  await expect(page.getByText("Playing main")).toBeVisible();
  await page.getByRole("button", { name: "Back" }).click();
  await expect(page).toHaveURL(/\/$/);

  await page.getByRole("link", { name: "検索" }).click();
  await expect(page).toHaveURL(/\/search$/);
  await expect(page.getByRole("heading", { name: "Creator search structure" })).toBeVisible();

  await page.getByRole("link", { name: "マイ" }).click();
  await expect(page).toHaveURL(/\/fan$/);
  await expect(page.getByRole("heading", { name: "Fan hub structure" })).toBeVisible();

  await page.goto("/shorts/rooftop");
  await expect(page.getByRole("link", { name: /Back/i })).toBeVisible();
  await expect(page.getByText("quiet rooftop preview.")).toBeVisible();

  await page.goto("/shorts/softlight");
  await page.getByRole("button", { name: /Continue main/i }).click();
  await expect(page).toHaveURL(/\/mains\/main_aoi_blue_balcony\?fromShortId=softlight&grant=/);
  await page.getByRole("button", { name: "Back" }).click();

  await page.goto("/shorts/afterrain");
  await page.getByRole("button", { name: /Unlock/i }).click();
  await expect(page).toHaveURL(/\/mains\/main_sora_after_rain\?fromShortId=afterrain&grant=/);
  await page.getByRole("button", { name: "Back" }).click();

  await page.goto("/shorts/balcony");
  await page.getByRole("button", { name: /Owner preview/i }).click();
  await expect(page).toHaveURL(/\/mains\/main_aoi_blue_balcony\?fromShortId=balcony&grant=/);
  await page.getByRole("button", { name: "Back" }).click();

  await page.goto("/mains/main_mina_quiet_rooftop");
  await expect(page.getByRole("heading", { name: "指定された surface はまだ用意されていません。" })).toBeVisible();

  await page.goto("/mains/main_mina_quiet_rooftop?fromShortId=rooftop&grant=invalid");
  await expect(page.getByRole("heading", { name: "この main はまだ unlock されていません。" })).toBeVisible();

  await page.goto("/creators/mina");
  await expect(page.getByRole("heading", { name: "Creator profile structure" })).toBeVisible();
});

test("undefined routes fall back to the shared not-found page", async ({ page }) => {
  await page.goto("/missing-route");
  await expect(page.getByRole("heading", { name: "指定された surface はまだ用意されていません。" })).toBeVisible();
  await page.getByRole("link", { name: "feed に戻る" }).click();
  await expect(page).toHaveURL(/\/$/);
});
