import { ref, watch, type Ref } from 'vue'

/**
 * 每个表格组件/页面独立的每页条数浏览器持久化缓存
 * @param storageKey 本地存储使用的 key（如 'bhp_tasks_page_size'）
 * @param defaultSize 默认每页条数，默认为 10
 */
export function usePageSize(storageKey: string, defaultSize: number = 10): {
  pageSize: Ref<number>
  handlePageSizeChange: (newSize: number, onReload?: () => void) => void
} {
  const readSize = (): number => {
    try {
      const val = localStorage.getItem(storageKey)
      if (val) {
        const parsed = parseInt(val, 10)
        if (!isNaN(parsed) && parsed > 0) {
          return parsed
        }
      }
    } catch { }
    return defaultSize
  }

  const pageSize = ref<number>(readSize())

  watch(pageSize, (newVal) => {
    try {
      localStorage.setItem(storageKey, String(newVal))
    } catch { }
  })

  function handlePageSizeChange(newSize: number, onReload?: () => void) {
    if (!newSize || isNaN(newSize) || newSize <= 0) return
    pageSize.value = newSize
    try {
      localStorage.setItem(storageKey, String(newSize))
    } catch { }
    if (onReload) {
      onReload()
    }
  }

  return {
    pageSize,
    handlePageSizeChange
  }
}
