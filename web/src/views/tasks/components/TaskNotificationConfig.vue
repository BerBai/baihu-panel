<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { Bell, BellOff, CheckCircle2, XCircle, Clock, FileText, Variable, Plus, Trash2, ExternalLink, Info } from 'lucide-vue-next'
import { api, type NotifyChannel, type NotifyBinding, type CustomVarRule } from '@/api'

const props = defineProps<{
  taskId?: string
}>()

const notifyChannels = ref<NotifyChannel[]>([])
const notifyWayId = ref<string>('none')
const notifyOnSuccess = ref(false)
const notifyOnFailure = ref(false)
const notifyOnTimeout = ref(false)
const notifyIncludeLog = ref(false)
const notifyLogLimit = ref(1000)
const customVars = ref<CustomVarRule[]>([])

function addCustomVar() {
  customVars.value.push({ name: '', regex: '', default: '' })
}

function removeCustomVar(index: number) {
  customVars.value.splice(index, 1)
}

const currentChannel = computed(() => {
  if (notifyWayId.value === 'none') return null
  return notifyChannels.value.find(c => c.id === notifyWayId.value) || null
})

const currentChannelName = computed(() => {
  if (notifyWayId.value === 'none') return '不启用通知'
  return currentChannel.value ? currentChannel.value.name : '已配置渠道'
})

onMounted(async () => {
  try {
    notifyChannels.value = await api.notify.getChannels()
  } catch (e) {
    console.error('Fetch channels failed', e)
  }
})

function resetConfig() {
  notifyWayId.value = 'none'
  notifyOnSuccess.value = false
  notifyOnFailure.value = false
  notifyOnTimeout.value = false
  notifyIncludeLog.value = false
  notifyLogLimit.value = 1000
  customVars.value = []
}

async function loadConfig(taskId?: string) {
  if (!taskId) {
    resetConfig()
    return
  }

  try {
    const allBindings = await api.notify.getBindings()
    // 过滤出该任务的所有绑定
    const taskBindings = allBindings.filter(b => b.data_id === taskId && b.type === 'task')

    if (taskBindings.length > 0 && taskBindings[0]) {
      notifyWayId.value = taskBindings[0].way_id || 'none'
      
      // 直接设置，无需 setTimeout，内部已经安全了
      notifyOnSuccess.value = taskBindings.some(b => b.event === 'task_success')
      notifyOnFailure.value = taskBindings.some(b => b.event === 'task_failed')
      notifyOnTimeout.value = taskBindings.some(b => b.event === 'task_timeout')
      
      const extraBinding = taskBindings.find(b => b.extra && b.extra !== '')
      if (extraBinding && extraBinding.extra) {
        try {
          const extra = JSON.parse(extraBinding.extra)
          notifyIncludeLog.value = !!extra.enable_log
          notifyLogLimit.value = extra.log_limit || 1000
          customVars.value = Array.isArray(extra.custom_vars) ? extra.custom_vars : []
        } catch {
          notifyIncludeLog.value = false
          notifyLogLimit.value = 1000
          customVars.value = []
        }
      } else {
        notifyIncludeLog.value = false
        notifyLogLimit.value = 1000
        customVars.value = []
      }
    } else {
      resetConfig()
    }
  } catch (e) {
    console.error('Load notifications failed', e)
    resetConfig()
  }
}

async function saveConfig(taskId: string) {
  try {
    const bindings: Partial<NotifyBinding>[] = []

    if (notifyWayId.value !== 'none') {
      const events = [
        { type: 'task_success', enabled: notifyOnSuccess.value },
        { type: 'task_failed', enabled: notifyOnFailure.value },
        { type: 'task_timeout', enabled: notifyOnTimeout.value }
      ]

      // 仅保留名称非空的有效规则
      const validCustomVars = customVars.value.filter(v => v.name && v.name.trim() !== '')

      const extra = JSON.stringify({
        enable_log: notifyIncludeLog.value,
        log_limit: notifyLogLimit.value,
        custom_vars: validCustomVars
      })

      for (const event of events) {
        if (event.enabled) {
          bindings.push({
            event: event.type,
            way_id: notifyWayId.value,
            extra: extra
          })
        }
      }
    }

    // 调用批量保存，后端会自动清理该任务旧的绑定
    await api.notify.saveBindingsBatch({
      type: 'task',
      data_id: taskId,
      bindings: bindings
    })
  } catch (e) {
    console.error('Save notifications failed', e)
  }
}

