<script setup lang="ts">
import { RotateCcw } from 'lucide-vue-next'

const props = defineProps<{
  originalValue?: any
  label?: string
  showReset?: boolean
}>()

const emit = defineEmits<{
  (e: 'reset'): void
}>()

function formatValue(val: any): string {
  if (val === undefined || val === null) return '空'
  if (typeof val === 'boolean') return val ? '开启' : '关闭'
  if (typeof val === 'string' && val.trim() === '') return '空'
  return String(val)
}
</script>

<template>
  <span
    class="inline-flex items-center gap-1 text-[10px] font-normal text-amber-600 dark:text-amber-400 bg-amber-500/10 px-1.5 py-0.5 rounded border border-amber-500/25 animate-in fade-in zoom-in-95 duration-150 select-none group/modbadge shrink-0"
    :title="originalValue !== undefined ? `已修改 (原值: ${formatValue(originalValue)})${showReset !== false ? ' · 点击重置' : ''}` : '已修改'"
  >
    <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse shrink-0"></span>
    <span class="font-medium whitespace-nowrap">已修改</span>
    <span
      v-if="originalValue !== undefined"
      class="text-[9px] text-muted-foreground/80 hidden sm:inline max-w-[100px] truncate"
    >
      (原: {{ formatValue(originalValue) }})
    </span>
    <button
      v-if="showReset !== false"
      type="button"
      @click.stop="emit('reset')"
      class="ml-0.5 p-0.5 rounded hover:bg-amber-500/20 text-amber-600 dark:text-amber-400 opacity-60 hover:opacity-100 transition-all shrink-0 cursor-pointer"
      title="撤销此项修改，还原为初始值"
    >
      <RotateCcw class="w-2.5 h-2.5" />
    </button>
  </span>
</template>
