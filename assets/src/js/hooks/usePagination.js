import { useCallback, useEffect, useRef } from "react";

export const PAGE_SIZE = 50;

export function usePagination({ list, total, fetchPage }) {
  const hasMore = list.length < total;
  const busyRef = useRef(false);
  const sentinelRef = useRef(null);

  const loadMore = useCallback(async () => {
    if (busyRef.current || !hasMore) return;
    busyRef.current = true;
    try {
      await fetchPage(list.length);
    } finally {
      busyRef.current = false;
    }
  }, [hasMore, list.length, fetchPage]);

  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !hasMore) return;
    const obs = new IntersectionObserver(
      (entries) => { if (entries[0].isIntersecting) loadMore(); },
      { rootMargin: "120px" }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [hasMore, loadMore]);

  return { hasMore, sentinelRef };
}