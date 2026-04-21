import {
  render,
  screen,
} from "@testing-library/react";

import AdminPage from "./page";

vi.mock("./_lib/admin-ui-access", () => ({
  assertAdminUiEnabled: vi.fn(),
}));

describe("AdminPage", () => {
  it("renders UI links to creator and video review queues", () => {
    render(<AdminPage />);
    const reviewQueueLinks = screen
      .getAllByRole("link")
      .filter((link) => ["/admin/creator-reviews", "/admin/submission-reviews"].includes(link.getAttribute("href") ?? ""));

    expect(screen.getByRole("heading", { name: "審査 surface を選択する" })).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /Creator 審査/i })[0]).toHaveAttribute(
      "href",
      "/admin/creator-reviews",
    );
    expect(screen.getAllByRole("link", { name: /Video 審査/i })[0]).toHaveAttribute(
      "href",
      "/admin/submission-reviews",
    );
    expect(reviewQueueLinks).toHaveLength(4);
    expect(reviewQueueLinks.every((link) => link.getAttribute("data-prefetch") === "false")).toBe(true);
  });
});
