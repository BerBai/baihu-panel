<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ListTodo, Variable, Clock, Play, ScrollText, Activity, ArrowUpRight, CheckCircle2, PieChart as PieIcon, TrendingUp } from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { api, type Stats, type DailyStats, type TaskStatsItem } from '@/api'
import ApexCharts from 'apexcharts'

const router = useRouter()
const stats = ref<Stats>({ tasks: 0, today_execs: 0, envs: 0, logs: 0, scheduled: 0, running: 0 })
const displayStats = ref<Stats>({ tasks: 0, today_execs: 0, envs: 0, logs: 0, scheduled: 0, running: 0 })
const sendStats = ref<DailyStats[]>([])
const taskStats = ref<TaskStatsItem[]>([])
const chartsLoaded = ref(false)
const isMobile = ref(window.innerWidth < 768)
const chartDays = computed(() => isMobile.value ? 15 : 30)

let lineChart: ApexCharts | null = null
let pieChart: ApexCharts | null = null
let themeObserver: MutationObserver | null = null
let themeChangeTimeout: number | null = null
let isInitializing = true

const statItems = [
  {
    key: 'today_execs',
    label: '今日执行',
    unit: '次',
    icon: Play,
    route: '/history?date=today',
    dotClass: 'bg-sky-500/40 dark:bg-sky-400/35',
    topBarClass: 'bg-gradient-to-r from-sky-500/30 via-sky-400/10 to-transparent',
    iconWrapperClass: 'bg-sky-500/3 text-sky-600/55 dark:text-sky-400/45 dark:bg-white/[0.025]',
    iconColor: 'text-sky-500'
  },
  {
    key: 'tasks',
    label: '任务总数',
    unit: '项',
    icon: ListTodo,
    route: '/tasks',
    dotClass: 'bg-zinc-400/50 dark:bg-zinc-400/40',
    topBarClass: 'bg-gradient-to-r from-zinc-400/30 via-zinc-400/10 to-transparent',
    iconWrapperClass: 'bg-zinc-500/3 text-zinc-500/70 dark:text-zinc-400/50 dark:bg-white/[0.025]',
    iconColor: 'text-zinc-400'
  },
  {
    key: 'envs',
    label: '环境变量',
    unit: '个',
    icon: Variable,
    route: '/environments',
    dotClass: 'bg-amber-500/40 dark:bg-amber-400/35',
    topBarClass: 'bg-gradient-to-r from-amber-500/30 via-amber-400/10 to-transparent',
    iconWrapperClass: 'bg-amber-500/3 text-amber-600/55 dark:text-amber-400/45 dark:bg-white/[0.025]',
    iconColor: 'text-amber-500'
  },
  {
    key: 'logs',
    label: '日志总数',
    unit: '条',
    icon: ScrollText,
    route: '/history',
    dotClass: 'bg-slate-400/50 dark:bg-slate-400/40',
    topBarClass: 'bg-gradient-to-r from-slate-400/30 via-slate-400/10 to-transparent',
    iconWrapperClass: 'bg-slate-500/3 text-slate-500/70 dark:text-slate-400/50 dark:bg-white/[0.025]',
    iconColor: 'text-slate-400'
  },
  {
    key: 'scheduled',
    label: '调度注册',
    unit: '项',
    icon: Clock,
    route: '/tasks?enabled=1&type=all',
    dotClass: 'bg-indigo-500/40 dark:bg-indigo-400/35',
    topBarClass: 'bg-gradient-to-r from-indigo-500/30 via-indigo-400/10 to-transparent',
    iconWrapperClass: 'bg-indigo-500/3 text-indigo-600/55 dark:text-indigo-400/45 dark:bg-white/[0.025]',
    iconColor: 'text-indigo-500'
  },
  {
    key: 'running',
    label: '正在运行',
    unit: '个',
    icon: Activity,
    route: '/history?status=running',
    dotClass: 'bg-emerald-500/40 dark:bg-emerald-400/35',
    topBarClass: 'bg-gradient-to-r from-emerald-500/30 via-emerald-400/10 to-transparent',
    iconWrapperClass: 'bg-emerald-500/3 text-emerald-600/55 dark:text-emerald-400/45 dark:bg-white/[0.025]',
    iconColor: 'text-emerald-500'
  },
]

