import { render, screen } from "@testing-library/react";

import { getCreatorModeUnauthenticatedState } from "../model/creator-mode-shell";
import {
  CreatorModeWorkspaceFrame,
  CreatorShellBlockedState,
} from "./creator-mode-shell-blocked-state";

describe("creator mode shell frames", () => {
  it("keeps the width cap while removing the desktop height clamp", () => {
    render(
      <CreatorModeWorkspaceFrame>
        <div>creator workspace</div>
      </CreatorModeWorkspaceFrame>,
    );

    const content = screen.getByText("creator workspace");
    const frame = content.parentElement;
    const shell = frame?.parentElement;

    expect(frame?.className).toContain("max-w-[408px]");
    expect(frame?.className).not.toContain("sm:min-h-[calc(100svh-48px)]");
    expect(frame?.className).not.toContain("sm:rounded-[36px]");
    expect(shell?.className).not.toContain("sm:py-6");
  });

  it("renders blocked state inside the shared creator frame", () => {
    const { container } = render(
      <CreatorShellBlockedState state={getCreatorModeUnauthenticatedState()} />,
    );

    const shell = container.querySelector("main");

    expect(screen.getByRole("heading", { name: "creator mode を開くにはログインが必要です。" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインへ進む" })).toHaveAttribute("href", "/login");
    expect(shell?.className).not.toContain("sm:py-6");
  });
});
