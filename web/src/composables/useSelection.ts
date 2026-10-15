import { ref, computed, type Ref } from 'vue'

export interface SelectionItem {
  id: string
  type?: string
  name?: string
  [key: string]: any
}

export function useSelection<T extends SelectionItem>(items: Ref<T[]>) {
  const selectedIdSet = ref<Set<string>>(new Set())

  // 当前已选 ID 数组
  const selectedIds = computed<string[]>({
    get: () => Array.from(selectedIdSet.value),
    set: (ids: string[]) => {
      selectedIdSet.value = new Set(ids)
    }
  })

  // 已选中的条目总数
  const selectedCount = computed(() => selectedIdSet.value.size)

  // 当前页中被选中的条目
  const selectedItems = computed<T[]>(() => {
    return items.value.filter(item => selectedIdSet.value.has(item.id))
  })

  // 当前页是否全选
  const isAllSelected = computed(() => {
    if (items.value.length === 0) return false
    return items.value.every(item => selectedIdSet.value.has(item.id))
  })

  // 当前页是否为半选状态 (部分选中)
  const isIndeterminate = computed(() => {
    if (items.value.length === 0) return false
    const countInCurrentPage = items.value.filter(item => selectedIdSet.value.has(item.id)).length
    return countInCurrentPage > 0 && countInCurrentPage < items.value.length
  })

  // 判断单个 ID 是否被选中
  function isSelected(id: string): boolean {
    return selectedIdSet.value.has(id)
  }

  // 切换单个项的选中状态
  function toggle(id: string) {
    const next = new Set(selectedIdSet.value)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    selectedIdSet.value = next
  }

  // 设置单个项的选中状态
  function setSelected(id: string, selected: boolean) {
    const next = new Set(selectedIdSet.value)
    if (selected) {
      next.add(id)
    } else {
      next.delete(id)
    }
    selectedIdSet.value = next
  }

  // 切换当前页全选 / 取消全选
  function toggleAll() {
    const next = new Set(selectedIdSet.value)
    if (isAllSelected.value) {
      items.value.forEach(item => next.delete(item.id))
    } else {
      items.value.forEach(item => next.add(item.id))
    }
    selectedIdSet.value = next
  }

  // 清空所有选中
  function clear() {
    selectedIdSet.value = new Set()
  }

  // 选中指定的所有 ID
  function selectMany(ids: string[]) {
    const next = new Set(selectedIdSet.value)
    ids.forEach(id => next.add(id))
    selectedIdSet.value = next
  }

  // 取消选中的多个 ID
  function unselectMany(ids: string[]) {
    const next = new Set(selectedIdSet.value)
    ids.forEach(id => next.delete(id))
    selectedIdSet.value = next
  }

  // 选中的实体类型结构分析
  const selectionTypeMeta = computed(() => {
    const typeCounts: Record<string, number> = {}
    selectedItems.value.forEach(item => {
      const t = item.type || 'task'
      typeCounts[t] = (typeCounts[t] || 0) + 1
    })
    const types = Object.keys(typeCounts)
    const isSingleType = types.length === 1
    const primaryType = types.length > 0 ? types[0] : null
    const isMixed = types.length > 1

    return {
      typeCounts,
      types,
      isSingleType,
      primaryType,
      isMixed
    }
  })

  return {
    selectedIdSet,
    selectedIds,
    selectedCount,
    selectedItems,
    isAllSelected,
    isIndeterminate,
    selectionTypeMeta,
    isSelected,
    toggle,
    setSelected,
    toggleAll,
    clear,
    selectMany,
    unselectMany
  }
}
