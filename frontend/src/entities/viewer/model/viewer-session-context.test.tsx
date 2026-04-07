import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";

import {
  useHasViewerSession,
  useSetViewerSession,
  ViewerSessionProvider,
} from "@/entities/viewer";

function ViewerSessionConsumer() {
  return <p>{useHasViewerSession() ? "session-present" : "session-missing"}</p>;
}

function ViewerSessionOverrideConsumer() {
  const setViewerSession = useSetViewerSession();

  return (
    <button onClick={() => setViewerSession(false)} type="button">
      clear session
    </button>
  );
}

let mountSequence = 0;

function ViewerSessionMountMarker() {
  const [mountId] = useState(() => {
    mountSequence += 1;
    return mountSequence;
  });

  return <p>{`mount-${mountId}`}</p>;
}

describe("ViewerSessionProvider", () => {
  beforeEach(() => {
    mountSequence = 0;
  });

  it("provides the presence of the viewer session cookie", () => {
    render(
      <ViewerSessionProvider hasSession>
        <ViewerSessionConsumer />
      </ViewerSessionProvider>,
    );

    expect(screen.getByText("session-present")).toBeInTheDocument();
  });

  it("defaults descendants to unauthenticated when no session exists", () => {
    render(
      <ViewerSessionProvider hasSession={false}>
        <ViewerSessionConsumer />
      </ViewerSessionProvider>,
    );

    expect(screen.getByText("session-missing")).toBeInTheDocument();
  });

  it("resets overridden state when the parent remounts it with a new bootstrap key", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <ViewerSessionProvider hasSession key="session-present">
        <ViewerSessionConsumer />
        <ViewerSessionOverrideConsumer />
        <ViewerSessionMountMarker />
      </ViewerSessionProvider>,
    );

    expect(screen.getByText("mount-1")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "clear session" }));
    expect(screen.getByText("session-missing")).toBeInTheDocument();
    expect(screen.getByText("mount-1")).toBeInTheDocument();

    rerender(
      <ViewerSessionProvider hasSession={false} key="session-missing">
        <ViewerSessionConsumer />
        <ViewerSessionOverrideConsumer />
        <ViewerSessionMountMarker />
      </ViewerSessionProvider>,
    );

    expect(screen.getByText("session-missing")).toBeInTheDocument();
    expect(screen.getByText("mount-2")).toBeInTheDocument();
  });
});
