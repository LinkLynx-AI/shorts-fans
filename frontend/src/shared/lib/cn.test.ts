import { describe, expect, it } from "vitest";

import { cn } from "@/shared/lib";

describe("cn", () => {
  it("merges tailwind classes predictably", () => {
    expect(cn("px-3", "px-5", false && "hidden", "text-sm")).toBe("px-5 text-sm");
  });
});
