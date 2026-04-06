import { render, screen } from "@testing-library/react";

import { buildMockMainPlaybackGrantContext } from "@/features/unlock-entry";
import { issueMockSignedToken } from "@/shared/lib/mock-signed-token";

import MainPlaybackPage from "./page";

describe("MainPlaybackPage", () => {
  it("renders the locked state when a signed grant is replayed for a different short context", async () => {
    const mismatchedGrant = issueMockSignedToken(
      buildMockMainPlaybackGrantContext("main_mina_quiet_rooftop", "mirror", "purchased"),
    );

    render(
      await MainPlaybackPage({
        params: Promise.resolve({
          mainId: "main_mina_quiet_rooftop",
        }),
        searchParams: Promise.resolve({
          fromShortId: "rooftop",
          grant: mismatchedGrant,
        }),
      }),
    );

    expect(screen.getByRole("heading", { name: "この main はまだ unlock されていません。" })).toBeInTheDocument();
    expect(screen.queryByText("Playing main")).not.toBeInTheDocument();
  });
});
