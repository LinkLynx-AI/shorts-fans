import {
  render,
  screen,
} from "@testing-library/react";

import { AdminReviewNavigation } from "./admin-review-navigation";

describe("AdminReviewNavigation", () => {
  it("links the admin review surfaces and marks the active surface", () => {
    render(<AdminReviewNavigation active="submission-reviews" />);
    const links = screen.getAllByRole("link");

    expect(screen.getByRole("link", { name: /Admin 審査入口/i })).toHaveAttribute("href", "/admin");
    expect(screen.getByRole("link", { name: /Creator 審査 登録申請/i })).toHaveAttribute(
      "href",
      "/admin/creator-reviews",
    );
    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "href",
      "/admin/submission-reviews",
    );
    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "data-prefetch",
      "false",
    );
    expect(links.every((link) => link.getAttribute("data-prefetch") === "false")).toBe(true);
    expect(screen.getByRole("link", { name: /Video 審査 main \/ short/i })).toHaveAttribute(
      "aria-current",
      "page",
    );
  });
});
