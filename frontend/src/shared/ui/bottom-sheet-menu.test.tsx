import userEvent from "@testing-library/user-event";
import {
  render,
  screen,
  waitFor,
} from "@testing-library/react";

import {
  BottomSheetMenu,
  BottomSheetMenuAction,
  BottomSheetMenuClose,
  BottomSheetMenuGroup,
} from "@/shared/ui";

describe("BottomSheetMenu", () => {
  it("opens from the trigger and closes from a close-wrapped action", async () => {
    const user = userEvent.setup();

    render(
      <BottomSheetMenu
        description="menu description"
        title="menu title"
        trigger={<button type="button">open menu</button>}
      >
        <BottomSheetMenuGroup>
          <BottomSheetMenuClose asChild>
            <BottomSheetMenuAction>
              <span>close action</span>
            </BottomSheetMenuAction>
          </BottomSheetMenuClose>
        </BottomSheetMenuGroup>
      </BottomSheetMenu>,
    );

    const trigger = screen.getByRole("button", { name: "open menu" });

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

    await user.click(trigger);

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "close action" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "close action" }));

    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
    expect(trigger).toHaveFocus();
  });
});
