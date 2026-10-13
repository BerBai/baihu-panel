<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { api } from '@/api'
import { toast } from 'vue-sonner'
import { copyToClipboard } from '@/utils/clipboard'
import {
  Loader2,
  Save,
  Fingerprint,
  Globe,
  KeyRound,
  ShieldCheck,
  User,
  Copy,
  Check,
  Eye,
  EyeOff,
  ExternalLink,
  AlertTriangle,
  Sparkles,
  HelpCircle,
  Lock,
  RotateCcw
} from 'lucide-vue-next'

const loading = ref(false)
const saving = ref(false)
const demoMode = ref(false)
const showSecret = ref(false)
const copied = ref(false)
const showGuide = ref(false)

const form = ref({
  enabled: false,
  issuer: '',
  client_id: '',
  client_secret: '',
  redirect_url: '',
  scopes: '',
  auto_create: false,
  default_role: 'user',
})

// 系统自动推导的标准回调地址
const defaultRedirectUrl = computed(() => {
  if (typeof window === 'undefined') return ''
  const base = window.location.origin
  const prefix = (window as any).__BASE_URL__ || ''
  return `${base}${prefix}/api/v1/auth/oidc/callback`
})

// 当前生效的回调地址（用户自定义或默认）
const activeRedirectUrl = computed(() => {
  return form.value.redirect_url.trim() || defaultRedirectUrl.value
})

// 探测文档地址
const discoveryUrl = computed(() => {
  const iss = form.value.issuer.trim()
  if (!iss || (!iss.startsWith('http://') && !iss.startsWith('https://'))) return ''
  return iss.replace(/\/+$/, '') + '/.well-known/openid-configuration'
})

// 常见快捷 Scope 列表
const commonScopes = ['groups', 'roles', 'offline_access']

function addScope(scope: string) {
  const current = form.value.scopes.trim()
  const list = current ? current.split(/[\s,]+/).filter(Boolean) : []
  if (!list.includes(scope)) {
    list.push(scope)
    form.value.scopes = list.join(' ')
  }
}

async function copyRedirectUrl() {
  if (!activeRedirectUrl.value) return
  const success = await copyToClipboard(activeRedirectUrl.value)
  if (success) {
    copied.value = true
    toast.success('回调地址已复制到剪贴板')
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } else {
    toast.error('复制失败，请手动选择复制')
  }
}

function resetDefaultRedirectUrl() {
  form.value.redirect_url = ''
  toast.info('已重置为系统默认回调地址')
}

