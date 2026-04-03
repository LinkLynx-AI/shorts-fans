import { getEnumQueryParam, getSingleQueryParam } from "@/shared/lib";

describe("route params helpers", () => {
  it("returns the first query param value", () => {
    expect(getSingleQueryParam(undefined)).toBeUndefined();
    expect(getSingleQueryParam("mina")).toBe("mina");
    expect(getSingleQueryParam(["mina", "aoi"])).toBe("mina");
  });

  it("returns only allowed enum query params", () => {
    expect(getEnumQueryParam("search", ["feed", "search", "short"])).toBe("search");
    expect(getEnumQueryParam("twitter", ["feed", "search", "short"])).toBeUndefined();
    expect(getEnumQueryParam(["recommended", "following"], ["following", "recommended"])).toBe("recommended");
  });
});
