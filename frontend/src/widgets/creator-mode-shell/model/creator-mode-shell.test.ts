import {
  getMockCreatorModeShellState,
  resolveCreatorModeShellState,
} from "@/widgets/creator-mode-shell";

describe("creator mode shell state", () => {
  it("returns the ready shell state for creator mode viewers", () => {
    expect(
      resolveCreatorModeShellState({
        activeMode: "creator",
        canAccessCreatorMode: true,
        id: "viewer_creator_001",
      }),
    ).toEqual(
      expect.objectContaining({
        activeNavigation: "dashboard",
        creator: expect.objectContaining({
          id: "creator_mina_rei",
        }),
        kind: "ready",
        workspace: expect.objectContaining({
          summaryStats: expect.arrayContaining([
            expect.objectContaining({
              label: "revenue",
              value: "¥120K",
            }),
          ]),
        }),
      }),
    );
  });

  it("requires authentication before opening the creator shell", () => {
    expect(resolveCreatorModeShellState(null)).toEqual(
      expect.objectContaining({
        ctaHref: "/login",
        kind: "unauthenticated",
      }),
    );
  });

  it("blocks viewers without creator capability", () => {
    expect(
      resolveCreatorModeShellState({
        activeMode: "fan",
        canAccessCreatorMode: false,
        id: "viewer_fan_001",
      }),
    ).toEqual(
      expect.objectContaining({
        kind: "capability_required",
        title: "creator mode はまだ利用できません。",
      }),
    );
  });

  it("shows a mode mismatch state when creator capability exists but active mode is still fan", () => {
    expect(
      resolveCreatorModeShellState({
        activeMode: "fan",
        canAccessCreatorMode: true,
        id: "viewer_creator_002",
      }),
    ).toEqual(
      expect.objectContaining({
        kind: "mode_required",
        title: "creator mode に切り替えてから開いてください。",
      }),
    );
  });

  it("builds the shell from the shared mock creator profile", () => {
    const state = getMockCreatorModeShellState();

    expect(state.creator.id).toBe("creator_mina_rei");
    expect(state.workspace.summaryStats).toHaveLength(3);
    expect(state.workspace.managedCollections.itemsByTab.shorts).toHaveLength(3);
    expect(state.workspace.managedCollections.itemsByTab.main).toHaveLength(2);
  });
});
