import { render, screen } from "@testing-library/react";

import Loading from "./loading";

describe("Loading", () => {
  it("renders the loading shell placeholders", () => {
    const { container } = render(<Loading />);

    expect(container.querySelectorAll(".animate-pulse")).toHaveLength(12);
    expect(screen.getByRole("main")).toBeInTheDocument();
  });
});