defineExpose({
  loadConfig,
  saveConfig
})
</script>

<template>
  <section class="space-y-4">
    <div class="flex items-center gap-2 mb-2">
      <div class="h-4 w-1 bg-primary rounded-full shadow-sm shadow-primary/20" />
      <h3 class="text-sm font-bold text-foreground/90">通知配置</h3>
    </div>

    <div class="grid gap-5 pl-3 border-l border-muted">
      <!-- 通知渠道 (通栏左文右控卡片) -->
      <div class="grid grid-cols-1 sm:grid-cols-4 items-start gap-3">
        <Label class="sm:text-right text-xs text-foreground/70 uppercase tracking-wider font-bold pt-2.5">通知渠道</Label>
        <div class="sm:col-span-3">
          <div class="p-3 rounded-xl bg-muted/15 border border-muted-foreground/10 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div class="space-y-0.5">
              <div class="flex items-center gap-2">
                <Bell v-if="notifyWayId !== 'none'" class="h-4 w-4 text-primary shrink-0" />
                <BellOff v-else class="h-4 w-4 text-muted-foreground shrink-0" />
                <span class="text-xs font-bold text-foreground">
                  {{ currentChannelName }}
                </span>
                <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="notifyWayId !== 'none' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'">
                  {{ notifyWayId !== 'none' ? '已配置推送' : '未启用' }}
                </span>
              </div>
              <p class="text-[11px] text-muted-foreground">
                {{ notifyWayId !== 'none' ? '任务运行状态变化时，将向该渠道发送推送通知' : '当前不向外部推送任何执行结果通知' }}
              </p>
            </div>
            <div class="self-start sm:self-auto shrink-0">
              <Select v-model="notifyWayId">
                <SelectTrigger class="w-48 h-8 bg-background border-muted-foreground/15 px-2.5 text-xs">
                  <SelectValue placeholder="选择通知渠道..." class="truncate">
                    <div class="flex items-center gap-1.5 truncate">
                      <div class="w-1.5 h-1.5 rounded-full" :class="notifyWayId !== 'none' ? 'bg-primary' : 'bg-muted-foreground/40'" />
                      <span>{{ currentChannelName }}</span>
                    </div>
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none" class="text-xs">
                    <div class="flex items-center gap-2">
                      <div class="w-1.5 h-1.5 rounded-full bg-muted-foreground/40" />
                      <span>不启用通知</span>
                    </div>
                  </SelectItem>
                  <SelectItem v-for="ch in notifyChannels" :key="ch.id" :value="ch.id" class="text-xs">
                    <div class="flex items-center gap-2">
                      <div class="w-1.5 h-1.5 rounded-full" :class="ch.enabled ? 'bg-emerald-500' : 'bg-muted-foreground'" />
                      <span class="font-medium">{{ ch.name }}</span>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
      </div>

      <template v-if="notifyWayId !== 'none'">
        <!-- 通知时机 -->
        <div class="grid grid-cols-1 sm:grid-cols-4 items-start gap-3">
          <Label class="sm:text-right text-xs text-foreground/70 uppercase tracking-wider font-bold pt-2.5">通知时机</Label>
          <div class="sm:col-span-3 space-y-2.5">
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              <!-- 成功时 -->
              <div
                class="p-2.5 rounded-xl border cursor-pointer transition-all flex items-center justify-between select-none"
                :class="notifyOnSuccess
                  ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 shadow-xs'
                  : 'border-muted-foreground/15 bg-muted/10 hover:bg-muted/20 text-muted-foreground'"
                @click="notifyOnSuccess = !notifyOnSuccess"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <CheckCircle2 class="h-4 w-4 shrink-0" :class="notifyOnSuccess ? 'text-emerald-500' : 'text-muted-foreground/60'" />
                  <span class="text-xs font-semibold">执行成功</span>
                </div>
                <Checkbox :checked="notifyOnSuccess" class="pointer-events-none data-[state=checked]:bg-emerald-500 data-[state=checked]:border-emerald-500" />
              </div>

              <!-- 失败时 -->
              <div
                class="p-2.5 rounded-xl border cursor-pointer transition-all flex items-center justify-between select-none"
                :class="notifyOnFailure
                  ? 'border-rose-500/40 bg-rose-500/10 text-rose-700 dark:text-rose-300 shadow-xs'
                  : 'border-muted-foreground/15 bg-muted/10 hover:bg-muted/20 text-muted-foreground'"
                @click="notifyOnFailure = !notifyOnFailure"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <XCircle class="h-4 w-4 shrink-0" :class="notifyOnFailure ? 'text-rose-500' : 'text-muted-foreground/60'" />
                  <span class="text-xs font-semibold">执行失败</span>
                </div>
                <Checkbox :checked="notifyOnFailure" class="pointer-events-none data-[state=checked]:bg-rose-500 data-[state=checked]:border-rose-500" />
              </div>

              <!-- 超时时 -->
              <div
                class="p-2.5 rounded-xl border cursor-pointer transition-all flex items-center justify-between select-none"
                :class="notifyOnTimeout
                  ? 'border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300 shadow-xs'
                  : 'border-muted-foreground/15 bg-muted/10 hover:bg-muted/20 text-muted-foreground'"
                @click="notifyOnTimeout = !notifyOnTimeout"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <Clock class="h-4 w-4 shrink-0" :class="notifyOnTimeout ? 'text-amber-500' : 'text-muted-foreground/60'" />
                  <span class="text-xs font-semibold">执行超时</span>
                </div>
                <Checkbox :checked="notifyOnTimeout" class="pointer-events-none data-[state=checked]:bg-amber-500 data-[state=checked]:border-amber-500" />
              </div>
            </div>
          </div>
        </div>

        <!-- 消息内容 -->
        <div class="grid grid-cols-1 sm:grid-cols-4 items-start gap-3">
          <Label class="sm:text-right text-xs text-foreground/70 uppercase tracking-wider font-bold pt-2.5">消息内容</Label>
          <div class="sm:col-span-3">
            <div class="p-3 rounded-xl bg-muted/15 border border-muted-foreground/10 space-y-3">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div class="space-y-0.5">
                  <div class="flex items-center gap-2">
                    <FileText class="h-4 w-4 text-muted-foreground shrink-0" />
                    <span class="text-xs font-bold text-foreground">附带运行日志</span>
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="notifyIncludeLog ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'">
                      {{ notifyIncludeLog ? '已包含' : '仅基础信息' }}
                    </span>
                  </div>
                  <p class="text-[11px] text-muted-foreground">
                    在推送的消息中附加该任务末尾的运行日志输出
                  </p>
                </div>
                <div class="self-start sm:self-auto shrink-0">
                  <Switch v-model="notifyIncludeLog" class="data-[state=checked]:bg-primary" />
                </div>
              </div>

              <!-- 开启日志后的字数设置 -->
              <div v-if="notifyIncludeLog" class="pt-2.5 border-t border-muted-foreground/10 flex items-center justify-between animate-in fade-in slide-in-from-top-1 duration-200">
                <span class="text-xs text-muted-foreground">日志截取长度限制</span>
                <div class="flex items-center gap-1.5">
                  <span class="text-xs text-muted-foreground font-medium">保留末尾</span>
                  <input
                    type="text"
                    inputmode="numeric"
                    :value="notifyLogLimit"
                    @input="(e: any) => notifyLogLimit = Number(e.target.value.replace(/\D/g, ''))"
                    class="w-20 h-8 text-xs font-semibold text-center bg-background rounded-md border border-muted-foreground/20 focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                  <span class="text-xs text-muted-foreground font-medium">字</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 变量提取 (Issue #196) -->
        <div class="grid grid-cols-1 sm:grid-cols-4 items-start gap-3">
          <Label class="sm:text-right text-xs text-foreground/70 uppercase tracking-wider font-bold pt-2.5">变量提取</Label>
          <div class="sm:col-span-3">
            <div class="p-3 rounded-xl bg-muted/15 border border-muted-foreground/10 space-y-3">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div class="space-y-0.5">
                  <div class="flex items-center gap-2">
                    <Variable class="h-4 w-4 text-primary shrink-0" />
                    <span class="text-xs font-bold text-foreground">日志变量正则提取</span>
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-medium bg-primary/10 text-primary">
                      {{ customVars.length > 0 ? `${customVars.length} 个规则` : '未配置' }}
                    </span>
                  </div>
                  <p class="text-[11px] text-muted-foreground">
                    通过正则提取日志内容注入到通知模板变量（如 <span v-pre class="inline-block whitespace-nowrap px-1.5 py-0.5 rounded bg-primary/10 text-primary font-medium text-[10.5px]">{{coins}}</span>）
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  class="h-7 text-xs gap-1 border-muted-foreground/20 hover:border-primary/50 hover:bg-primary/5 self-start sm:self-auto shrink-0"
                  @click="addCustomVar"
                >
                  <Plus class="h-3.5 w-3.5" />
                  <span>添加规则</span>
                </Button>
              </div>

              <!-- 规则列表 -->
              <div v-if="customVars.length > 0" class="space-y-2 pt-1">
                <div
                  v-for="(rule, idx) in customVars"
                  :key="idx"
                  class="p-2.5 rounded-lg bg-background border border-muted-foreground/15 flex flex-col md:flex-row items-stretch md:items-center gap-2 animate-in fade-in-50 duration-200"
                >
                  <!-- 变量名 -->
                  <div class="flex-1 min-w-[110px]">
                    <div class="text-[10px] font-medium text-muted-foreground mb-1">变量名称 (Key)</div>
                    <input
                      v-model="rule.name"
                      type="text"
                      placeholder="如: coins"
                      class="w-full h-7 px-2 text-xs font-inter bg-muted/20 rounded border border-muted-foreground/20 focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>

                  <!-- 正则表达式 -->
                  <div class="flex-[2] min-w-[180px]">
                    <div class="text-[10px] font-medium text-muted-foreground mb-1">捕获正则 (优先匹配第 1 捕获组)</div>
                    <input
                      v-model="rule.regex"
                      type="text"
                      placeholder="如: 获得硬币[：:]\s*(\d+)"
                      class="w-full h-7 px-2 text-xs font-inter bg-muted/20 rounded border border-muted-foreground/20 focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>

                  <!-- 缺省值 -->
                  <div class="flex-1 min-w-[90px]">
                    <div class="text-[10px] font-medium text-muted-foreground mb-1">未匹配默认值 (可选)</div>
                    <input
                      v-model="rule.default"
                      type="text"
                      placeholder="如: 0"
                      class="w-full h-7 px-2 text-xs bg-muted/20 rounded border border-muted-foreground/20 focus:outline-none focus:ring-1 focus:ring-primary"
                    />
                  </div>

                  <!-- 删除按钮 -->
                  <div class="shrink-0 self-end md:self-center md:pt-4">
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      class="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                      title="删除规则"
                      @click="removeCustomVar(idx)"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              </div>

              <!-- 协议与使用说明 (专业重排布局) -->
              <div class="pt-3 border-t border-muted-foreground/10 space-y-2">
                <div class="flex items-center justify-between text-xs">
                  <div class="flex items-center gap-1.5 text-foreground/90 font-medium">
                    <Info class="h-3.5 w-3.5 text-primary shrink-0" />
                    <span>免配置直出协议与使用说明</span>
                  </div>
                  <a
                    href="https://engigu.github.io/baihu-panel/guide/notify.html#进阶-自定义变量提取与模板注入-issue-196"
                    target="_blank"
                    rel="noreferrer"
                    class="inline-flex items-center gap-1 text-[11px] text-primary hover:underline font-medium transition-colors"
                  >
                    <span>查看完整文档</span>
                    <ExternalLink class="h-3 w-3" />
                  </a>
                </div>

                <div class="grid gap-1.5 pl-5 text-[11px] text-muted-foreground leading-relaxed">
                  <div class="flex flex-wrap items-center gap-x-1.5 gap-y-1">
                    <span>• 脚本可在日志直接打印</span>
                    <span class="inline-block whitespace-nowrap px-1.5 py-0.5 rounded bg-muted/80 text-foreground border border-muted-foreground/15 text-[10.5px]">
                      [BHP_VAR] coins=50
                    </span>
                    <span>或</span>
                    <span class="inline-block whitespace-nowrap px-1.5 py-0.5 rounded bg-muted/80 text-foreground border border-muted-foreground/15 text-[10.5px]">
                      ::set-output name=coins::50
                    </span>
                    <span>，系统将自动无感提取。</span>
                  </div>
                  <div class="flex flex-wrap items-center gap-x-1.5 gap-y-1">
                    <span>• 提取后可在渠道模板中直接使用</span>
                    <span v-pre class="inline-block whitespace-nowrap px-1.5 py-0.5 rounded bg-primary/10 text-primary font-medium text-[10.5px]">
                      {{coins}}
                    </span>
                    <span>占位符引用。</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </section>
</template>

<style scoped>
:deep(*) {
  text-rendering: optimizeLegibility;
}
:deep(label) {
  text-rendering: optimizeLegibility;
  letter-spacing: 0.01em;
}
:deep(.font-inter),
.font-inter {
  font-family: 'Inter', 'Inter Variable', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif !important;
  letter-spacing: -0.01em;
}
</style>

