import { render, screen } from "@testing-library/react";

import { mockedUsePathname } from "@/test/mocks/next-navigation";
import { ConsumerShell } from "@/widgets/consumer-shell";

describe("ConsumerShell", () => {
  it("renders the shared shell and active navigation state", () => {
    mockedUsePathname.mockReturnValue("/shorts");

    render(
      <ConsumerShell>
        <div>shell body</div>
      </ConsumerShell>,
    );

    expect(screen.getByText("shell body")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /shorts primary lane/i })).toHaveAttribute(
      "aria-current",
      "page",
    );
  });
});
