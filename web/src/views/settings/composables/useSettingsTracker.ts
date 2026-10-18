import { shallowRef, computed, unref, provide, inject, onUnmounted, type Ref, type ComputedRef } from 'vue'

export interface SettingSection {
  id: string
  tab: string
  title: string
  isDirty: Ref<boolean> | ComputedRef<boolean>
  dirtyCount?: Ref<number> | ComputedRef<number>
  saving: Ref<boolean>
  save: () => Promise<void> | void
  reset: () => void
}

const SETTINGS_TRACKER_KEY = Symbol('SETTINGS_TRACKER')

export interface SettingsTrackerContext {
  registerSection: (section: SettingSection) => void
  unregisterSection: (id: string) => void
  dirtySections: ComputedRef<SettingSection[]>
  currentDirtySection: ComputedRef<SettingSection | null>
  totalDirtyCount: ComputedRef<number>
  isSaving: ComputedRef<boolean>
  activeTab: Ref<string>
  saveCurrent: () => Promise<void>
  resetCurrent: () => void
  saveAll: () => Promise<void>
  resetAll: () => void
}

export function provideSettingsTracker(activeTab: Ref<string>): SettingsTrackerContext {
  const sections = shallowRef<Map<string, SettingSection>>(new Map())

  function registerSection(section: SettingSection) {
    const next = new Map(sections.value)
    next.set(section.id, section)
    sections.value = next
  }

  function unregisterSection(id: string) {
    const next = new Map(sections.value)
    next.delete(id)
    sections.value = next
  }

  // 所有处于脏状态（有变动）的模块
  const dirtySections = computed<SettingSection[]>(() => {
    const list: SettingSection[] = []
    for (const sec of sections.value.values()) {
      if (unref(sec.isDirty)) {
        list.push(sec)
      }
    }
    return list
  })

  // 优先匹配当前活跃 Tab 的变动模块，若当前 Tab 没有则回退至第一个变动模块
  const currentDirtySection = computed<SettingSection | null>(() => {
    const list = dirtySections.value
    if (list.length === 0) return null
    const matched = list.find(s => s.tab === activeTab.value)
    return (matched ?? list[0]) ?? null
  })

  // 变动字段总计数
  const totalDirtyCount = computed<number>(() => {
    return dirtySections.value.reduce((acc, sec) => {
      const c = sec.dirtyCount ? unref(sec.dirtyCount) : (unref(sec.isDirty) ? 1 : 0)
      return acc + (c > 0 ? c : 1)
    }, 0)
  })

  // 是否正在保存中
  const isSaving = computed<boolean>(() => {
    return dirtySections.value.some(sec => unref(sec.saving))
  })

  async function saveCurrent() {
    const target = currentDirtySection.value
    if (target && !unref(target.saving)) {
      await target.save()
    }
  }

  function resetCurrent() {
    const target = currentDirtySection.value
    if (target) {
      target.reset()
    }
  }

  async function saveAll() {
    for (const sec of dirtySections.value) {
      if (!unref(sec.saving)) {
        await sec.save()
      }
    }
  }

  function resetAll() {
    for (const sec of dirtySections.value) {
      sec.reset()
    }
  }

  const context: SettingsTrackerContext = {
    registerSection,
    unregisterSection,
    dirtySections,
    currentDirtySection,
    totalDirtyCount,
    isSaving,
    activeTab,
    saveCurrent,
    resetCurrent,
    saveAll,
    resetAll
  }

  provide(SETTINGS_TRACKER_KEY, context)
  return context
}

export function useSettingsTracker(): SettingsTrackerContext | null {
  return inject(SETTINGS_TRACKER_KEY, null)
}

export function useRegisterSettingSection(section: SettingSection) {
  const tracker = useSettingsTracker()
  if (tracker) {
    tracker.registerSection(section)
    onUnmounted(() => {
      tracker.unregisterSection(section.id)
    })
  }
}

/**
 * 字段被修改时应用的高亮边框与光晕样式
 */
export const MODIFIED_INPUT_CLASS = 'border-amber-500/80 dark:border-amber-400/90 ring-2 ring-amber-500/15 dark:ring-amber-400/20 bg-amber-500/[0.03] transition-all'

/**
 * 设置项表单字段变动辅助工具
 */
export function useFieldDiff<T extends Record<string, any>>(
  form: Ref<T>,
  initialForm: Ref<T | null>
) {
  function isDiff(key: keyof T): boolean {
    if (!initialForm.value) return false
    return form.value[key] !== initialForm.value[key]
  }

  function getOriginal(key: keyof T): any {
    return initialForm.value ? initialForm.value[key] : undefined
  }

  function resetField(key: keyof T) {
    if (!initialForm.value) return
    form.value[key] = initialForm.value[key]
  }

  const changedFields = computed<Array<keyof T>>(() => {
    if (!initialForm.value) return []
    const diffs: Array<keyof T> = []
    for (const key of Object.keys(form.value) as Array<keyof T>) {
      if (form.value[key] !== initialForm.value[key]) {
        diffs.push(key)
      }
    }
    return diffs
  })

  return {
    isDiff,
    getOriginal,
    resetField,
    changedFields,
    diffClass: (key: keyof T) => (isDiff(key) ? MODIFIED_INPUT_CLASS : '')
  }
}

