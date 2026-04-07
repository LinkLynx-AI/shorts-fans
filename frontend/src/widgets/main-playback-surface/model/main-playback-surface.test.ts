import { getMainPlaybackSurfaceById } from "@/widgets/main-playback-surface";

import {
  getMainPlaybackStatusCopy,
  getMainPlaybackStatusMeta,
  getMainPlaybackStatusTitle,
} from "./main-playback-surface";

describe("main playback surface status helpers", () => {
  it("returns owner preview copy for owner surfaces", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "balcony", "owner");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    expect(getMainPlaybackStatusTitle(surface)).toBe("Owner preview");
    expect(getMainPlaybackStatusCopy(surface)).toBe("purchase confirmation is skipped for your own main");
    expect(getMainPlaybackStatusMeta(surface)).toBe("Owner preview");
  });

  it("returns resume copy and timestamp for purchased surfaces with resume position", () => {
    const surface = getMainPlaybackSurfaceById("main_aoi_blue_balcony", "softlight", "purchased");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    expect(getMainPlaybackStatusTitle(surface)).toBe("Playing main");
    expect(getMainPlaybackStatusCopy(surface)).toBe("resume without another confirmation");
    expect(getMainPlaybackStatusMeta(surface)).toBe("3:18");
  });

  it("returns duration copy for purchased surfaces without resume position", () => {
    const surface = getMainPlaybackSurfaceById("main_sora_after_rain", "afterrain", "purchased");

    expect(surface).toBeDefined();

    if (!surface) {
      throw new Error("fixture missing");
    }

    expect(getMainPlaybackStatusTitle(surface)).toBe("Playing main");
    expect(getMainPlaybackStatusCopy(surface)).toBe("continue from this short without another confirmation");
    expect(getMainPlaybackStatusMeta(surface)).toBe("9分");
  });
});
