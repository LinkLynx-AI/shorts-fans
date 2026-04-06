import { render, screen } from "@testing-library/react";

import {
  CurrentViewerProvider,
  useCurrentViewer,
} from "@/entities/viewer";

function CurrentViewerConsumer() {
  const currentViewer = useCurrentViewer();

  return (
    <p>
      {currentViewer
        ? `${currentViewer.id}:${currentViewer.activeMode}`
        : "anonymous"}
    </p>
  );
}

describe("CurrentViewerProvider", () => {
  it("provides the current viewer state to descendants", () => {
    render(
      <CurrentViewerProvider
        currentViewer={{
          activeMode: "creator",
          canAccessCreatorMode: true,
          id: "viewer_123",
        }}
      >
        <CurrentViewerConsumer />
      </CurrentViewerProvider>,
    );

    expect(screen.getByText("viewer_123:creator")).toBeInTheDocument();
  });

  it("allows unauthenticated bootstrap state", () => {
    render(
      <CurrentViewerProvider currentViewer={null}>
        <CurrentViewerConsumer />
      </CurrentViewerProvider>,
    );

    expect(screen.getByText("anonymous")).toBeInTheDocument();
  });
});
