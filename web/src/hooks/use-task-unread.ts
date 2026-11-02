"use client";

import * as React from "react";

import { api } from "@/lib/api";
import {
  readUnreadCursor,
  saveUnreadBaseline,
  saveUnreadCursor,
  TASK_UNREAD_CHANGED,
  type UnreadCursor,
} from "@/lib/task-notifications";

const POLL_MS = 30_000;

/**
 * 任务列表的未读漏洞角标。
 *
 * - 每次轮询把每个任务各自的快照游标带给后端，后端只数“不在快照里”的新漏洞；
 * - 第一次访问（没有游标）时后端返回基线快照，我们把它存成全局基线，因此上线后
 *   历史漏洞不会一夜之间全变红点；
 * - 打开任务时调用 markRead 推进该任务的游标，角标立即消失；
 * - 请求失败保留上一次计数，不伪造 0；页面不可见时不轮询。
 */
export function useTaskUnread(taskIDs: string[]) {
  const ids = React.useMemo(() => [...new Set(taskIDs)].filter(Boolean).sort(), [taskIDs]);
  const scope = ids.join(",");
  const [counts, setCounts] = React.useState<Record<string, number>>({});
  const latest = React.useRef<UnreadCursor | null>(null);
  const inflight = React.useRef<AbortController | null>(null);

  const refresh = React.useCallback(async () => {
    if (typeof document !== "undefined" && document.visibilityState !== "visible") return;
    const list = scope ? scope.split(",") : [];
    if (!list.length) {
      setCounts({});
      return;
    }
    inflight.current?.abort();
    const controller = new AbortController();
    inflight.current = controller;
    try {
      const queries = list.map((id) => ({
        task_id: id,
        findings: readUnreadCursor(id)?.snapshot ?? "",
      }));
      const result = await api.taskNotifications(queries);
      if (controller.signal.aborted) return;
      latest.current = { snapshot: result.snapshot, observed_at: result.observed_at };
      // 首次访问没有任何游标：把当前快照存成基线，历史漏洞不算未读。
      if (list.every((id) => !readUnreadCursor(id))) saveUnreadBaseline(latest.current);
      const next: Record<string, number> = {};
      for (const item of result.items) {
        if (item.findings > 0) next[item.task_id] = item.findings;
      }
      setCounts(next);
    } catch {
      /* 保留上一次的计数：网络抖动不该把角标清零。 */
    }
  }, [scope]);

  React.useEffect(() => {
    void refresh();
  }, [refresh]);

  React.useEffect(() => {
    const timer = window.setInterval(() => void refresh(), POLL_MS);
    const onVisible = () => {
      if (document.visibilityState === "visible") void refresh();
    };
    window.addEventListener(TASK_UNREAD_CHANGED, refresh);
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      window.clearInterval(timer);
      window.removeEventListener(TASK_UNREAD_CHANGED, refresh);
      document.removeEventListener("visibilitychange", onVisible);
      inflight.current?.abort();
    };
  }, [refresh]);

  const markRead = React.useCallback((taskID: string) => {
    const cursor = latest.current;
    if (cursor) saveUnreadCursor(taskID, cursor);
    setCounts((prev) => {
      if (!(taskID in prev)) return prev;
      const next = { ...prev };
      delete next[taskID];
      return next;
    });
  }, []);

  return { counts, markRead };
}
