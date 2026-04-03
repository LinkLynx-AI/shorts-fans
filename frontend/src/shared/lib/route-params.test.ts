import { getSingleQueryParam } from "@/shared/lib";

describe("route params helpers", () => {
  it("returns the first query param value", () => {
    expect(getSingleQueryParam(undefined)).toBeUndefined();
    expect(getSingleQueryParam("mina")).toBe("mina");
    expect(getSingleQueryParam(["mina", "aoi"])).toBe("mina");
  });
});
