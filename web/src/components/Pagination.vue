<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const props = withDefaults(defineProps<{
  total: number
  page: number
  pageSize?: number
  pageSizeOptions?: number[]
  showSizeChanger?: boolean
}>(), {
  pageSize: 10,
  pageSizeOptions: () => [10, 20, 50, 100],
  showSizeChanger: true
})

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [pageSize: number]
}>()

// 当前生效的 pageSize
const currentPageSize = computed(() => props.pageSize || 10)

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
  <div class="flex items-center justify-between px-3 py-1.5 border-t border-border/40 text-xs text-muted-foreground select-none min-h-[34px]">
    <!-- 左侧信息与条数选择器 -->
    <div class="flex items-center gap-2">
      <span class="tabular-nums whitespace-nowrap text-[11px] text-muted-foreground font-['Inter',sans-serif]">
        共 <strong class="text-foreground font-semibold font-['Inter',sans-serif]">{{ total }}</strong> 条
        <span class="hidden sm:inline text-muted-foreground/60 ml-0.5 text-[10px]">({{ currentRange }})</span>
      </span>

      <!-- 快捷每页条数切换 -->
      <div v-if="showSizeChanger" class="flex items-center">
        <Select :model-value="String(currentPageSize)" @update:model-value="handleSizeChange">
          <SelectTrigger class="!h-[22px] !min-h-0 !py-0 !px-1.5 !text-[11px] !gap-1 bg-muted/20 hover:bg-muted/50 border border-border/40 hover:border-border/70 rounded shadow-none text-muted-foreground hover:text-foreground transition-colors [&_svg]:!size-3 [&_svg]:!opacity-40 hover:[&_svg]:!opacity-80 font-['Inter',sans-serif]">
            <SelectValue :placeholder="`${currentPageSize} 条/页`" />
          </SelectTrigger>
          <SelectContent align="start" class="min-w-[78px] p-0.5 text-xs shadow-md border-border/50">
            <SelectItem
              v-for="opt in pageSizeOptions"
              :key="opt"
              :value="String(opt)"
              class="text-[11px] py-1 pl-2 pr-5 cursor-pointer rounded-xs font-['Inter',sans-serif]"
            >
              {{ opt }} 条/页
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- 右侧翻页操作区 -->
    <div class="flex items-center gap-0.5 sm:gap-1">
      <Button
        variant="ghost"
        size="icon"
        class="h-[22px] w-[22px] rounded hover:bg-muted text-muted-foreground hover:text-foreground disabled:opacity-25"
        :disabled="page <= 1"
        title="上一页"
        @click="prevPage"
      >
        <ChevronLeft class="h-3 w-3" />
      </Button>
      <span class="px-1 text-[11px] tabular-nums font-['Inter',sans-serif] font-medium text-foreground tracking-tight">
        {{ page }} <span class="text-muted-foreground/50 font-normal">/ {{ totalPages }}</span>
      </span>
      <Button
        variant="ghost"
        size="icon"
        class="h-[22px] w-[22px] rounded hover:bg-muted text-muted-foreground hover:text-foreground disabled:opacity-25"
        :disabled="page >= totalPages"
        title="下一页"
        @click="nextPage"
      >
        <ChevronRight class="h-3 w-3" />
      </Button>
    </div>
  </div>
</template>