const isDark = ref(document.documentElement.classList.contains('dark'))

// 计算近 N 天总执行数与平均成功率
const totalExecutions = computed(() => {
  return sendStats.value.reduce((sum, item) => sum + item.total, 0)
})

const overallSuccessRate = computed(() => {
  const total = totalExecutions.value
  if (!total) return '100.0'
  const success = sendStats.value.reduce((sum, item) => sum + item.success, 0)
  return ((success / total) * 100).toFixed(1)
})

// 数字滚动动画
function animateNumber(key: keyof Stats, target: number) {
  const start = displayStats.value[key]
  const duration = 800 // 动画持续时间
  const startTime = performance.now()

  function update(currentTime: number) {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)

    // 使用 easeOutCubic 缓动函数
    const easeProgress = 1 - Math.pow(1 - progress, 3)

    displayStats.value[key] = Math.round(start + (target - start) * easeProgress)

    if (progress < 1) {
      requestAnimationFrame(update)
    } else {
      displayStats.value[key] = target
    }
  }

  requestAnimationFrame(update)
}

// 更新统计数据并触发动画
function updateStats(newStats: Stats) {
  Object.keys(newStats).forEach((key) => {
    const statKey = key as keyof Stats
    if (newStats[statKey] !== stats.value[statKey]) {
      animateNumber(statKey, newStats[statKey])
    }
  })
  stats.value = newStats
}

function getTextColor() {
  return isDark.value ? '#94a3b8' : '#64748b'
}

function getGridColor() {
  return isDark.value ? '#334155' : '#e2e8f0'
}

function handleThemeChange() {
  // 初始化期间忽略主题变化
  if (isInitializing) return

  const newIsDark = document.documentElement.classList.contains('dark')
  if (newIsDark !== isDark.value) {
    isDark.value = newIsDark
    chartsLoaded.value = false
    if (themeChangeTimeout) clearTimeout(themeChangeTimeout)
    themeChangeTimeout = window.setTimeout(() => {
      reloadCharts()
    }, 50)
  }
}

function navigateTo(route?: string) {
  if (route) router.push(route)
}

function handleResize() {
  const wasMobile = isMobile.value
  isMobile.value = window.innerWidth < 768
  if (wasMobile !== isMobile.value) {
    reloadCharts()
  }
}

async function reloadCharts() {
  // 销毁旧图表
  if (lineChart) {
    lineChart.destroy()
    lineChart = null
  }
  if (pieChart) {
    pieChart.destroy()
    pieChart = null
  }

  chartsLoaded.value = false

  // 重新获取数据
  const [sendStatsData, taskStatsData] = await Promise.all([
    api.dashboard.sendStats(chartDays.value),
    api.dashboard.taskStats(chartDays.value)
  ])
  sendStats.value = sendStatsData
  taskStats.value = taskStatsData

  // 等待 Vue 更新 DOM
  await nextTick()
  await new Promise(resolve => setTimeout(resolve, 100))

  // 检查容器是否存在再渲染
  const statsChart = document.querySelector("#stats-chart")
  const pieChartEl = document.querySelector("#pie-chart")

  if (statsChart && pieChartEl && statsChart.parentElement && pieChartEl.parentElement) {
    renderLineChart()
    renderPieChart()
    await nextTick()
    chartsLoaded.value = true
  }
}

