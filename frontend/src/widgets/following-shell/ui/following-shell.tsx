import {
  FollowingCreatorList,
  type FollowingCreatorListProps,
} from "@/features/following-creator-list";
import { useFanAuthDialog } from "@/features/fan-auth";

/**
 * following 詳細画面を表示する。
 */
export function FollowingShell({
  layout = "standalone",
  items,
  updateFollowingCreatorRelation,
}: FollowingCreatorListProps) {
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
