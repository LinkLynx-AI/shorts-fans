import { render, screen } from "@testing-library/react";
import type { ComponentPropsWithoutRef } from "react";
import { vi } from "vitest";

import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui";

vi.mock("@radix-ui/react-avatar", () => ({
  Root: ({ children, ...props }: ComponentPropsWithoutRef<"span">) => <span {...props}>{children}</span>,
  Image: ({ alt, ...props }: ComponentPropsWithoutRef<"img">) => <span data-alt={alt} data-testid="avatar-image" {...props} />,
  Fallback: ({ children, ...props }: ComponentPropsWithoutRef<"span">) => <span {...props}>{children}</span>,
}));

describe("Avatar", () => {
  it("renders the root with default and custom styles", () => {
    const { container } = render(<Avatar className="custom-avatar" />);

    expect(container.firstChild).toHaveClass("size-12");
    expect(container.firstChild).toHaveClass("custom-avatar");
  });

  it("renders an image element when source props are provided", () => {
    render(
      <Avatar>
        <AvatarImage alt="creator avatar" src="https://example.com/avatar.png" />
      </Avatar>,
    );

    expect(screen.getByTestId("avatar-image")).toHaveClass("object-cover");
    expect(screen.getByTestId("avatar-image")).toHaveAttribute("data-alt", "creator avatar");
  });

  it("renders a fallback label when no image is available", () => {
    render(
      <Avatar>
        <AvatarFallback className="custom-fallback">cf</AvatarFallback>
      </Avatar>,
    );

    expect(screen.getByText("cf")).toHaveClass("custom-fallback");
    expect(screen.getByText("cf")).toHaveClass("uppercase");
  });
});
