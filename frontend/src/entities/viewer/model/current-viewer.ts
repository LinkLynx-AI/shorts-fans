export const viewerActiveModes = ["fan", "creator"] as const;

export type ViewerActiveMode = (typeof viewerActiveModes)[number];

export const viewerSessionCookieName = "shorts_fans_session";

export type CurrentViewer = {
  activeMode: ViewerActiveMode;
  canAccessCreatorMode: boolean;
  id: string;
};
