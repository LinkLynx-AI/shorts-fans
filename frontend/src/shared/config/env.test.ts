import { describe, expect, it, vi } from "vitest";

import { getClientEnv, getOptionalClientEnv, parseClientEnv, parseOptionalClientEnv } from "@/shared/config";

describe("parseClientEnv", () => {
  it("returns parsed env values when the contract is satisfied", () => {
    expect(
      parseClientEnv({
        NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
      }),
    ).toEqual({
      NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
    });
  });

  it("throws when the public API base URL is missing", () => {
    expect(() =>
      parseClientEnv({
        NEXT_PUBLIC_API_BASE_URL: undefined,
      }),
    ).toThrowError();
  });

  it("reads the public API base URL from process.env", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "https://api.example.com");

    expect(getClientEnv()).toEqual({
      NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
    });
  });
});

describe("parseOptionalClientEnv", () => {
  it("returns an empty object when the public API base URL is missing", () => {
    expect(
      parseOptionalClientEnv({
        NEXT_PUBLIC_API_BASE_URL: undefined,
      }),
    ).toEqual({});
  });

  it("returns parsed values when the public API base URL exists", () => {
    expect(
      parseOptionalClientEnv({
        NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
      }),
    ).toEqual({
      NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
    });
  });

  it("reads optional values from process.env", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", undefined);
    expect(getOptionalClientEnv()).toEqual({});

    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "https://api.example.com");
    expect(getOptionalClientEnv()).toEqual({
      NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
    });
  });
});
