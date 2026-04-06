import { act, fireEvent, render, screen } from "@testing-library/react";

import type { CreatorSearchState } from "@/features/creator-search";
import { CreatorSearchPanel } from "@/features/creator-search";
import { getCreatorSearchResults } from "@/entities/creator";

vi.mock("@/entities/creator", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/entities/creator")>();

  return {
    ...actual,
    getCreatorSearchResults: vi.fn(),
  };
});

const mockedGetCreatorSearchResults = vi.mocked(getCreatorSearchResults);

const initialState: CreatorSearchState = {
  items: [
    {
      avatar: null,
      bio: "quiet rooftop と hotel light の preview を軸に投稿。",
      displayName: "Mina Rei",
      handle: "@minarei",
      id: "creator_mina_rei",
    },
  ],
  kind: "ready" as const,
  query: "mina",
};

describe("CreatorSearchPanel", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    mockedGetCreatorSearchResults.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders the initial query and recent creators", () => {
    render(<CreatorSearchPanel initialQuery="mina" initialState={initialState} />);

    expect(screen.getByRole("searchbox", { name: "クリエイターを検索" })).toHaveValue("mina");
    expect(screen.queryByText("最近")).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mina Rei/i })).toHaveAttribute(
      "href",
      "/creators/creator_mina_rei?from=search&q=mina",
    );
    expect(screen.queryByText("open")).not.toBeInTheDocument();
  });

  it("applies the filter after a short delay", async () => {
    const recentState: CreatorSearchState = {
      items: [
        {
          avatar: null,
          bio: "soft light と close framing の short を中心に更新中。",
          displayName: "Aoi N",
          handle: "@aoina",
          id: "creator_aoi_n",
        },
        {
          avatar: null,
          bio: "quiet rooftop と hotel light の preview を軸に投稿。",
          displayName: "Mina Rei",
          handle: "@minarei",
          id: "creator_mina_rei",
        },
      ],
      kind: "ready",
      query: "",
    };

    mockedGetCreatorSearchResults.mockResolvedValue({
      items: [
        {
          avatar: null,
          bio: "after rain と balcony mood の short をまとめています。",
          displayName: "Sora Vale",
          handle: "@soravale",
          id: "creator_sora_vale",
        },
      ],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      query: "sora",
      requestId: "req_search_filtered_001",
    });

    render(
      <CreatorSearchPanel
        initialQuery=""
        initialState={recentState}
      />,
    );

    expect(screen.getByText("最近")).toBeInTheDocument();

    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "sora" },
    });

    expect(screen.queryByText("最近")).not.toBeInTheDocument();
    expect(screen.queryByText("Sora Vale")).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("読み込み中...");

    await act(async () => {
      vi.advanceTimersByTime(250);
      await Promise.resolve();
    });

    expect(screen.getByText("Sora Vale")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Sora Vale/i })).toHaveAttribute(
      "href",
      "/creators/creator_sora_vale?from=search&q=sora",
    );
    expect(screen.queryByText("Aoi N")).not.toBeInTheDocument();
  });

  it("renders the empty state when no creators match", async () => {
    mockedGetCreatorSearchResults.mockResolvedValue({
      items: [],
      page: {
        hasNext: false,
        nextCursor: null,
      },
      query: "missing",
      requestId: "req_search_empty_001",
    });

    render(<CreatorSearchPanel initialQuery="" initialState={initialState} />);

    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "missing" },
    });

    await act(async () => {
      vi.advanceTimersByTime(250);
      await Promise.resolve();
    });

    expect(screen.getByText("一致する creator は見つかりませんでした。")).toBeInTheDocument();
  });

  it("renders the error state and retries the request", async () => {
    let shouldSucceed = false;
    mockedGetCreatorSearchResults.mockImplementation(async () => {
      if (!shouldSucceed) {
        throw new Error("boom");
      }

      return {
        items: [
          {
            avatar: null,
            bio: "Public shorts から paid main へつながる creator mock profile.",
            displayName: "Mika Aoi",
            handle: "@mikaaoi",
            id: "creator_11111111111111111111111111111111",
          },
        ],
        page: {
          hasNext: false,
          nextCursor: null,
        },
        query: "mika",
        requestId: "req_search_retry_001",
      };
    });

    render(<CreatorSearchPanel initialQuery="" initialState={initialState} />);

    fireEvent.change(screen.getByRole("searchbox"), {
      target: { value: "mika" },
    });

    await act(async () => {
      vi.advanceTimersByTime(250);
      await Promise.resolve();
    });

    expect(screen.getByText("検索結果を読み込めませんでした。もう一度お試しください。")).toBeInTheDocument();

    shouldSucceed = true;
    fireEvent.click(screen.getByRole("button", { name: "再読み込み" }));

    await act(async () => {
      await Promise.resolve();
    });

    expect(screen.getByText("Mika Aoi")).toBeInTheDocument();
  });
});
