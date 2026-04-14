"use client";

import {
  FollowingCreatorList,
  type FollowingCreatorListProps,
} from "@/features/following-creator-list";
import { useFanAuthDialog } from "@/features/fan-auth";

type FollowingShellProps = Omit<FollowingCreatorListProps, "onAuthRequired">;

/**
 * following 詳細画面を表示する。
 */
export function FollowingShell({
  layout = "standalone",
  items,
  updateFollowingCreatorRelation,
}: FollowingShellProps) {
  const { openFanAuthDialog } = useFanAuthDialog();

  return (
    <FollowingCreatorList
      items={items}
      layout={layout}
      onAuthRequired={openFanAuthDialog}
      updateFollowingCreatorRelation={updateFollowingCreatorRelation}
    />
  );
}
