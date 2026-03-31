import { render } from "@testing-library/react";

import { Separator } from "@/shared/ui";

describe("Separator", () => {
  it("supports vertical orientation", () => {
    const { container } = render(<Separator orientation="vertical" />);

    expect(container.firstChild).toHaveClass("w-px");
  });
});
