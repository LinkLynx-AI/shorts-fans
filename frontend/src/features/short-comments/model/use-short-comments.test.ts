import {
  countCommentBodyCharacters,
  limitCommentDraftInput,
  maxCommentBodyLength,
  maxCommentDraftInputLength,
} from "./use-short-comments";

describe("short comment body helpers", () => {
  it("counts visible characters after trimming surrounding whitespace", () => {
    expect(countCommentBodyCharacters(" \n\t hello \t\n ")).toBe(5);
    expect(countCommentBodyCharacters(" \n\t ")).toBe(0);
  });

  it("distinguishes the 500 and 501 character boundaries", () => {
    expect(countCommentBodyCharacters("a".repeat(maxCommentBodyLength))).toBe(maxCommentBodyLength);
    expect(countCommentBodyCharacters("a".repeat(maxCommentBodyLength + 1))).toBe(maxCommentBodyLength + 1);
  });

  it("counts emoji as one visible character", () => {
    expect(countCommentBodyCharacters("😀".repeat(maxCommentBodyLength))).toBe(maxCommentBodyLength);
    expect(countCommentBodyCharacters("😀".repeat(maxCommentBodyLength + 1))).toBe(maxCommentBodyLength + 1);
  });

  it("caps draft input without splitting surrogate pairs", () => {
    const draft = `${"😀".repeat(maxCommentDraftInputLength)}😀`;

    expect(limitCommentDraftInput(draft)).toBe("😀".repeat(maxCommentDraftInputLength));
  });
});
