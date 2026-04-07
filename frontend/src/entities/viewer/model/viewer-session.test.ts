import {
  hasViewerSession,
  readViewerSessionToken,
} from "@/entities/viewer";

describe("viewer session cookie helpers", () => {
  it("reads the session token from a cookie header", () => {
    expect(readViewerSessionToken("foo=bar; shorts_fans_session=raw-session-token; hello=world")).toBe(
      "raw-session-token",
    );
  });

  it("returns null when the viewer session cookie is missing", () => {
    expect(readViewerSessionToken("foo=bar")).toBeNull();
  });

  it("treats empty viewer session values as unauthenticated", () => {
    expect(hasViewerSession("shorts_fans_session=")).toBe(false);
  });
});