const renderLineChart = () => {
  const container = document.querySelector("#stats-chart")
  if (!container || !container.parentElement) return

  if (lineChart) {
    lineChart.destroy()
    lineChart = null
  }

  // 清空容器
  container.innerHTML = ''

  const textColor = getTextColor()
  const gridColor = getGridColor()
  const options = {
    series: [
      { name: '执行总数', data: sendStats.value.map(item => item.total) },
      { name: '执行成功', data: sendStats.value.map(item => item.success) },
      { name: '执行失败', data: sendStats.value.map(item => item.failed) }
    ],
    chart: {
      type: 'line' as const,
      height: 290,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: 800,
        animateGradually: { enabled: true, delay: 150 },
        dynamicAnimation: { enabled: true, speed: 350 }
      },
      dropShadow: {
        enabled: true,
        color: '#000',
        top: 18,
        left: 7,
        blur: 10,
        opacity: 0.2
      }
    },
    stroke: { curve: 'smooth' as const, width: 3, lineCap: 'round' as const },
    dataLabels: {
      enabled: false
    },
    markers: {
      size: 6,
      colors: ['#3b82f6', '#10b981', '#ef4444'],
      strokeColors: isDark.value ? '#1e293b' : '#fff',
      strokeWidth: 2,
      hover: { size: 8, sizeOffset: 3 }
    },
    xaxis: {
      categories: sendStats.value.map(item => item.day.slice(5)),
      axisBorder: { show: false },
      axisTicks: { show: false },
      labels: { style: { colors: textColor, fontSize: '11px', fontFamily: 'Inter, sans-serif' } },
      tooltip: { enabled: false }
    },
    yaxis: {
      labels: {
        style: { colors: textColor, fontSize: '12px', fontFamily: 'Inter, sans-serif' },
        formatter: (val: number) => String(val),
        offsetX: -15,
        minWidth: 20
      }
    },
    colors: ['#3b82f6', '#10b981', '#ef4444'],
    fill: {
      type: 'gradient',
      gradient: {
        shade: isDark.value ? 'dark' : 'light',
        type: 'vertical',
        shadeIntensity: 0.5,
        gradientToColors: ['#60a5fa', '#34d399', '#f87171'],
        inverseColors: false,
        opacityFrom: 0.8,
        opacityTo: 0.1,
        stops: [0, 100]
      }
    },
    grid: {
      borderColor: gridColor,
      strokeDashArray: 3,
      xaxis: { lines: { show: false } },
      yaxis: { lines: { show: true } },
      padding: { top: 5, right: 5, bottom: 0, left: 0 }
    },
    legend: {
      position: 'top' as const,
      horizontalAlign: 'right' as const,
      floating: true,
      offsetY: -25,
      offsetX: -5,
      fontSize: '12px',
      fontFamily: 'Inter, sans-serif',
      labels: { colors: textColor },
      markers: { width: 8, height: 8, radius: 4 }
    },
    tooltip: {
      enabled: true,
      shared: true,
      intersect: false,
      theme: isDark.value ? 'dark' : 'light',
      style: { fontSize: '12px', fontFamily: 'Inter, sans-serif' },
      x: { show: true },
      marker: { show: true },
      custom: ({ series, dataPointIndex, w }: { series: number[][], dataPointIndex: number, w: any }) => {
        const total = series[0]?.[dataPointIndex] ?? 0
        const success = series[1]?.[dataPointIndex] ?? 0
        const failed = series[2]?.[dataPointIndex] ?? 0
        const rate = total > 0 ? ((success / total) * 100).toFixed(1) : '0.0'
        return `<div class="bg-card text-foreground p-2 rounded-lg shadow-lg border border-border text-xs">
          <div class="font-medium mb-1.5">${w.globals.categoryLabels[dataPointIndex]}</div>
          <div class="space-y-0.5">
            <div class="flex items-center justify-between gap-3">
              <span class="flex items-center"><span class="w-1.5 h-1.5 bg-blue-500 rounded-full mr-1.5"></span><span class="text-muted-foreground">总数:</span></span>
              <span class="font-medium">${total} 次</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="flex items-center"><span class="w-1.5 h-1.5 bg-green-500 rounded-full mr-1.5"></span><span class="text-muted-foreground">成功:</span></span>
              <span class="font-medium">${success} 次</span>
            </div>
            <div class="flex items-center justify-between gap-3">
              <span class="flex items-center"><span class="w-1.5 h-1.5 bg-red-500 rounded-full mr-1.5"></span><span class="text-muted-foreground">失败:</span></span>
              <span class="font-medium">${failed} 次</span>
            </div>
            <div class="border-t border-border pt-1 mt-1">
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">成功率:</span>
                <span class="font-medium">${rate}%</span>
              </div>
            </div>
          </div>
        </div>`
      }
    },
    responsive: [{ breakpoint: 768, options: { chart: { height: 260 }, legend: { position: 'top', horizontalAlign: 'center', floating: false, offsetY: 0 } } }]
  }
  lineChart = new ApexCharts(container as HTMLElement, options as any)
  lineChart.render()
}

