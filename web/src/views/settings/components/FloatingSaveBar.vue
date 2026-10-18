<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Save, RotateCcw, X, Loader2, ExternalLink } from 'lucide-vue-next'
import { useSettingsTracker } from '../composables/useSettingsTracker'

const tracker = useSettingsTracker()

const isVisible = computed(() => {
  return tracker && tracker.dirtySections.value.length > 0
})

const currentSection = computed(() => {
  return tracker?.currentDirtySection.value || null
})

const totalCount = computed(() => {
  return tracker?.totalDirtyCount.value || 0
})

const isSaving = computed(() => {
  return tracker?.isSaving.value || false
})

const dirtySections = computed(() => {
  return tracker?.dirtySections.value || []
})

// 检查当前变动模块是否在当前激活的 Tab
const isCurrentTab = computed(() => {
  if (!tracker || !currentSection.value) return true
  return tracker.activeTab.value === currentSection.value.tab
})

function switchTabTo(tab: string) {
  if (tracker) {
    tracker.activeTab.value = tab
  }
}

async function handleSave() {
  if (tracker && !isSaving.value) {
    await tracker.saveCurrent()
  }
}

function handleReset() {
  if (tracker && !isSaving.value) {
    tracker.resetCurrent()
  }
}

// 监听键盘 Ctrl+S / Cmd+S 快捷保存
function handleKeyDown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    if (isVisible.value && !isSaving.value) {
      e.preventDefault()
      handleSave()
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="transform translate-y-8 opacity-0 scale-95"
    enter-to-class="transform translate-y-0 opacity-100 scale-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="transform translate-y-0 opacity-100 scale-100"
    leave-to-class="transform translate-y-8 opacity-0 scale-95"
  >
    <div
      v-if="isVisible && currentSection"
      class="fixed bottom-4 sm:bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-1.5 sm:gap-2.5 px-2.5 sm:px-4 py-2 sm:py-2.5 rounded-2xl bg-card/95 backdrop-blur-xl border border-primary/25 shadow-2xl shadow-primary/10 max-w-[calc(100vw-16px)] sm:max-w-fit text-xs select-none hide-scrollbar overflow-x-auto ring-1 ring-border/50"
    >
      <!-- 变动计数与模块指示区 -->
      <div class="flex items-center gap-2 shrink-0 pl-0.5 pr-1.5 border-r border-border/60">
        <span
          class="inline-flex items-center justify-center min-w-5 h-5 px-1.5 rounded-full text-[11px] sm:text-xs font-bold bg-primary text-primary-foreground shadow-xs animate-pulse shrink-0"
        >
          {{ totalCount }}
        </span>
        <div class="flex flex-col min-w-0">
          <div class="flex items-center gap-1.5">
            <span class="text-[11px] font-semibold text-foreground leading-none whitespace-nowrap">
              配置已修改
            </span>
            <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse shrink-0"></span>
          </div>
          <span
            class="text-[10px] text-muted-foreground leading-tight mt-0.5 truncate max-w-[130px] sm:max-w-[180px]"
            :title="currentSection.title"
          >
            {{ currentSection.title }}
            <span v-if="dirtySections.length > 1" class="text-primary font-medium">
              (+{{ dirtySections.length - 1 }})
            </span>
          </span>
        </div>
      </div>

      <!-- 若变动模块在其他 Tab，提供快速切回提示按钮 -->
      <button
        v-if="!isCurrentTab"
        type="button"
        @click="switchTabTo(currentSection.tab)"
        class="flex items-center gap-1 px-2 py-1 rounded-lg text-xs font-medium bg-amber-500/10 text-amber-600 dark:text-amber-400 hover:bg-amber-500/20 active:scale-95 shrink-0 border border-amber-500/20 transition-colors"
        title="点击跳转查看变动配置所在页面"
      >
        <span class="text-[11px]">查看变动位置</span>
        <ExternalLink class="w-3 h-3" />
      </button>

      <!-- 快捷快捷键提示 (PC端) -->
      <div class="hidden lg:flex items-center gap-1 px-1.5 py-1 rounded-md bg-muted/60 text-[10px] text-muted-foreground font-mono shrink-0 select-none border border-border/40">
        <span>Ctrl+S</span>
      </div>

      <!-- 放弃本次修改 / 还原按钮 -->
      <Button
        variant="outline"
        size="sm"
        class="h-7.5 sm:h-8 px-2 sm:px-3 text-xs font-medium gap-1 sm:gap-1.5 hover:bg-muted/80 text-foreground shrink-0 border-border/60"
        :disabled="isSaving"
        title="放弃当前未保存的修改并还原初始状态"
        @click="handleReset"
      >
        <RotateCcw class="h-3.5 w-3.5 text-muted-foreground" />
        <span class="hidden sm:inline">放弃修改</span>
        <span class="sm:hidden">还原</span>
      </Button>

      <!-- 核心保存按钮 (主色凸显) -->
      <Button
        variant="default"
        size="sm"
        class="h-7.5 sm:h-8 px-3 sm:px-4 text-xs font-semibold gap-1.5 bg-primary text-primary-foreground hover:bg-primary/90 shadow-xs shrink-0"
        :disabled="isSaving"
        @click="handleSave"
      >
        <Loader2 v-if="isSaving" class="h-3.5 w-3.5 animate-spin" />
        <Save v-else class="h-3.5 w-3.5" />
        <span>{{ isSaving ? '保存中...' : '保存配置' }}</span>
      </Button>

      <!-- 关闭/放弃快捷图标 -->
      <Button
        variant="ghost"
        size="icon"
        class="h-6.5 w-6.5 sm:h-7 sm:w-7 rounded-full text-muted-foreground hover:text-foreground shrink-0 ml-0.5"
        title="放弃修改"
        :disabled="isSaving"
        @click="handleReset"
      >
        <X class="h-3.5 w-3.5 sm:h-4 sm:w-4" />
      </Button>
    </div>
  </Transition>
</template>
