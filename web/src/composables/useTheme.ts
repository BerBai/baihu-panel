import { computed, onMounted, ref, watch } from 'vue'

export type Theme = 'light' | 'dark' | 'system'

function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'system'
  try {
    const saved = localStorage.getItem('theme') as Theme | null
    if (saved && ['light', 'dark', 'system'].includes(saved)) {
      return saved
    }
  } catch {}
  return 'system'
}

const theme = ref<Theme>(getInitialTheme())
const systemTheme = ref<'light' | 'dark'>(getSystemTheme())

function applyTheme(t: Theme) {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  const effectiveTheme = t === 'system' ? getSystemTheme() : t

  root.classList.remove('light', 'dark')
  root.classList.add(effectiveTheme)
}

// 模块加载即时同步一次，防组件挂载前状态脱节
applyTheme(theme.value)

let mediaListenerRegistered = false

export function useTheme() {
  onMounted(() => {
    systemTheme.value = getSystemTheme()

    if (!mediaListenerRegistered && typeof window !== 'undefined') {
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        systemTheme.value = getSystemTheme()
        if (theme.value === 'system') {
          applyTheme('system')
        }
      })
      mediaListenerRegistered = true
    }
  })

  watch(theme, (newTheme) => {
    localStorage.setItem('theme', newTheme)
    applyTheme(newTheme)
  })

  const resolvedTheme = computed<'light' | 'dark'>(() => {
    return theme.value === 'system' ? systemTheme.value : theme.value
  })

  function setTheme(t: Theme) {
    theme.value = t
  }

  return {
    theme,
    resolvedTheme,
    setTheme,
  }
}
