<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useSiteSettings } from '@/composables/useSiteSettings'

const props = withDefaults(defineProps<{
  total: number
  page: number
  pageSize?: number
  pageSizeOptions?: number[]
  showSizeChanger?: boolean
}>(), {
  pageSize: undefined,
  pageSizeOptions: () => [10, 20, 50, 100],
  showSizeChanger: true
})

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [pageSize: number]
}>()

const { pageSize: globalPageSize } = useSiteSettings()

// 当前生效的 pageSize (优先使用组件传入的 prop，兜底使用全局设置)
const currentPageSize = computed(() => props.pageSize || globalPageSize.value || 10)

const totalPages = computed(() => Math.ceil(props.total / currentPageSize.value) || 1)

// 计算当前展示区间，如 "1-20"
const currentRange = computed(() => {
  if (props.total === 0) return '0'
  const start = (props.page - 1) * currentPageSize.value + 1
  const end = Math.min(props.page * currentPageSize.value, props.total)
  return `${start}-${end}`
})

function prevPage() {
  if (props.page > 1) {
    emit('update:page', props.page - 1)
  }
}

function nextPage() {
  if (props.page < totalPages.value) {
    emit('update:page', props.page + 1)
  }
}

function handleSizeChange(val: any) {
  const newSize = parseInt(String(val), 10)
  if (!newSize || isNaN(newSize)) return
  emit('update:pageSize', newSize)
  // 当改变每页条数导致当前页超出总页数时，自动切回第一页
  const newTotalPages = Math.ceil(props.total / newSize) || 1
  if (props.page > newTotalPages) {
    emit('update:page', 1)
  }
}
</script>

<template>
  <div class="flex items-center justify-between px-3 py-2 border-t text-xs text-muted-foreground select-none">
    <!-- 左侧信息与条数选择器 -->
    <div class="flex items-center gap-2 sm:gap-3">
      <span class="tabular-nums whitespace-nowrap">
        共 <strong class="text-foreground font-medium">{{ total }}</strong> 条
        <span class="hidden sm:inline opacity-75">({{ currentRange }})</span>
      </span>

      <!-- 快捷每页条数切换 -->
      <div v-if="showSizeChanger" class="flex items-center gap-1.5">
        <Select :model-value="String(currentPageSize)" @update:model-value="handleSizeChange">
          <SelectTrigger class="h-6 px-2 py-0 text-[11px] bg-muted/20 border-muted-foreground/15 rounded gap-1 min-w-[76px] justify-between">
            <SelectValue :placeholder="`${currentPageSize} 条/页`" />
          </SelectTrigger>
          <SelectContent align="start" class="min-w-[84px] p-1 text-xs">
            <SelectItem
              v-for="opt in pageSizeOptions"
              :key="opt"
              :value="String(opt)"
              class="text-[11px] py-1 pl-2 pr-6 cursor-pointer"
            >
              {{ opt }} 条/页
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- 右侧翻页操作区 -->
    <div class="flex items-center gap-1 sm:gap-1.5">
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 hover:bg-muted"
        :disabled="page <= 1"
        title="上一页"
        @click="prevPage"
      >
        <ChevronLeft class="h-3 w-3" />
      </Button>
      <span class="px-1.5 text-[11px] tabular-nums font-medium text-foreground">
        {{ page }} <span class="text-muted-foreground font-normal">/ {{ totalPages }}</span>
      </span>
      <Button
        variant="ghost"
        size="icon"
        class="h-6 w-6 hover:bg-muted"
        :disabled="page >= totalPages"
        title="下一页"
        @click="nextPage"
      >
        <ChevronRight class="h-3 w-3" />
      </Button>
    </div>
  </div>
</template>
