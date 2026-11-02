// 任务未读角标的前端游标存储。
//
// 服务端返回的 snapshot 是 PostgreSQL 的可见性快照（xmin:xmax:xip），不是时间戳：
// 并发提交时时间戳可能相同或乱序，事务快照一定严格有序。浏览器只负责记住“读到哪
// 一刻”，服务端据此只数“不在这张快照里”的新漏洞。
//
// 存两份：每个任务一份（打开任务时推进），一份全局基线（第一次访问时写入，避免首
// 次进入就把所有历史漏洞标成未读）。

export type UnreadCursor = { snapshot: string; observed_at: number };

const STORAGE_PREFIX = "artex:task-unread:";
const BASELINE_KEY = `${STORAGE_PREFIX}baseline`;
export const TASK_UNREAD_CHANGED = "artex:task-unread-changed";

export function taskUnreadKey(taskID: string): string {
  return STORAGE_PREFIX + taskID;
}

function parse(raw: string | null): UnreadCursor | null {
  if (!raw) return null;
  try {
    const v = JSON.parse(raw) as Partial<UnreadCursor>;
    if (typeof v?.snapshot !== "string" || typeof v?.observed_at !== "number") return null;
    return { snapshot: v.snapshot, observed_at: v.observed_at };
  } catch {
    return null;
  }
}

function read(storageKey: string): UnreadCursor | null {
  try {
    return parse(localStorage.getItem(storageKey));
  } catch {
    return null;
  }
}

function newer(a: UnreadCursor | null, b: UnreadCursor): UnreadCursor {
  if (a && a.observed_at >= b.observed_at) return a;
  return b;
}

function write(storageKey: string, cursor: UnreadCursor) {
  try {
    localStorage.setItem(storageKey, JSON.stringify(cursor));
  } catch {
    /* 隐私模式下 localStorage 不可写：本次会话内仍然按内存值工作。 */
  }
}

/** 每个任务各自的进度；没有就回落到全局基线（首次访问 = 0 条未读）。 */
export function readUnreadCursor(taskID: string): UnreadCursor | null {
  return read(taskUnreadKey(taskID)) ?? read(BASELINE_KEY);
}

export function saveUnreadCursor(taskID: string, incoming: UnreadCursor) {
  const current = read(taskUnreadKey(taskID));
  const next = newer(current, incoming);
  if (current && next === current) return;
  write(taskUnreadKey(taskID), next);
  window.dispatchEvent(new Event(TASK_UNREAD_CHANGED));
}

export function saveUnreadBaseline(incoming: UnreadCursor) {
  const current = read(BASELINE_KEY);
  // 基线只向前走；多个标签页并发首次访问时取更早的那个，避免漏报。
  if (current && current.observed_at <= incoming.observed_at) return;
  write(BASELINE_KEY, incoming);
}
