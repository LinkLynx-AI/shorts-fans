import { getMainPlaybackSurfaceById } from "@/widgets/main-playback-surface";

describe("mock main playback surface", () => {
  it("uses unlocked playback state for a resumed short", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "unlocked");

    expect(surface).toBeDefined();
    expect(surface?.access.status).toBe("unlocked");
    expect(surface?.resumePositionSeconds).toBe(198);
    expect(surface?.themeShort.id).toBe("softlight");
  });

  it("uses owner preview state when the entry short is owner-only", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "balcony", "owner");

    expect(surface).toBeDefined();
    expect(surface?.access.status).toBe("owner");
    expect(surface?.resumePositionSeconds).toBeNull();
    expect(surface?.themeShort.id).toBe("balcony");
  });

  it("blocks playback when the entry short context is missing or invalid", () => {
    expect(getMainPlaybackSurfaceById("main_mina_quiet_rooftop")).toBeUndefined();
    expect(getMainPlaybackSurfaceById("main_mina_quiet_rooftop", "softlight", "unlocked")).toBeUndefined();
  });
});
