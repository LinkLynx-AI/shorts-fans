import { fireEvent, render, screen } from "@testing-library/react";

import ErrorPage from "./error";

describe("ErrorPage", () => {
  it("renders the boundary UI and retries when requested", () => {
    const unstableRetry = vi.fn();
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    render(
      <ErrorPage
        error={Object.assign(new Error("boom"), {
          digest: "digest-123",
        })}
        unstable_retry={unstableRetry}
      />,
    );

    expect(screen.getByText("digest: digest-123")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "再試行する" }));

    expect(unstableRetry).toHaveBeenCalledTimes(1);

    consoleError.mockRestore();
  });
});