const renderPieChart = () => {
  if (taskStats.value.length === 0) return

  const container = document.querySelector("#pie-chart")
  if (!container || !container.parentElement) return

  if (pieChart) {
    pieChart.destroy()
    pieChart = null
  }

  // 清空容器
  container.innerHTML = ''

  const totalCount = taskStats.value.reduce((sum, item) => sum + item.count, 0)
  const textColor = getTextColor()
  const options = {
    series: taskStats.value.map(item => item.count),
    chart: {
      type: 'donut' as const,
      height: 275,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        speed: 500,
        animateGradually: { enabled: true, delay: 100 },
        dynamicAnimation: { enabled: true, speed: 350 }
      }
    },
    labels: taskStats.value.map(item => item.task_name),
    colors: ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#64748b'],
    legend: {
      show: false
    },
    stroke: {
      show: true,
      width: 2.5,
      colors: [isDark.value ? '#18181b' : '#ffffff']
    },
    plotOptions: {
      pie: {
        donut: {
          size: '70%',
          labels: {
            show: true,
            name: {
              show: true,
              fontSize: '11px',
              fontFamily: 'Inter, sans-serif',
              fontWeight: 500,
              color: textColor,
              offsetY: -8
            },
            value: {
              show: true,
              fontSize: '15px',
              fontFamily: 'Inter, sans-serif',
              fontWeight: 700,
              color: isDark.value ? '#f8fafc' : '#0f172a',
              offsetY: 4,
              formatter: (val: string) => val + ' 次'
            },
            total: {
              show: true,
              showAlways: true,
              label: '累计执行',
              fontSize: '11px',
              fontFamily: 'Inter, sans-serif',
              fontWeight: 500,
              color: textColor,
              formatter: () => String(totalCount) + ' 次'
            }
          }
        },
        expandOnClick: true
      }
    },
    dataLabels: {
      enabled: false
    },
    tooltip: {
      enabled: true,
      theme: isDark.value ? 'dark' : 'light',
      style: { fontSize: '12px', fontFamily: 'Inter, sans-serif' },
      y: { formatter: (val: number) => val + ' 次' }
    },
    responsive: [{ breakpoint: 768, options: { chart: { height: 240 }, legend: { position: 'bottom' } } }]
  }
  pieChart = new ApexCharts(container as HTMLElement, options as any)
  pieChart.render()
}

