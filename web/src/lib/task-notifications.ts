// 移植自 SERIES `web/src/lib/task-notifications.ts`（notificationKey /
// parseNotificationCursor / newerNotificationCursor / notificationLabel 逐函数对齐）。
//
// 唯一改动：类别只保留 findings（SERIES 还有 assets / intercepts，两者依赖资产审批
// 体系，ARTEX 无宿主），因此去掉 MOCK 分支与两个类别常量；存储前缀沿用已上线的
// artex:task-unread:，不丢已有基线。

export const notificationCategories = ["findings"] as const;
export type NotificationCategory = (typeof notificationCategories)[number];
export type NotificationCursor = { snapshot: string; observed_at: number };
export type NotificationQuery = { task_id: string } & Record<NotificationCategory, string>;
export type NotificationCounts = { task_id: string } & Record<NotificationCategory, number>;
export type NotificationSummary = NotificationCursor & { items: NotificationCounts[] };
export const notificationStoragePrefix = "artex:task-unread:";

export function notificationKey(task: string, category: NotificationCategory) {
  return `${notificationStoragePrefix}${encodeURIComponent(task)}:${category}`;
}
export function parseNotificationCursor(raw: string | null): NotificationCursor | undefined {
  try {
    const value = JSON.parse(raw ?? "null");
    if (
      value &&
      typeof value.snapshot === "string" &&
      value.snapshot.length > 0 &&
      Number.isSafeInteger(value.observed_at) &&
      value.observed_at > 0
    )
      return value;
  } catch {
    /* Corrupt/old browser storage establishes a fresh baseline. */
  }
}
export function newerNotificationCursor(a: NotificationCursor | undefined, b: NotificationCursor) {
  return !a || b.observed_at > a.observed_at ? b : a;
}
export function notificationLabel(count: number) {
  return count > 99 ? "99+" : String(count);
}