async function loadSection() {
  loading.value = true
  try {
    const res = await api.settings.getSection('security')
    form.value = {
      enabled: res.oidc_enabled === 'true',
      issuer: res.oidc_issuer || '',
      client_id: res.oidc_client_id || '',
      client_secret: res.oidc_client_secret || '',
      redirect_url: res.oidc_redirect_url || '',
      scopes: res.oidc_scopes || '',
      auto_create: res.oidc_auto_create === 'true',
      default_role: res.oidc_default_role === 'admin' ? 'admin' : 'user',
    }
  } catch (e: any) {
    toast.error(e.message || '读取 OIDC 配置失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  if (demoMode.value) {
    toast.error('演示模式下禁止修改 OIDC 配置')
    return
  }

  if (form.value.enabled && (!form.value.issuer || !form.value.client_id)) {
    toast.error('启用 OIDC 单点登录前，必须填写 Issuer 与 Client ID')
    return
  }

  saving.value = true
  try {
    await api.settings.setSection('security', {
      oidc_enabled: form.value.enabled ? 'true' : 'false',
      oidc_issuer: form.value.issuer.trim(),
      oidc_client_id: form.value.client_id.trim(),
      oidc_client_secret: form.value.client_secret,
      oidc_redirect_url: form.value.redirect_url.trim(),
      oidc_scopes: form.value.scopes.trim(),
      oidc_auto_create: form.value.auto_create ? 'true' : 'false',
      oidc_default_role: form.value.default_role,
    })
    toast.success('OIDC 单点登录配置已保存')
    await loadSection()
  } catch (e: any) {
    toast.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function loadDemoMode() {
  try {
    const res = await api.settings.getPublicSite()
    demoMode.value = !!res.demo_mode
  } catch {
    demoMode.value = false
  }
}

onMounted(() => {
  loadSection()
  loadDemoMode()
})
</script>

<template>
  <Card class="shadow-sm">
    <!-- 头部：标题、状态与右上角集成开关 -->
    <CardHeader class="pb-3.5 space-y-2">
      <div class="flex items-center justify-between gap-3">
        <div class="flex items-center gap-2.5 min-w-0">
          <div class="p-1.5 rounded-lg bg-sky-500/10 text-sky-600 dark:text-sky-400 shrink-0">
            <Fingerprint class="w-4 h-4" />
          </div>
          <CardTitle class="text-base font-semibold flex items-center gap-2 truncate">
            <span>OIDC 单点登录</span>
            <Badge
              v-if="demoMode"
              variant="destructive"
              class="text-[10px] px-1.5 py-0 font-normal shrink-0"
            >
              演示模式禁用
            </Badge>
            <Badge
              v-else-if="form.enabled"
              variant="secondary"
              class="text-[10px] px-1.5 py-0 font-normal bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 shrink-0"
            >
              已开启
            </Badge>
          </CardTitle>
        </div>

        <!-- 右上角：帮助与主开关 -->
        <div class="flex items-center gap-3 shrink-0">
          <button
            type="button"
            @click="showGuide = !showGuide"
            class="hidden sm:flex items-center gap-1 text-xs text-sky-600 dark:text-sky-400 hover:underline"
          >
            <HelpCircle class="w-3.5 h-3.5" />
            <span>{{ showGuide ? '收起指引' : '对接参考' }}</span>
          </button>
          <div class="flex items-center gap-2 pl-3 border-l border-border/60">
            <Switch
              v-model="form.enabled"
              id="oidc-enabled"
              :disabled="demoMode || loading"
            />
            <Label for="oidc-enabled" class="text-xs cursor-pointer select-none">
              {{ form.enabled ? '已启用' : '开启' }}
            </Label>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-between gap-2 text-xs text-muted-foreground">
        <span class="truncate">接入 Pocket ID、Keycloak、Authentik 等标准 OpenID Connect 身份源实现免密登录</span>
        <button
          type="button"
          @click="showGuide = !showGuide"
          class="sm:hidden flex items-center gap-1 text-xs text-sky-600 dark:text-sky-400 hover:underline shrink-0"
        >
          <HelpCircle class="w-3.5 h-3.5" />
          <span>{{ showGuide ? '收起' : '参考' }}</span>
        </button>
      </div>

      <!-- 演示模式警示条 -->
      <div
        v-if="demoMode"
        class="flex items-center gap-2 p-2 rounded-lg bg-destructive/10 text-destructive text-xs"
      >
        <AlertTriangle class="h-3.5 w-3.5 shrink-0" />
        <span>演示模式下已锁定 OIDC 访问凭据与安全参数配置。</span>
      </div>

      <!-- 对接指引折叠面板 -->
      <div
        v-if="showGuide"
        class="p-3 rounded-xl border bg-gradient-to-br from-sky-500/5 via-background to-purple-500/5 border-sky-500/25 space-y-2 text-xs"
      >
        <div class="flex items-center gap-1.5 font-semibold text-foreground">
          <Sparkles class="w-3.5 h-3.5 text-sky-500" />
          <span>常见 IdP 对接参考速查</span>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[11px] text-muted-foreground">
          <div class="p-1.5 px-2.5 rounded-lg bg-muted/40 border border-border/40">
            <b class="text-foreground">Pocket ID:</b> Issuer 填 <code class="text-sky-600 dark:text-sky-400">https://id.domain.com</code>，Scope 留空
          </div>
          <div class="p-1.5 px-2.5 rounded-lg bg-muted/40 border border-border/40">
            <b class="text-foreground">Authentik:</b> Issuer 填 <code class="text-sky-600 dark:text-sky-400">https://&lt;host&gt;/application/o/&lt;slug&gt;/</code>
          </div>
          <div class="p-1.5 px-2.5 rounded-lg bg-muted/40 border border-border/40">
            <b class="text-foreground">Keycloak:</b> Issuer 填 <code class="text-sky-600 dark:text-sky-400">https://&lt;host&gt;/realms/&lt;realm&gt;</code>
          </div>
          <div class="p-1.5 px-2.5 rounded-lg bg-muted/40 border border-border/40">
            <b class="text-foreground">重定向 URI:</b> 务必在 IdP 客户端中填入下方右侧的<b>回调地址</b>
          </div>
        </div>
      </div>
    </CardHeader>

    <CardContent class="space-y-3.5 pt-0">
      <!-- 行 1：端点网络配置（双列对称：Issuer + 回调地址） -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3.5">
        <!-- 发现地址 -->
        <div class="space-y-1.5">
          <div class="flex items-center justify-between">
            <Label class="text-xs font-medium text-foreground flex items-center gap-1">
              <Globe class="w-3.5 h-3.5 text-muted-foreground" />
              <span>Issuer（发现文档地址）</span>
              <span class="text-destructive">*</span>
            </Label>
            <a
              v-if="discoveryUrl"
              :href="discoveryUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-[11px] text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-0.5"
            >
              <span>校验</span>
              <ExternalLink class="w-2.5 h-2.5" />
            </a>
          </div>
          <Input
            v-model="form.issuer"
            placeholder="https://id.example.com"
            class="h-9 text-xs font-mono"
            :disabled="demoMode || loading"
          />
        </div>

        <!-- 回调地址 (Redirect URI) 带内嵌复制 -->
        <div class="space-y-1.5">
          <div class="flex items-center justify-between">
            <Label class="text-xs font-medium text-foreground flex items-center gap-1">
              <span>回调地址（Redirect URI）</span>
              <span class="text-[10px] text-muted-foreground font-normal">需在 IdP 严格登记</span>
            </Label>
            <button
              v-if="form.redirect_url"
              type="button"
              @click="resetDefaultRedirectUrl"
              class="text-[11px] text-muted-foreground hover:text-foreground flex items-center gap-0.5"
              title="重置为默认自动推导地址"
            >
              <RotateCcw class="w-2.5 h-2.5" />
              <span>默认</span>
            </button>
          </div>
          <div class="flex items-center gap-1.5">
            <Input
              v-model="form.redirect_url"
              :placeholder="defaultRedirectUrl"
              class="h-9 text-xs font-mono flex-1 min-w-0"
              :disabled="demoMode || loading"
            />
            <Button
              type="button"
              variant="outline"
              size="sm"
              class="h-9 px-2.5 shrink-0 text-xs gap-1 font-medium"
              @click="copyRedirectUrl"
              :title="'复制生效回调地址'"
            >
              <Check v-if="copied" class="w-3.5 h-3.5 text-emerald-500" />
              <Copy v-else class="w-3.5 h-3.5" />
              <span>{{ copied ? '已复制' : '复制' }}</span>
            </Button>
          </div>
        </div>
      </div>

      <!-- 行 2：客户端凭据（双列对称：Client ID + Client Secret） -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3.5">
        <div class="space-y-1.5">
          <Label class="text-xs font-medium text-foreground flex items-center gap-1">
            <KeyRound class="w-3.5 h-3.5 text-muted-foreground" />
            <span>Client ID（客户端标识）</span>
            <span class="text-destructive">*</span>
          </Label>
          <Input
            v-model="form.client_id"
            placeholder="例如 baihu-panel"
            class="h-9 text-xs font-mono"
            :disabled="demoMode || loading"
          />
        </div>

        <div class="space-y-1.5">
          <Label class="text-xs font-medium text-foreground flex items-center gap-1">
            <Lock class="w-3.5 h-3.5 text-muted-foreground" />
            <span>Client Secret（客户端密钥）</span>
          </Label>
          <div class="relative">
            <Input
              v-model="form.client_secret"
              :type="showSecret ? 'text' : 'password'"
              class="h-9 pr-8 text-xs font-mono"
              :disabled="demoMode || loading"
              placeholder="留空保持不变"
            />
            <button
              type="button"
              @click="showSecret = !showSecret"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors p-0.5"
              :title="showSecret ? '隐藏密钥' : '显示明文'"
            >
              <EyeOff v-if="showSecret" class="w-3.5 h-3.5" />
              <Eye v-else class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      <!-- 行 3：附加 Scope -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between">
          <Label class="text-xs font-medium text-foreground">
            附加 Scope <span class="text-[10px] text-muted-foreground font-normal">（默认已包含 openid profile email）</span>
          </Label>
          <div class="flex items-center gap-1">
            <button
              v-for="s in commonScopes"
              :key="s"
              type="button"
              @click="addScope(s)"
              :disabled="demoMode || loading"
              class="text-[10px] px-1.5 py-0.5 rounded bg-muted hover:bg-muted/80 text-foreground border border-border/50 transition-colors disabled:opacity-50"
            >
              +{{ s }}
            </button>
          </div>
        </div>
        <Input
          v-model="form.scopes"
          placeholder="例如 groups 或 roles，多项以空格分隔"
          class="h-9 text-xs font-mono"
          :disabled="demoMode || loading"
        />
      </div>

      <!-- 行 4：建号策略与默认角色（单行两端紧凑整合） -->
      <div class="p-2.5 px-3 rounded-lg border bg-muted/20 flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
        <div class="flex items-center gap-2.5">
          <Switch
            v-model="form.auto_create"
            id="auto-create"
            :disabled="demoMode || loading"
          />
          <div>
            <Label for="auto-create" class="text-xs font-medium cursor-pointer select-none">首次登录自动建号</Label>
            <p class="text-[10px] text-muted-foreground">未关联本地账号的陌生身份首次登录时自动开通</p>
          </div>
        </div>

        <!-- 角色切换器（行内 Segmented Control） -->
        <div v-if="form.auto_create" class="flex items-center gap-2 self-start sm:self-center">
          <span class="text-[11px] text-muted-foreground shrink-0">默认角色:</span>
          <div class="inline-flex rounded-lg border bg-background p-0.5 text-xs shadow-xs">
            <button
              type="button"
              @click="!demoMode && (form.default_role = 'user')"
              class="px-2.5 py-1 rounded-md transition-all flex items-center gap-1 text-xs select-none"
              :class="form.default_role === 'user' ? 'bg-primary text-primary-foreground font-medium shadow-xs' : 'text-muted-foreground hover:text-foreground'"
            >
              <User class="w-3.5 h-3.5" />
              <span>普通用户</span>
            </button>
            <button
              type="button"
              @click="!demoMode && (form.default_role = 'admin')"
              class="px-2.5 py-1 rounded-md transition-all flex items-center gap-1 text-xs select-none"
              :class="form.default_role === 'admin' ? 'bg-amber-500 text-white font-medium shadow-xs' : 'text-muted-foreground hover:text-foreground'"
            >
              <ShieldCheck class="w-3.5 h-3.5" />
              <span>管理员</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 底栏：保存按钮 -->
      <div class="flex items-center justify-between pt-2 border-t border-border/50">
        <span class="text-xs text-muted-foreground">
          配置保存后即时在登录页生效
        </span>
        <Button
          @click="save"
          :disabled="saving || loading || demoMode"
          class="h-8 px-4 text-xs font-medium gap-1.5 shadow-sm"
        >
          <Loader2 v-if="saving" class="h-3.5 w-3.5 animate-spin" />
          <Save v-else class="h-3.5 w-3.5" />
          <span>保存配置</span>
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
