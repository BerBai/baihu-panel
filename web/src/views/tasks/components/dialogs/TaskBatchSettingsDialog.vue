<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import {
  Wrench, Clock, RotateCcw, Tag, Server, Laptop,
  AlertCircle, Loader2, Sparkles
} from 'lucide-vue-next'
import { api, type Task, type Agent } from '@/api'
import { TASK_TYPE } from '@/constants'
import { toast } from 'vue-sonner'
import TaskTagsConfig from '../TaskTagsConfig.vue'
import TaskLangConfig, { type LangConfig } from '../TaskLangConfig.vue'

// ==========================================
// 1. 常量 Map：参数全集元数据定义 (Param Registry)
// ==========================================
type TaskTypeKey = 'task' | 'repo' | 'app' | 'agent_sync_script'

interface ParamDefinition {
  id: string
  label: string
  icon: any
  description: string
  category: 'common' | 'language' | 'script_only'
  // 允许受此参数影响的实体类型
  supportedTypes: TaskTypeKey[]
}

type ParamKey = 'timeout' | 'tags' | 'retry' | 'envs' | 'languages' | 'agent' // | 'commands'

const PARAM_DEFINITIONS: Record<ParamKey, ParamDefinition> = {
  timeout: {
    id: 'timeout',
    label: '执行超时时间 (分钟)',
    icon: Clock,
    description: '单次运行的最长允许时间（默认 30 分钟，0是不限制）',
    category: 'common',
    supportedTypes: [TASK_TYPE.NORMAL, TASK_TYPE.REPO, TASK_TYPE.APP, TASK_TYPE.AGENT_SYNC_SCRIPT]
  },
  tags: {
    id: 'tags',
    label: '批量标签管理 (Tags)',
    icon: Tag,
    description: '追加或全量覆盖任务的分类标签',
    category: 'common',
    supportedTypes: [TASK_TYPE.NORMAL, TASK_TYPE.REPO, TASK_TYPE.APP, TASK_TYPE.AGENT_SYNC_SCRIPT]
  },
  retry: {
    id: 'retry',
    label: '调度重试与随机延迟策略',
    icon: RotateCcw,
    description: '配置失败自动重试次数、重试间隔及启动延迟打散',
    category: 'common',
    supportedTypes: [TASK_TYPE.NORMAL, TASK_TYPE.REPO, TASK_TYPE.APP, TASK_TYPE.AGENT_SYNC_SCRIPT]
  },
  envs: {
    id: 'envs',
    label: '全量环境变量注入开关',
    icon: AlertCircle,
    description: '控制任务在执行时是否自动注入面板全部系统环境变量',
    category: 'common',
    supportedTypes: [TASK_TYPE.NORMAL, TASK_TYPE.REPO, TASK_TYPE.APP, TASK_TYPE.AGENT_SYNC_SCRIPT]
  },
  languages: {
    id: 'languages',
    label: '运行时多语言环境 (Mise)',
    icon: Sparkles,
    description: '批量将所选任务的运行环境替换为指定的语言运行时版本 (支持脚本、仓库和应用)',
    category: 'language', // 单独分类：语言环境
    supportedTypes: [TASK_TYPE.NORMAL, TASK_TYPE.REPO, TASK_TYPE.APP] // 全部普通脚本、仓库同步和应用实体
  },
  agent: {
    id: 'agent',
    label: '统一执行位置 (Agent 节点)',
    icon: Server,
    description: '将普通脚本任务一键重定向至指定的本地面板服务或远程 Agent 执行 (仅对普通脚本任务有效)',
    category: 'script_only',
    supportedTypes: [TASK_TYPE.NORMAL] // 仅普通脚本任务有效
  }
  // commands: {
  //   id: 'commands',
  //   label: '前置 / 后置指令钩子 (CLI Hooks)',
  //   icon: Terminal,
  //   description: '在任务主脚本运行前或运行后附加执行的自定义 Shell 钩子命令',
  //   category: 'script_only',
  //   supportedTypes: [TASK_TYPE.NORMAL]
  // }
}

