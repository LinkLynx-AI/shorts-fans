import {
  fetchFanProfileLibraryPage,
  type FanLibraryItem,
} from "@/entities/fan-profile";

export type LibraryMainReelState = {
  initialIndex: number;
  items: readonly FanLibraryItem[];
};

type LoadLibraryMainReelStateOptions = {
  mainId: string;
  sessionToken?: string | undefined;
};

/**
 * fan profile library 起点の main reel state を取得する。
 */
export async function loadLibraryMainReelState({
  mainId,
  sessionToken,
}: LoadLibraryMainReelStateOptions): Promise<LibraryMainReelState | null> {
  const page = await fetchFanProfileLibraryPage({
    sessionToken,
  });
  const initialIndex = page.items.findIndex((item) => item.main.id === mainId);

  if (initialIndex < 0) {
    return null;
  }

  return {
    initialIndex,
    items: page.items,
  };
}
