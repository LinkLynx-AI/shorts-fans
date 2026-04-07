import { render, screen } from "@testing-library/react";

import { FanAuthEntryShell } from "@/widgets/fan-auth-entry-shell";

describe("FanAuthEntryShell", () => {
  it("renders the protected fan login entry shell", () => {
    render(<FanAuthEntryShell />);

    expect(screen.getByRole("heading", { name: "続けるにはログインが必要です" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Email" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "サインインを続ける" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "feed に戻る" })).toHaveAttribute("href", "/");
  });
});