onMounted(async () => {
  window.addEventListener('resize', handleResize)

  try {
    const [statsData, sendStatsData, taskStatsData] = await Promise.all([
      api.dashboard.stats(),
      api.dashboard.sendStats(chartDays.value),
      api.dashboard.taskStats(chartDays.value)
    ])
    updateStats(statsData)
    sendStats.value = sendStatsData
    taskStats.value = taskStatsData

    // 渲染图表
    setTimeout(() => {
      renderLineChart()
      renderPieChart()
      chartsLoaded.value = true

      // 图表渲染完成后，延迟启动主题监听，避免初始化时的闪烁
      setTimeout(() => {
        isInitializing = false
        // 监听主题变化
        themeObserver = new MutationObserver(() => {
          handleThemeChange()
        })
        themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
      }, 500)
    }, 100)
  } catch { }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (themeObserver) {
    themeObserver.disconnect()
    themeObserver = null
  }
  if (themeChangeTimeout) {
    clearTimeout(themeChangeTimeout)
    themeChangeTimeout = null
  }
  if (lineChart) lineChart.destroy()
  if (pieChart) pieChart.destroy()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 头部问候与概览标题 -->
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <h2 class="text-2xl font-bold tracking-tight">数据仪表</h2>
        <p class="text-xs sm:text-sm text-muted-foreground mt-0.5">实时系统运行状态、任务调度活跃度与指标汇总</p>
      </div>
      <div class="shrink-0 pt-0.5 sm:pt-1">
        <Badge variant="outline" class="h-6 px-2 sm:px-2.5 text-[11px] font-normal border-border/80 text-muted-foreground gap-1.5 whitespace-nowrap">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          服务正常运行
        </Badge>
      </div>
    </div>

    <!-- 顶部 6 大关键指标卡片网格 -->
    <div class="grid gap-3.5 grid-cols-2 sm:grid-cols-3 lg:grid-cols-6">
      <div
        v-for="item in statItems"
        :key="item.key"
        @click="navigateTo(item.route)"
        class="group relative overflow-hidden rounded-xl border border-border/60 hover:border-border dark:hover:border-zinc-700/80 bg-card/70 hover:bg-card p-3.5 sm:p-4 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md cursor-pointer flex flex-col justify-between min-h-[96px] sm:min-h-[102px]"
      >
        <!-- 顶部 1.5px 极细素雅强调微光条 -->
        <div
          class="absolute top-0 left-0 right-0 h-[1.5px] transition-opacity duration-300 opacity-40 group-hover:opacity-85"
          :class="item.topBarClass"
        />

        <!-- 右下角巨幅微弱水印背景图标 (小屏隐藏，仅在具有充裕空间的大屏展示) -->
        <component
          :is="item.icon"
          class="hidden sm:block absolute -right-2.5 -bottom-2.5 w-16 h-16 opacity-[0.015] dark:opacity-[0.02] group-hover:opacity-[0.03] dark:group-hover:opacity-[0.04] group-hover:scale-105 transition-all duration-500 pointer-events-none select-none"
          :class="item.iconColor"
        />

        <!-- 卡片顶部：主题色小药丸点 + 标题 + 右侧轻量图标胶囊 -->
        <div class="flex items-center justify-between z-10">
          <div class="flex items-center gap-1.5 min-w-0">
            <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="item.dotClass" />
            <span class="text-xs font-medium text-muted-foreground group-hover:text-foreground transition-colors truncate">
              {{ item.label }}
            </span>
          </div>
          <div
            class="w-6 h-6 sm:w-7 sm:h-7 rounded-md flex items-center justify-center transition-all duration-300 group-hover:scale-105 shrink-0"
            :class="item.iconWrapperClass"
          >
            <component :is="item.icon" class="w-3.5 h-3.5" />
          </div>
        </div>

        <!-- 卡片底部：醒目大数字 + 单位 + 动态呼吸态 -->
        <div class="mt-2.5 flex items-baseline justify-between z-10">
          <div class="flex items-baseline gap-1">
            <span
              class="text-2xl sm:text-3xl font-extrabold tracking-tight font-['Inter'] tabular-nums"
              :class="item.key === 'running' && displayStats.running > 0 ? 'text-emerald-500' : 'text-foreground'"
            >
              {{ displayStats[item.key as keyof Stats] }}
            </span>
            <span class="text-[11px] text-muted-foreground/60 font-normal">
              {{ item.unit }}
            </span>
          </div>

          <!-- 状态指示灯或悬停微动微标 -->
          <div v-if="item.key === 'running' && displayStats.running > 0" class="flex items-center gap-1 text-[10px] font-medium text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 px-1.5 py-0.5 rounded-full">
            <span class="relative flex h-1.5 w-1.5">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500"></span>
            </span>
            <span>运行中</span>
          </div>
          <ArrowUpRight
            v-else
            class="w-3.5 h-3.5 text-muted-foreground/15 dark:text-muted-foreground/10 group-hover:text-muted-foreground/50 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-all shrink-0"
          />
        </div>
      </div>
    </div>

    <!-- 中部图表分析区域 -->
    <div class="grid gap-4 lg:grid-cols-10 items-stretch">
      <!-- 左侧：执行统计（经典折线图） -->
      <Card class="lg:col-span-7 shadow-sm border-border/60 hover:border-border transition-colors flex flex-col relative overflow-hidden py-0">
        <!-- 顶部素雅极细微光条 -->
        <div class="absolute top-0 left-0 right-0 h-[1.5px] bg-gradient-to-r from-sky-500/30 via-slate-400/20 to-transparent" />
        <CardHeader class="px-4 sm:px-5 pt-3.5 pb-2.5 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-border/40">
          <div class="space-y-0.5">
            <CardTitle class="text-sm sm:text-base flex items-center gap-2">
              <div class="p-1.5 rounded-lg bg-sky-500/10 text-sky-600 dark:text-sky-400 shrink-0">
                <TrendingUp class="w-4 h-4" />
              </div>
              <span>执行统计趋势</span>
            </CardTitle>
            <CardDescription class="text-xs">最近 {{ chartDays }} 天任务触发与成功/失败分布</CardDescription>
          </div>
          <div class="flex items-center gap-2 self-start sm:self-auto">
            <Badge variant="secondary" class="font-normal text-[11px] font-['Inter'] tabular-nums px-2 py-0.5 bg-muted/60 text-muted-foreground border-border/60">
              周期总计: <span class="font-medium text-foreground ml-1">{{ totalExecutions }} 次</span>
            </Badge>
            <Badge variant="outline" class="font-normal text-[11px] font-['Inter'] tabular-nums px-2 py-0.5 text-emerald-600 dark:text-emerald-400 border-emerald-500/30 bg-emerald-500/5 gap-1">
              <CheckCircle2 class="w-3 h-3 text-emerald-500" />
              {{ overallSuccessRate }}% 成功率
            </Badge>
          </div>
        </CardHeader>
        <CardContent class="relative pt-2 pb-2.5 px-2 sm:px-4 flex-1 min-h-[280px] sm:min-h-[310px] flex flex-col justify-center">
          <div id="stats-chart" class="w-full h-full min-h-[260px] sm:min-h-[290px]"></div>
          <Transition
            enter-active-class="transition-opacity duration-200"
            leave-active-class="transition-opacity duration-200 pointer-events-none"
            enter-from-class="opacity-0"
            leave-to-class="opacity-0"
          >
            <div v-if="!chartsLoaded"
              class="absolute inset-0 flex items-center justify-center text-muted-foreground text-xs bg-card/80 backdrop-blur-sm z-10">
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-primary animate-ping"></span>
                <span>数据加载中...</span>
              </div>
            </div>
          </Transition>
        </CardContent>
      </Card>

      <!-- 右侧：任务占比（甜甜圈环形图） -->
      <Card class="lg:col-span-3 shadow-sm border-border/60 hover:border-border transition-colors flex flex-col relative overflow-hidden py-0">
        <!-- 顶部素雅极细微光条 -->
        <div class="absolute top-0 left-0 right-0 h-[1.5px] bg-gradient-to-r from-indigo-500/30 via-slate-400/20 to-transparent" />
        <CardHeader class="px-4 sm:px-5 pt-3.5 pb-2.5 flex flex-row items-center justify-between gap-2 border-b border-border/40">
          <div class="space-y-0.5">
            <CardTitle class="text-sm sm:text-base flex items-center gap-2">
              <div class="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 shrink-0">
                <PieIcon class="w-4 h-4" />
              </div>
              <span>任务占比</span>
            </CardTitle>
            <CardDescription class="text-xs">各任务执行频次权重分布</CardDescription>
          </div>
          <Badge variant="secondary" class="font-normal text-[10px] font-['Inter'] px-1.5 py-0 bg-muted/60 text-muted-foreground border-border/60">
            Top 任务
          </Badge>
        </CardHeader>
        <CardContent class="relative pt-2 pb-2.5 px-4 flex-1 min-h-[280px] sm:min-h-[310px] flex items-center justify-center">
          <div id="pie-chart" class="w-full h-full min-h-[240px] sm:min-h-[275px] flex items-center justify-center"></div>
          <Transition
            enter-active-class="transition-opacity duration-200"
            leave-active-class="transition-opacity duration-200 pointer-events-none"
            enter-from-class="opacity-0"
            leave-to-class="opacity-0"
          >
            <div v-if="!chartsLoaded"
              class="absolute inset-0 flex items-center justify-center text-muted-foreground text-xs bg-card/80 backdrop-blur-sm z-10">
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-primary animate-ping"></span>
                <span>数据加载中...</span>
              </div>
            </div>
          </Transition>
          <div v-if="chartsLoaded && taskStats.length === 0"
            class="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground text-xs bg-card/80">
            <PieIcon class="w-8 h-8 opacity-20 mb-2" />
            <span>暂无任务执行数据</span>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

