<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Wrench, Trash2, Zap, ZapOff, X } from 'lucide-vue-next'

const props = defineProps<{
  selectedCount: number
  selectedTypes?: string[]
  isMixed?: boolean
  loading?: boolean
  isAllSelected?: boolean
  isIndeterminate?: boolean
  totalInPage?: number
}>()

const emit = defineEmits<{
  'openBatchSettings': []
  'batchToggle': [enabled: boolean]
  'confirmBatchDelete': []
  'clear': []
  'toggleAll': []
}>()

const typeDescription = computed(() => {
  if (!props.selectedTypes || props.selectedTypes.length === 0) return ''
  if (!props.isMixed) {
    const t = props.selectedTypes[0]
    if (t === 'repo') return '仓库同步任务'
    if (t === 'app') return '白虎应用实体'
    if (t === 'agent_sync_script') return 'Agent 脚本同步'
    return '普通脚本任务'
  }
  return '多类型混合'
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
      v-if="selectedCount > 0"
      class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-1.5 sm:gap-2.5 px-3 sm:px-4 py-2 sm:py-2.5 rounded-2xl bg-card/95 backdrop-blur-xl border border-primary/20 shadow-2xl shadow-primary/10 max-w-[calc(100vw-24px)] overflow-x-auto text-xs"
    >
      <!-- 计数徽章 -->
      <div class="flex items-center gap-1.5 shrink-0 pl-0.5 pr-1 border-r border-border/60">
        <span class="inline-flex items-center justify-center px-2 py-0.5 rounded-full text-xs font-bold bg-primary text-primary-foreground shadow-xs">
          {{ selectedCount }}
        </span>
        <div class="hidden sm:flex flex-col">
          <span class="text-[11px] font-semibold text-foreground leading-none">项已选择</span>
          <span v-if="typeDescription" class="text-[10px] text-muted-foreground leading-tight mt-0.5 truncate max-w-[100px]" :title="typeDescription">
            {{ typeDescription }}
          </span>
        </div>
      </div>

      <!-- 全选/取消全选本页 (小屏触屏黄金热区) -->
      <button
        type="button"
        class="flex items-center gap-1.5 px-2 py-1 rounded-lg text-xs font-medium transition-colors hover:bg-muted/80 active:scale-95 text-foreground shrink-0 border border-border/40 select-none"
        title="全选/取消全选当前页"
        @click="$emit('toggleAll')"
      >
        <Checkbox
          :checked="isAllSelected ? true : (isIndeterminate ? 'indeterminate' : false)"
          class="pointer-events-none w-3.5 h-3.5"
        />
        <span class="text-[11px] font-medium hidden xs:inline sm:inline">全选本页</span>
      </button>

      <!-- 批量修改参数 (核心统一入口) -->
      <Button
        variant="default"
        size="sm"
        class="h-8 px-2.5 sm:px-3 text-xs font-semibold gap-1.5 bg-primary text-primary-foreground hover:bg-primary/90 shadow-xs shrink-0"
        :disabled="loading"
        @click="$emit('openBatchSettings')"
      >
        <Wrench class="h-3.5 w-3.5" />
        <span>批量设置参数</span>
      </Button>

      <!-- 快捷批量启用 -->
      <Button
        variant="outline"
        size="sm"
        class="h-8 px-2 sm:px-2.5 text-xs font-medium gap-1 hover:bg-emerald-500/10 hover:text-emerald-500 hover:border-emerald-500/30 shrink-0"
        :disabled="loading"
        title="批量启用选中的任务"
        @click="$emit('batchToggle', true)"
      >
        <Zap class="h-3.5 w-3.5 text-emerald-500 fill-emerald-500" />
        <span class="hidden md:inline">启用</span>
      </Button>

      <!-- 快捷批量停用 -->
      <Button
        variant="outline"
        size="sm"
        class="h-8 px-2 sm:px-2.5 text-xs font-medium gap-1 hover:bg-amber-500/10 hover:text-amber-500 hover:border-amber-500/30 shrink-0"
        :disabled="loading"
        title="批量禁用选中的任务"
        @click="$emit('batchToggle', false)"
      >
        <ZapOff class="h-3.5 w-3.5 text-muted-foreground" />
        <span class="hidden md:inline">禁用</span>
      </Button>

      <!-- 批量删除按钮 -->
      <Button
        variant="destructive"
        size="sm"
        class="h-8 px-2.5 sm:px-3 text-xs font-semibold gap-1.5 shadow-xs shrink-0"
        :disabled="loading"
        @click="$emit('confirmBatchDelete')"
      >
        <Trash2 class="h-3.5 w-3.5" />
        <span>删除选中</span>
      </Button>

      <!-- 清空选择 -->
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7 rounded-full text-muted-foreground hover:text-foreground shrink-0 ml-1"
        title="取消全选 (ESC)"
        @click="$emit('clear')"
      >
        <X class="h-4 w-4" />
      </Button>
    </div>
  </Transition>
</template>