// ==========================================
// 2. 组件 Props 与 Emits
// ==========================================
const props = defineProps<{
  open: boolean
  selectedTasks?: Task[]
  selectedIds?: string[]
  isQueryMode?: boolean
  queryFilter?: Record<string, any>
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'saved': []
}>()

const loading = ref(false)
const agents = ref<Agent[]>([])
const activeCategoryTab = ref<'all' | 'common' | 'language' | 'script_only'>('all')

// 字段级生效开关 (Patch Masks)
const applyAgent = ref(false)
const applyTimeout = ref(false)
const applyRetry = ref(false)
const applyTags = ref(false)
// const applyPreCmd = ref(false)
// const applyPostCmd = ref(false)
const applyLanguages = ref(false)
const applyEnvs = ref(false)

// 表单字段输入值状态
const selectedAgentId = ref('local')
const timeoutVal = ref(30)
const retryCountVal = ref(0)
const retryIntervalVal = ref(0)
const randomRangeVal = ref(0)
const tagsVal = ref('')
const tagMode = ref<'append' | 'replace'>('append')
// const preCommandVal = ref('')
// const postCommandVal = ref('')
const selectedLangs = ref<LangConfig[]>([])
const allEnvsVal = ref(false)

// ==========================================
// 3. 任务类型统计与支持度计算 (Scope Calculation)
// ==========================================
const totalCount = computed(() => {
  return props.selectedTasks?.length || props.selectedIds?.length || 0
})

const typeStats = computed(() => {
  const stats: Record<string, number> = {}
  ;(props.selectedTasks || []).forEach(t => {
    const type = t.type || TASK_TYPE.NORMAL
    stats[type] = (stats[type] || 0) + 1
  })
  return stats
})

/**
 * 动态计算各参数在当前选中的任务中的生效范围
 */
function getParamScope(paramKey: keyof typeof PARAM_DEFINITIONS) {
  const meta = PARAM_DEFINITIONS[paramKey]
  if (!meta) return { isSupported: false, isFull: false, applicableCount: 0, badgeText: '' }

  if (props.isQueryMode) {
    const queryType = props.queryFilter?.type
    if (!queryType || queryType === 'all') {
      const isFull = meta.supportedTypes.length >= 4
      let badge = '全部适用'
      if (!isFull) {
        if (paramKey === 'agent') badge = '仅普通脚本生效'
        else if (paramKey === 'languages') badge = '仅脚本/仓库/应用生效'
        else badge = '部分类型生效'
      }
      return { isSupported: true, isFull, applicableCount: 0, badgeText: badge }
    }
    const matches = meta.supportedTypes.includes(queryType as TaskTypeKey)
    return { isSupported: matches, isFull: matches, applicableCount: 0, badgeText: matches ? '全部适用' : '当前类型不支持' }
  }

  const tasks = props.selectedTasks || []
  const total = tasks.length
  if (total === 0) {
    return { isSupported: true, isFull: true, applicableCount: 0, badgeText: '全部适用' }
  }

  const applicableCount = tasks.filter(t => meta.supportedTypes.includes((t.type || TASK_TYPE.NORMAL) as TaskTypeKey)).length
  const isSupported = applicableCount > 0
  const isFull = applicableCount === total

  let badgeText = ''
  if (isFull) {
    badgeText = '全部适用'
  } else if (applicableCount > 0) {
    if (paramKey === 'agent') {
      badgeText = `仅对普通脚本生效 (${applicableCount}/${total})`
    } else if (paramKey === 'languages') {
      badgeText = `仅对脚本/仓库/应用生效 (${applicableCount}/${total})`
    } else {
      badgeText = `仅对部分任务生效 (${applicableCount}/${total})`
    }
  } else {
    if (paramKey === 'agent') {
      badgeText = '仅普通脚本支持'
    } else if (paramKey === 'languages') {
      badgeText = '仅脚本/仓库/应用支持'
    } else {
      badgeText = '当前选中项不支持'
    }
  }

  return {
    isSupported,
    isFull,
    applicableCount,
    totalCount: total,
    badgeText
  }
}

