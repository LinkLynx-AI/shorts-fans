import { describe, expect, it } from "vitest";

import { isNavigationItemActive } from "./is-navigation-item-active";

describe("isNavigationItemActive", () => {
  it("matches non-nested routes exactly", () => {
    expect(isNavigationItemActive("/home", "/home")).toBe(true);
    expect(isNavigationItemActive("/shorts", "/home")).toBe(false);
  });

  it("matches nested routes for non-root tabs", () => {
    expect(isNavigationItemActive("/subscriptions", "/subscriptions")).toBe(true);
    expect(isNavigationItemActive("/subscriptions/history", "/subscriptions")).toBe(true);
    expect(isNavigationItemActive("/profile", "/subscriptions")).toBe(false);
  });
});
