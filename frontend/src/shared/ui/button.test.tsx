import { render, screen } from "@testing-library/react";

import { Button } from "@/shared/ui";

describe("Button", () => {
  it("renders as a button by default", () => {
    render(<Button>action</Button>);

    expect(screen.getByRole("button", { name: "action" })).toBeInTheDocument();
  });

  it("renders as its child when asChild is enabled", () => {
    render(
      <Button asChild variant="secondary">
        <a href="/home">home</a>
      </Button>,
    );

    expect(screen.getByRole("link", { name: "home" })).toHaveAttribute("href", "/home");
  });
});