// 各参数支持度缓存
const scopeAgent = computed(() => getParamScope('agent'))
const scopeTimeout = computed(() => getParamScope('timeout'))
const scopeTags = computed(() => getParamScope('tags'))
const scopeRetry = computed(() => getParamScope('retry'))
const scopeEnvs = computed(() => getParamScope('envs'))
const scopeLanguages = computed(() => getParamScope('languages'))
// const scopeCommands = computed(() => getParamScope('commands'))

// 检查是否存在特化分类参数受支持 (如语言环境或脚本专属)
const hasCategoryTabs = computed(() => {
  return scopeLanguages.value.isSupported || scopeAgent.value.isSupported
})

// 检查当前选择是否为多类型混用
const isMixedTypes = computed(() => {
  const activeTypes = Object.keys(typeStats.value).filter(k => (typeStats.value[k] || 0) > 0)
  return activeTypes.length > 1
})

// 计算已激活开关的数量
const activeCount = computed(() => {
  let count = 0
  if (applyAgent.value && scopeAgent.value.isSupported) count++
  if (applyTimeout.value && scopeTimeout.value.isSupported) count++
  if (applyRetry.value && scopeRetry.value.isSupported) count++
  if (applyTags.value && scopeTags.value.isSupported) count++
  if (applyLanguages.value && scopeLanguages.value.isSupported) count++
  // if ((applyPreCmd.value || applyPostCmd.value) && scopeCommands.value.isSupported) count++
  if (applyEnvs.value && scopeEnvs.value.isSupported) count++
  return count
})

// ==========================================
// 4. 重置与初始化
// ==========================================
watch(() => props.open, async (val) => {
  if (val) {
    // 重置所有激活开关
    applyAgent.value = false
    applyTimeout.value = false
    applyRetry.value = false
    applyTags.value = false
    // applyPreCmd.value = false
    // applyPostCmd.value = false
    applyLanguages.value = false
    applyEnvs.value = false

    // 重置表单初始值
    selectedAgentId.value = 'local'
    timeoutVal.value = 30
    retryCountVal.value = 0
    retryIntervalVal.value = 0
    randomRangeVal.value = 0
    tagsVal.value = ''
    tagMode.value = 'append'
    // preCommandVal.value = ''
    // postCommandVal.value = ''
    selectedLangs.value = []
    allEnvsVal.value = false
    activeCategoryTab.value = 'all'

    try {
      agents.value = await api.agents.list()
    } catch { /* ignore */ }
  }
})

