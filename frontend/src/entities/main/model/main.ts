export type MainId = string;

export type MainMediaAsset = {
  durationSeconds: number;
  id: string;
  kind: "video";
  posterUrl: string | null;
  url: string;
};

export type MainDetail = {
  durationSeconds: number;
  id: MainId;
  leadShortId: string;
  media: MainMediaAsset;
  priceJpy: number;
};

export type MainSummary = Pick<MainDetail, "durationSeconds" | "id" | "priceJpy">;

const mains = [
  {
    durationSeconds: 540,
    id: "main_sora_after_rain",
    leadShortId: "afterrain",
    media: {
      durationSeconds: 540,
      id: "asset_main_sora_after_rain",
      kind: "video",
      posterUrl: "https://cdn.example.com/mains/sora-after-rain-poster.jpg",
      url: "https://cdn.example.com/mains/sora-after-rain.mp4",
    },
    priceJpy: 2100,
  },
  {
    durationSeconds: 720,
    id: "main_aoi_blue_balcony",
    leadShortId: "softlight",
    media: {
      durationSeconds: 720,
      id: "asset_main_aoi_blue_balcony",
      kind: "video",
      posterUrl: "https://cdn.example.com/mains/aoi-blue-balcony-poster.jpg",
      url: "https://cdn.example.com/mains/aoi-blue-balcony.mp4",
    },
    priceJpy: 2200,
  },
  {
    durationSeconds: 660,
    id: "main_mina_hotel_mirror",
    leadShortId: "mirror",
    media: {
      durationSeconds: 660,
      id: "asset_main_mina_hotel_mirror",
      kind: "video",
      posterUrl: "https://cdn.example.com/mains/mina-hotel-mirror-poster.jpg",
      url: "https://cdn.example.com/mains/mina-hotel-mirror.mp4",
    },
    priceJpy: 2400,
  },
  {
    durationSeconds: 480,
    id: "main_sora_poolside_cut",
    leadShortId: "poolcut",
    media: {
      durationSeconds: 480,
      id: "asset_main_sora_poolside_cut",
      kind: "video",
      posterUrl: "https://cdn.example.com/mains/sora-poolside-cut-poster.jpg",
      url: "https://cdn.example.com/mains/sora-poolside-cut.mp4",
    },
    priceJpy: 1900,
  },
  {
    durationSeconds: 480,
    id: "main_mina_quiet_rooftop",
    leadShortId: "rooftop",
    media: {
      durationSeconds: 480,
      id: "asset_main_mina_quiet_rooftop",
      kind: "video",
      posterUrl: "https://cdn.example.com/mains/mina-quiet-rooftop-poster.jpg",
      url: "https://cdn.example.com/mains/mina-quiet-rooftop.mp4",
    },
    priceJpy: 1800,
  },
] as const satisfies readonly MainDetail[];

/**
 * mock main 一覧を取得する。
 */
export function listMains(): readonly MainDetail[] {
  return mains;
}

/**
 * main ID から詳細情報を取得する。
 */
export function getMainById(id: MainId): MainDetail | undefined {
  return mains.find((main) => main.id === id);
}

/**
 * static route 用の main ID 一覧を取得する。
 */
export function getMainIds(): readonly MainId[] {
  return mains.map((main) => main.id);
}
