import { render, screen } from "@testing-library/react";

import {
  useHasViewerSession,
  ViewerSessionProvider,
} from "@/entities/viewer";

function ViewerSessionConsumer() {
  return <p>{useHasViewerSession() ? "session-present" : "session-missing"}</p>;
}

describe("ViewerSessionProvider", () => {
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
});