// ==========================================
// 5. 精准分流保存 (Selective Patching)
// ==========================================
async function handleSave() {
  if (activeCount.value === 0) {
    toast.warning('请先开启至少一项需要批量调整的参数开关')
    return
  }

  const allSelectedTasks = props.selectedTasks || []
  const allTargetIds = props.selectedIds || allSelectedTasks.map(t => t.id)
  if (!props.isQueryMode && allTargetIds.length === 0) {
    toast.error('未指定任何目标任务')
    return
  }

  loading.value = true
  try {
    // 1. 全员通用字段 (适用于全部实体类型: task, repo, app, agent_sync_script)
    const commonFields: Record<string, any> = {}
    if (applyTimeout.value && scopeTimeout.value.isSupported) {
      commonFields.timeout = Number(timeoutVal.value) || 30
    }
    if (applyRetry.value && scopeRetry.value.isSupported) {
      commonFields.retry_count = Number(retryCountVal.value) || 0
      commonFields.retry_interval = Number(retryIntervalVal.value) || 0
      commonFields.random_range = Number(randomRangeVal.value) || 0
    }
    if (applyTags.value && scopeTags.value.isSupported) {
      commonFields.tags = tagsVal.value
      commonFields.tag_mode = tagMode.value
    }
    if (applyEnvs.value && scopeEnvs.value.isSupported) {
      commonFields.unified_config_patch = {
        common: { task_all_envs: !!allEnvsVal.value }
      }
    }

    // 2. 普通脚本专属字段 (仅对 TASK_TYPE.NORMAL 生效: agent)
    const normalScriptFields: Record<string, any> = {}
    if (applyAgent.value && scopeAgent.value.isSupported) {
      normalScriptFields.agent_id = selectedAgentId.value === 'local' ? '' : selectedAgentId.value
    }

    // 3. 多语言环境字段 (对 task, repo, app 生效: languages)
    const langFields: Record<string, any> = {}
    // if (applyPreCmd.value && scopeCommands.value.isSupported) {
    //   scriptFields.pre_command = preCommandVal.value
    // }
    // if (applyPostCmd.value && scopeCommands.value.isSupported) {
    //   scriptFields.post_command = postCommandVal.value
    // }
    if (applyLanguages.value && scopeLanguages.value.isSupported && selectedLangs.value.length > 0) {
      langFields.language_update_mode = 'replace'
      langFields.languages = selectedLangs.value.map(l => ({
        name: l.name,
        version: l.version
      }))
    }

    const hasCommonFields = Object.keys(commonFields).length > 0
    const hasNormalFields = Object.keys(normalScriptFields).length > 0
    const hasLangFields = Object.keys(langFields).length > 0

    // 按任务类型进行精准分流提交 (杜绝向不支持的类型下发脏数据)
    if (props.isQueryMode) {
      // 查询模式：直接单请求
      const update_fields = { ...commonFields, ...normalScriptFields, ...langFields }
      await api.tasks.batchUpdate({
        filter: props.queryFilter,
        update_fields
      })
    } else {
      // 显式 ID 勾选模式
      const normalTaskIds = allSelectedTasks
        .filter(t => (t.type || TASK_TYPE.NORMAL) === TASK_TYPE.NORMAL)
        .map(t => t.id)

      const langSupportedTaskIds = allSelectedTasks
        .filter(t => [TASK_TYPE.NORMAL, TASK_TYPE.REPO, TASK_TYPE.APP].includes((t.type || TASK_TYPE.NORMAL) as any))
        .map(t => t.id)

      const isPureNormalTasks = normalTaskIds.length === allTargetIds.length

      if (isPureNormalTasks || (!hasNormalFields && !hasLangFields)) {
        // 场景 A: 全量为普通脚本任务，或者未修改专属字段 -> 单次更新全部任务
        const update_fields = { ...commonFields, ...normalScriptFields, ...langFields }
        await api.tasks.batchUpdate({
          ids: allTargetIds,
          update_fields
        })
      } else {
        // 场景 B: 混合类型多选 (Selective Patching)
        const requests = []
        // 1. 通用字段分发给全部选中任务
        if (hasCommonFields) {
          requests.push(api.tasks.batchUpdate({
            ids: allTargetIds,
            update_fields: commonFields
          }))
        }
        // 2. Agent 节点仅分发给普通脚本任务
        if (hasNormalFields && normalTaskIds.length > 0) {
          requests.push(api.tasks.batchUpdate({
            ids: normalTaskIds,
            update_fields: normalScriptFields
          }))
        }
        // 3. 语言环境仅分发给支持多语言的任务 (task, repo, app)
        if (hasLangFields && langSupportedTaskIds.length > 0) {
          requests.push(api.tasks.batchUpdate({
            ids: langSupportedTaskIds,
            update_fields: langFields
          }))
        }
        await Promise.all(requests)
      }
    }

    toast.success('已成功应用批量参数设置')
    emit('update:open', false)
    emit('saved')
  } catch (err: any) {
    toast.error(err.message || '批量设置失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-[95vw] sm:max-w-[650px] xl:max-w-[760px] p-0 overflow-hidden border-none bg-background shadow-2xl">
      <div class="flex flex-col max-h-[85vh]">
        <!-- 弹窗顶栏 -->
        <DialogHeader class="px-6 pt-5 pb-3 border-b border-border/40 shrink-0">
          <div class="flex items-center gap-2">
            <div class="p-2 rounded-xl bg-primary/10 text-primary">
              <Wrench class="w-5 h-5" />
            </div>
            <div>
              <DialogTitle class="text-lg font-bold">批量设置任务参数</DialogTitle>
              <DialogDescription class="text-xs text-muted-foreground mt-0.5">
                仅勾选开启的字段会统一更新；针对多类型混选，系统已自动计算生效范围并精准分发。
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <!-- 目标任务提示与混选信息条 -->
        <div class="px-6 py-2.5 bg-muted/30 border-b border-border/40 flex items-center justify-between gap-3 text-xs">
          <div class="flex items-center gap-2 min-w-0">
            <span class="text-muted-foreground shrink-0">目标范围:</span>
            <Badge variant="secondary" class="font-mono text-xs font-semibold px-2">
              {{ isQueryMode ? '当前筛选条件匹配任务' : `${totalCount} 个选中任务` }}
            </Badge>
            <Badge v-if="isMixedTypes" variant="outline" class="text-[10px] text-amber-500 bg-amber-500/10 border-amber-500/20 leading-none">
              混合类型
            </Badge>
            <div class="hidden sm:flex items-center gap-1.5 overflow-hidden text-[11px] text-muted-foreground">
              <span v-if="typeStats['task']" class="shrink-0">• {{ typeStats['task'] }} 脚本任务</span>
              <span v-if="typeStats['repo']" class="shrink-0">• {{ typeStats['repo'] }} 仓库同步</span>
              <span v-if="typeStats['app']" class="shrink-0">• {{ typeStats['app'] }} 应用实体</span>
              <span v-if="typeStats['agent_sync_script']" class="shrink-0">• {{ typeStats['agent_sync_script'] }} Agent同步</span>
            </div>
          </div>
          <div class="shrink-0 text-primary font-medium text-[11px] flex items-center gap-1.5">
            <span>已选修改项:</span>
            <strong class="font-mono text-xs px-1.5 py-0.2 bg-primary/10 text-primary rounded">{{ activeCount }}</strong>
          </div>
        </div>

        <!-- 分类切换过滤药丸 -->
        <div v-if="hasCategoryTabs" class="px-6 pt-3 pb-1 flex items-center gap-2 shrink-0 border-b border-border/20 bg-muted/10">
          <span class="text-[11px] text-muted-foreground shrink-0 font-medium">参数分类:</span>
          <div class="flex items-center gap-1.5">
            <button
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all"
              :class="activeCategoryTab === 'all' ? 'bg-primary text-primary-foreground shadow-xs' : 'bg-muted/50 text-muted-foreground hover:bg-muted'"
              @click="activeCategoryTab = 'all'"
            >
              全部参数
            </button>
            <button
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all"
              :class="activeCategoryTab === 'common' ? 'bg-primary text-primary-foreground shadow-xs' : 'bg-muted/50 text-muted-foreground hover:bg-muted'"
              @click="activeCategoryTab = 'common'"
            >
              通用参数
            </button>
            <button
              v-if="scopeLanguages.isSupported"
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all"
              :class="activeCategoryTab === 'language' ? 'bg-primary text-primary-foreground shadow-xs' : 'bg-muted/50 text-muted-foreground hover:bg-muted'"
              @click="activeCategoryTab = 'language'"
            >
              语言环境
            </button>
            <button
              v-if="scopeAgent.isSupported"
              type="button"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all"
              :class="activeCategoryTab === 'script_only' ? 'bg-primary text-primary-foreground shadow-xs' : 'bg-muted/50 text-muted-foreground hover:bg-muted'"
              @click="activeCategoryTab = 'script_only'"
            >
              脚本专属
            </button>
          </div>
        </div>

        <!-- 参数配置核心列表 -->
        <ScrollArea class="flex-1 min-h-0 px-6 py-4">
          <div class="space-y-4 pb-6">

            <!-- 1. 执行超时时间 (Timeout) - 通用 -->
            <div
              v-if="scopeTimeout.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'common')"
              class="p-3.5 rounded-xl border transition-all"
              :class="applyTimeout ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch v-model="applyTimeout" id="apply-timeout" />
                  <Label for="apply-timeout" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <Clock class="w-3.5 h-3.5 text-primary" />
                    <span>{{ PARAM_DEFINITIONS.timeout.label }}</span>
                  </Label>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <Badge v-if="scopeTimeout.isFull" variant="outline" class="text-[9px] font-mono px-1.5 py-0.5 bg-emerald-500/10 text-emerald-500 border-emerald-500/20 leading-none">
                    全部适用
                  </Badge>
                  <span class="text-[10px] text-muted-foreground">{{ applyTimeout ? '已激活' : '不修改' }}</span>
                </div>
              </div>
              <div v-if="applyTimeout" class="mt-3 pt-3 border-t border-border/40 pl-8 space-y-2 animate-in fade-in duration-200">
                <div class="flex items-center gap-3">
                  <Input v-model="timeoutVal" type="number" min="1" max="1440" class="w-32 h-8 text-xs bg-background" />
                  <span class="text-[11px] text-muted-foreground">单次运行的最长允许时间（默认 30 分钟，最大 1440）</span>
                </div>
              </div>
            </div>

            <!-- 2. 标签设置 (Tags) - 通用 -->
            <div
              v-if="scopeTags.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'common')"
              class="p-3.5 rounded-xl border transition-all"
              :class="applyTags ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch v-model="applyTags" id="apply-tags" />
                  <Label for="apply-tags" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <Tag class="w-3.5 h-3.5 text-primary" />
                    <span>{{ PARAM_DEFINITIONS.tags.label }}</span>
                  </Label>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <Badge v-if="scopeTags.isFull" variant="outline" class="text-[9px] font-mono px-1.5 py-0.5 bg-emerald-500/10 text-emerald-500 border-emerald-500/20 leading-none">
                    全部适用
                  </Badge>
                  <span class="text-[10px] text-muted-foreground">{{ applyTags ? '已激活' : '不修改' }}</span>
                </div>
              </div>
              <div v-if="applyTags" class="mt-3 pt-3 border-t border-border/40 pl-8 space-y-3 animate-in fade-in duration-200">
                <div class="flex items-center gap-4 text-xs">
                  <span class="text-muted-foreground text-[11px]">合并策略:</span>
                  <RadioGroup v-model="tagMode" class="flex items-center gap-4">
                    <div class="flex items-center space-x-1.5">
                      <RadioGroupItem value="append" id="tag-append" />
                      <Label for="tag-append" class="text-xs font-normal cursor-pointer">追加标签 (保留原有标签)</Label>
                    </div>
                    <div class="flex items-center space-x-1.5">
                      <RadioGroupItem value="replace" id="tag-replace" />
                      <Label for="tag-replace" class="text-xs font-normal cursor-pointer">全量覆盖 (清空并重置为新标签)</Label>
                    </div>
                  </RadioGroup>
                </div>
                <TaskTagsConfig v-model="tagsVal" />
              </div>
            </div>

            <!-- 3. 失败重试与随机延迟 (Retry Policy) - 通用 -->
            <div
              v-if="scopeRetry.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'common')"
              class="p-3.5 rounded-xl border transition-all"
              :class="applyRetry ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch v-model="applyRetry" id="apply-retry" />
                  <Label for="apply-retry" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <RotateCcw class="w-3.5 h-3.5 text-primary" />
                    <span>{{ PARAM_DEFINITIONS.retry.label }}</span>
                  </Label>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <Badge v-if="scopeRetry.isFull" variant="outline" class="text-[9px] font-mono px-1.5 py-0.5 bg-emerald-500/10 text-emerald-500 border-emerald-500/20 leading-none">
                    全部适用
                  </Badge>
                  <span class="text-[10px] text-muted-foreground">{{ applyRetry ? '已激活' : '不修改' }}</span>
                </div>
              </div>
              <div v-if="applyRetry" class="mt-3 pt-3 border-t border-border/40 pl-8 space-y-3 animate-in fade-in duration-200">
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <div class="space-y-1">
                    <Label class="text-[11px] text-muted-foreground">重试次数</Label>
                    <Input v-model="retryCountVal" type="number" min="0" max="10" placeholder="0" class="h-8 text-xs bg-background" />
                  </div>
                  <div class="space-y-1">
                    <Label class="text-[11px] text-muted-foreground">重试间隔 (秒)</Label>
                    <Input v-model="retryIntervalVal" type="number" min="0" placeholder="0" class="h-8 text-xs bg-background" />
                  </div>
                  <div class="space-y-1">
                    <Label class="text-[11px] text-muted-foreground">随机延迟上限 (秒)</Label>
                    <Input v-model="randomRangeVal" type="number" min="0" placeholder="0" class="h-8 text-xs bg-background" />
                  </div>
                </div>
              </div>
            </div>

            <!-- 4. 全量环境变量注入开关 - 通用 -->
            <div
              v-if="scopeEnvs.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'common')"
              class="p-3.5 rounded-xl border transition-all"
              :class="applyEnvs ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch v-model="applyEnvs" id="apply-envs" />
                  <Label for="apply-envs" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <AlertCircle class="w-3.5 h-3.5 text-primary" />
                    <span>{{ PARAM_DEFINITIONS.envs.label }}</span>
                  </Label>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <Badge v-if="scopeEnvs.isFull" variant="outline" class="text-[9px] font-mono px-1.5 py-0.5 bg-emerald-500/10 text-emerald-500 border-emerald-500/20 leading-none">
                    全部适用
                  </Badge>
                  <span class="text-[10px] text-muted-foreground">{{ applyEnvs ? '已激活' : '不修改' }}</span>
                </div>
              </div>
              <div v-if="applyEnvs" class="mt-3 pt-3 border-t border-border/40 pl-8 space-y-2 animate-in fade-in duration-200">
                <div class="flex items-center gap-3">
                  <Switch v-model="allEnvsVal" id="batch-all-envs" />
                  <Label for="batch-all-envs" class="text-xs cursor-pointer font-medium">统一开启全量环境变量自动注入</Label>
                </div>
              </div>
            </div>

            <!-- 5. 运行时语言环境 (Mise) - 语言环境独立分类 (支持脚本、仓库与应用) -->
            <div
              v-if="scopeLanguages.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'language')"
              class="p-3.5 rounded-xl border transition-all"
              :class="applyLanguages ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch v-model="applyLanguages" id="apply-langs" />
                  <Label for="apply-langs" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <Sparkles class="w-3.5 h-3.5 text-primary" />
                    <span>{{ PARAM_DEFINITIONS.languages.label }}</span>
                  </Label>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <!-- 范围徽章：全员适用 vs 部分任务生效 -->
                  <Badge
                    variant="outline"
                    class="text-[9px] font-mono px-1.5 py-0.5 leading-none"
                    :class="scopeLanguages.isFull ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' : 'bg-amber-500/10 text-amber-500 border-amber-500/25'"
                  >
                    {{ scopeLanguages.badgeText }}
                  </Badge>
                  <span class="text-[10px] text-muted-foreground">{{ applyLanguages ? '已激活' : '不修改' }}</span>
                </div>
              </div>
              <div v-if="applyLanguages" class="mt-3 pt-3 border-t border-border/40 pl-8 space-y-2 animate-in fade-in duration-200">
                <div v-if="!scopeLanguages.isFull" class="p-2 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 text-[11px] flex items-center gap-1.5">
                  <Sparkles class="w-3.5 h-3.5 shrink-0" />
                  <span>已启用混选保护：多语言环境仅支持脚本、仓库和应用任务 (共 {{ scopeLanguages.applicableCount }} 项)，Agent 同步任务将自动跳过。</span>
                </div>
                <p class="text-[11px] text-muted-foreground">批量将目标任务的运行环境替换为指定的运行时版本 (支持普通脚本、仓库同步和应用实体)</p>
                <TaskLangConfig v-model="selectedLangs" />
              </div>
            </div>

            <!-- 6. 执行位置 (Agent 节点) - 脚本专属 (仅普通脚本任务生效) -->
            <div
              v-if="scopeAgent.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'script_only')"
              class="p-3.5 rounded-xl border transition-all"
              :class="applyAgent ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch v-model="applyAgent" id="apply-agent" />
                  <Label for="apply-agent" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <Server class="w-3.5 h-3.5 text-primary" />
                    <span>{{ PARAM_DEFINITIONS.agent.label }}</span>
                  </Label>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <Badge
                    variant="outline"
                    class="text-[9px] font-mono px-1.5 py-0.5 leading-none"
                    :class="scopeAgent.isFull ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' : 'bg-amber-500/10 text-amber-500 border-amber-500/25'"
                  >
                    {{ scopeAgent.badgeText }}
                  </Badge>
                  <span class="text-[10px] text-muted-foreground">{{ applyAgent ? '已激活' : '不修改' }}</span>
                </div>
              </div>
              <div v-if="applyAgent" class="mt-3 pt-3 border-t border-border/40 pl-8 space-y-2 animate-in fade-in duration-200">
                <div v-if="!scopeAgent.isFull" class="p-2 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 text-[11px] flex items-center gap-1.5">
                  <AlertCircle class="w-3.5 h-3.5 shrink-0" />
                  <span>已启用混选保护：统一执行位置仅对普通脚本任务生效 (共 {{ scopeAgent.applicableCount }} 项)，选中的应用实体或仓库同步将自动跳过。</span>
                </div>
                <div class="flex items-center gap-3">
                  <Select v-model="selectedAgentId">
                    <SelectTrigger class="w-56 h-8 text-xs bg-background">
                      <SelectValue placeholder="选择执行节点..." />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="local" class="text-xs">
                        <div class="flex items-center gap-2">
                          <Laptop class="w-3.5 h-3.5 text-blue-500" />
                          <span>本地面板服务</span>
                        </div>
                      </SelectItem>
                      <SelectItem v-for="agent in agents" :key="agent.id" :value="String(agent.id)" class="text-xs">
                        <div class="flex items-center gap-2">
                          <div class="w-1.5 h-1.5 rounded-full" :class="agent.status === 'online' ? 'bg-green-500' : 'bg-muted-foreground'" />
                          <span>{{ agent.name }}</span>
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <span class="text-[11px] text-muted-foreground">将普通脚本任务一键重定向至指定的本地或远程 Agent 执行</span>
                </div>
              </div>
            </div>

            <!-- 7. 前置与后置 CLI 命令 (暂时注释) -->
            <!--
            <div
              v-if="scopeCommands?.isSupported && (activeCategoryTab === 'all' || activeCategoryTab === 'script_only')"
              class="p-3.5 rounded-xl border transition-all"
              :class="(applyPreCmd || applyPostCmd) ? 'bg-primary/[0.02] border-primary/40 shadow-xs' : 'bg-muted/10 border-border/60 opacity-80'"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-2.5 min-w-0">
                  <Switch :model-value="applyPreCmd || applyPostCmd" @update:model-value="(val) => { applyPreCmd = val; applyPostCmd = val; }" id="apply-cmds" />
                  <Label for="apply-cmds" class="text-xs font-bold cursor-pointer select-none flex items-center gap-1.5">
                    <Terminal class="w-3.5 h-3.5 text-primary" />
                    <span>前置 / 后置指令钩子 (CLI Hooks)</span>
                  </Label>
                </div>
              </div>
            </div>
            -->

          </div>
        </ScrollArea>

        <!-- 弹窗底栏 -->
        <div class="px-6 py-3.5 border-t border-border/40 bg-muted/20 shrink-0 flex items-center justify-between gap-4">
          <div class="text-xs text-muted-foreground">
            <template v-if="activeCount > 0">
              已启用 <strong class="text-primary">{{ activeCount }}</strong> 项修改，保存时将自动按任务类型精准生效
            </template>
            <template v-else>
              尚未开启任何需要修改的参数
            </template>
          </div>
          <div class="flex items-center gap-2">
            <Button variant="ghost" size="sm" :disabled="loading" @click="emit('update:open', false)">取消</Button>
            <Button
              size="sm"
              class="font-semibold px-5 bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
              :disabled="loading || activeCount === 0"
              @click="handleSave"
            >
              <Loader2 v-if="loading" class="w-3.5 h-3.5 mr-1.5 animate-spin" />
              <span>{{ loading ? '正在保存...' : '应用批量设置' }}</span>
            </Button>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
