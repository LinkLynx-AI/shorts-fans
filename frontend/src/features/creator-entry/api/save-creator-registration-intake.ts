import { requestJson } from "@/shared/api";

import {
  creatorRegistrationIntakeResponseSchema,
  type CreatorRegistrationIntake,
} from "./contracts";

type SaveCreatorRegistrationIntakeInput = {
  acceptsConsentResponsibility: boolean;
  birthDate: string;
  creatorBio: string;
  declaresNoProhibitedCategory: boolean;
  legalName: string;
  payoutRecipientName: string;
  payoutRecipientType: string;
};

type SaveCreatorRegistrationIntakeOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
};

/**
 * creator registration intake draft を保存する。
 */
export async function saveCreatorRegistrationIntake(
  input: SaveCreatorRegistrationIntakeInput,
  {
    baseUrl,
    credentials = "include",
    fetcher,
  }: SaveCreatorRegistrationIntakeOptions = {},
): Promise<CreatorRegistrationIntake> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      body: JSON.stringify(input),
      credentials,
      headers: {
        "Content-Type": "application/json",
      },
      method: "PUT",
    },
    path: "/api/viewer/creator-registration/intake",
    schema: creatorRegistrationIntakeResponseSchema,
  });

  return response.data.intake;
}
