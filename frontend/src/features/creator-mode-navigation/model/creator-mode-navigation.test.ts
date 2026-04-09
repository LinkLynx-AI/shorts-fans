import {
  getCreatorModeNavigationItems,
  isCreatorModeNavigationAvailable,
  resolveActiveCreatorModeNavigation,
} from "@/features/creator-mode-navigation";

describe("creator mode navigation", () => {
  it("returns the creator-side primary navigation vocabulary", () => {
    expect(getCreatorModeNavigationItems()).toEqual([
      expect.objectContaining({
        href: "/creator",
        key: "dashboard",
        label: "Dashboard",
      }),
      expect.objectContaining({
        href: "/creator/upload",
        key: "upload",
        label: "Upload",
      }),
      expect.objectContaining({
        href: "/creator/linkage",
        key: "linkage",
        label: "Linkage",
      }),
      expect.objectContaining({
        href: "/creator/review",
        key: "review",
        label: "Review",
      }),
    ]);
  });

  it("marks the dashboard and upload routes as available in this run", () => {
    expect(isCreatorModeNavigationAvailable("dashboard")).toBe(true);
    expect(isCreatorModeNavigationAvailable("upload")).toBe(true);
    expect(isCreatorModeNavigationAvailable("linkage")).toBe(false);
    expect(isCreatorModeNavigationAvailable("review")).toBe(false);
  });

  it("resolves the active item from creator mode pathnames", () => {
    expect(resolveActiveCreatorModeNavigation("/creator")).toBe("dashboard");
    expect(resolveActiveCreatorModeNavigation("/creator/upload")).toBe("upload");
    expect(resolveActiveCreatorModeNavigation("/creator/linkage/main")).toBe("linkage");
    expect(resolveActiveCreatorModeNavigation("/creator/review/moderation")).toBe("review");
  });
});
