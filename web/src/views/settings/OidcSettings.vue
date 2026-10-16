<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
  Zap,
  Layers,
  ChevronDown,
  ChevronUp
} from 'lucide-vue-next'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const demoMode = ref(false)
const showSecret = ref(false)
const copied = ref(false)
const showGuide = ref(false)
const configExpanded = ref(false)
const savedEnabled = ref(false)

const form = ref({
  enabled: false,
  type: 'oidc', // 'oidc' | 'oauth2'
  issuer: '',
  client_id: '',
  client_secret: '',
  scopes: '',
  auto_create: false,
  default_role: 'user',
  // 通用 OAuth 2.0
  oauth2_auth_url: '',
  oauth2_token_url: '',
  oauth2_userinfo_url: '',
  oauth2_name: '',
  oauth2_user_key: '',
})

// 系统自动推导的标准回调地址（供用户在第三方后台登记）
const defaultRedirectUrl = computed(() => {
  if (typeof window === 'undefined') return ''
  const base = window.location.origin
  const prefix = (window as any).__BASE_URL__ || ''
  return `${base}${prefix}/api/v1/auth/oidc/callback`
})

function cleanIssuer() {
  if (form.value.issuer) {
    form.value.issuer = form.value.issuer
      .trim()
      .replace(/\/+\.well-known\/openid-configuration\/?$/i, '')
      .replace(/\/+$/, '')
  }
}

// 探测文档地址（OIDC 模式）
const discoveryUrl = computed(() => {
  if (form.value.type !== 'oidc') return ''
  let iss = form.value.issuer.trim()
  if (!iss || (!iss.startsWith('http://') && !iss.startsWith('https://'))) return ''
  iss = iss.replace(/\/+\.well-known\/openid-configuration\/?$/i, '').replace(/\/+$/, '')
  return `${iss}/.well-known/openid-configuration`
})

// 常用 OAuth 2.0 平台预设模版
interface OAuth2Preset {
  name: string
  displayName: string
  authUrl: string
  tokenUrl: string
  userinfoUrl: string
  scopes: string
  userKey?: string
}

const oauth2Presets: OAuth2Preset[] = [
  {
    name: 'GitHub',
    displayName: 'GitHub',
    authUrl: 'https://github.com/login/oauth/authorize',
    tokenUrl: 'https://github.com/login/oauth/access_token',
    userinfoUrl: 'https://api.github.com/user',
    scopes: 'read:user user:email',
    userKey: 'id',
  },
  {
    name: 'Gitee (码云)',
    displayName: 'Gitee',
    authUrl: 'https://gitee.com/oauth/authorize',
    tokenUrl: 'https://gitee.com/oauth/token',
    userinfoUrl: 'https://gitee.com/api/v5/user',
    scopes: 'user_info emails',
    userKey: 'id',
  },
  {
    name: 'GitLab',
    displayName: 'GitLab',
    authUrl: 'https://gitlab.com/oauth/authorize',
    tokenUrl: 'https://gitlab.com/oauth/token',
    userinfoUrl: 'https://gitlab.com/api/v4/user',
    scopes: 'read_user',
    userKey: 'id',
  },
]

function applyPreset(p: OAuth2Preset) {
  form.value.oauth2_name = p.displayName
  form.value.oauth2_auth_url = p.authUrl
  form.value.oauth2_token_url = p.tokenUrl
  form.value.oauth2_userinfo_url = p.userinfoUrl
  form.value.scopes = p.scopes
  if (p.userKey) form.value.oauth2_user_key = p.userKey
  toast.success(`已自动填充 ${p.name} 端点配置`)
}

// 常见快捷 Scope 列表
const commonScopes = computed(() => {
  return form.value.type === 'oauth2'
    ? ['read:user', 'user:email', 'profile', 'email']
    : ['groups', 'roles', 'offline_access']
})

function addScope(scope: string) {
  const current = form.value.scopes.trim()
  const list = current ? current.split(/[\s,]+/).filter(Boolean) : []
  if (!list.includes(scope)) {
    list.push(scope)
    form.value.scopes = list.join(' ')
  }
}

async function copyRedirectUrl() {
  if (!defaultRedirectUrl.value) return
  const success = await copyToClipboard(defaultRedirectUrl.value)
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

async function loadSection() {
  loading.value = true
  try {
    const res = await api.settings.getSection('security')
    const isEnabled = res.oidc_enabled === 'true'
    savedEnabled.value = isEnabled
    form.value = {
      enabled: isEnabled,
      type: res.oidc_type === 'oauth2' ? 'oauth2' : 'oidc',
      issuer: res.oidc_issuer || '',
      client_id: res.oidc_client_id || '',
      client_secret: res.oidc_client_secret || '',
      scopes: res.oidc_scopes || '',
      auto_create: res.oidc_auto_create === 'true',
      default_role: res.oidc_default_role === 'admin' ? 'admin' : 'user',
      oauth2_auth_url: res.oauth2_auth_url || '',
      oauth2_token_url: res.oauth2_token_url || '',
      oauth2_userinfo_url: res.oauth2_userinfo_url || '',
      oauth2_name: res.oauth2_name || '',
      oauth2_user_key: res.oauth2_user_key || '',
    }
  } catch (e: any) {
    toast.error(e.message || '读取单点登录配置失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  if (demoMode.value) {
    toast.error('演示模式下禁止修改单点登录配置')
    return
  }

  cleanIssuer()

  if (form.value.enabled) {
    if (!form.value.client_id) {
      toast.error('请填写 Client ID（客户端标识）')
      return
    }
    if (form.value.type === 'oidc' && !form.value.issuer) {
      toast.error('OIDC 模式下必须填写 Issuer（发现文档地址）')
      return
    }
    if (form.value.type === 'oauth2' && (!form.value.oauth2_auth_url || !form.value.oauth2_token_url || !form.value.oauth2_userinfo_url)) {
      toast.error('通用 OAuth 2.0 模式下必须完整填写授权、令牌与用户信息端点')
      return
    }
  }

  saving.value = true
  try {
    await api.settings.setSection('security', {
      oidc_enabled: form.value.enabled ? 'true' : 'false',
      oidc_type: form.value.type,
      oidc_issuer: form.value.issuer.trim(),
      oidc_client_id: form.value.client_id.trim(),
      oidc_client_secret: form.value.client_secret,
      oidc_scopes: form.value.scopes.trim(),
      oidc_auto_create: form.value.auto_create ? 'true' : 'false',
      oidc_default_role: form.value.default_role,
      oauth2_auth_url: form.value.oauth2_auth_url.trim(),
      oauth2_token_url: form.value.oauth2_token_url.trim(),
      oauth2_userinfo_url: form.value.oauth2_userinfo_url.trim(),
      oauth2_name: form.value.oauth2_name.trim(),
      oauth2_user_key: form.value.oauth2_user_key.trim(),
    })
    toast.success('单点登录配置已保存')
    await loadSection()
    configExpanded.value = false
  } catch (e: any) {
    toast.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleTestConnection() {
  if (demoMode.value) {
    toast.error('演示模式下禁止测试')
    return
  }
  cleanIssuer()
  if (form.value.type === 'oidc') {
    if (!form.value.issuer.trim()) {
      toast.error('请先填写 Issuer（发现文档地址）')
      return
    }
  } else {
    if (!form.value.oauth2_auth_url || !form.value.oauth2_token_url || !form.value.oauth2_userinfo_url) {
      toast.error('请完整填写授权、令牌与用户信息端点')
      return
    }
  }

  testing.value = true
  try {
    const res = await api.auth.oidcTest({
      type: form.value.type,
      issuer: form.value.issuer.trim(),
      oauth2_auth_url: form.value.oauth2_auth_url.trim(),
      oauth2_token_url: form.value.oauth2_token_url.trim(),
      oauth2_userinfo_url: form.value.oauth2_userinfo_url.trim(),
    })
    toast.success(res.message || '连接测试通过')
  } catch (e: any) {
    toast.error(e.message || '连接测试失败，请检查端点或网络配置')
  } finally {
    testing.value = false
  }
}

function handleToggleEnabled(val: boolean) {
  form.value.enabled = val
  if (val) {
    configExpanded.value = true
    if (!savedEnabled.value) {
      toast.info('已开启开关，请点击下方「保存单点登录配置」按钮以生效')
    }
  }
}

function toggleGuide() {
  showGuide.value = !showGuide.value
  if (showGuide.value) {
    configExpanded.value = true
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
  <Card class="shadow-sm flex flex-col justify-between pt-4 pb-5">
    <!-- 卡片头部：图标、标题、副标题与主开关 -->
    <CardHeader class="pb-3 pt-0 px-4 sm:px-5">
      <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-2.5 sm:gap-4">
        <div class="space-y-1 min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <CardTitle class="text-base flex items-center gap-2">
              <Fingerprint class="w-4 h-4 text-sky-500 shrink-0" />
              <span class="whitespace-nowrap font-bold">统一单点登录 (SSO)</span>
            </CardTitle>
            <Badge
              v-if="demoMode"
              variant="destructive"
              class="text-[10px] px-1.5 py-0 font-normal shrink-0"
            >
              演示模式禁用
            </Badge>
            <Badge
              v-else-if="savedEnabled"
              variant="secondary"
              class="text-[10px] px-1.5 py-0 font-normal bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 shrink-0"
            >
              {{ form.type === 'oauth2' ? (form.oauth2_name || 'OAuth 2.0') : 'OIDC' }} 已激活
            </Badge>
            <Badge
              v-else-if="form.enabled && !savedEnabled"
              variant="secondary"
              class="text-[10px] px-1.5 py-0 font-normal bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20 shrink-0"
            >
              配置待保存
            </Badge>
          </div>
          <CardDescription class="text-xs leading-relaxed pt-0.5">
            集成标准 OIDC 身份中心与通用 OAuth 2.0 平台，实现第三方免密快捷登录与账号审计。
          </CardDescription>
        </div>

        <!-- 操作区：对接指引与主开关 -->
        <div class="flex items-center justify-between sm:justify-end gap-3 shrink-0 pt-1 sm:pt-0">
          <button
            type="button"
            @click="toggleGuide"
            class="flex items-center gap-1 text-xs text-sky-600 dark:text-sky-400 hover:underline transition-colors"
          >
            <HelpCircle class="w-3.5 h-3.5" />
            <span>{{ showGuide ? '收起指引' : '对接指引' }}</span>
          </button>
          <div class="flex items-center gap-2 pl-3 border-l border-border/60">
            <Switch
              :model-value="form.enabled"
              @update:model-value="handleToggleEnabled"
              id="sso-enabled"
              :disabled="demoMode || loading"
            />
            <Label for="sso-enabled" class="text-xs cursor-pointer select-none">
              {{ form.enabled ? '已启用' : '开启' }}
            </Label>
          </div>
        </div>
      </div>
    </CardHeader>

    <CardContent class="flex-1 flex flex-col justify-between px-4 sm:px-5 pb-0 pt-0 space-y-4">
      <div class="space-y-3.5">
        <!-- 演示模式警示条 -->
        <div
          v-if="demoMode"
          class="flex items-center gap-2 p-3 rounded-xl bg-destructive/10 text-destructive text-xs"
        >
          <AlertTriangle class="h-4 w-4 shrink-0" />
          <span>演示模式下已锁定单点登录凭据与安全参数配置。</span>
        </div>

        <!-- 状态指示卡片（与上方卡片保持一致的高级状态面板，集成互联风格折叠/展开入口） -->
        <div
          class="p-3 sm:p-3.5 rounded-xl border flex items-center justify-between transition-all gap-2"
          :class="savedEnabled
            ? 'border-emerald-500/20 bg-emerald-500/5'
            : (form.enabled ? 'border-amber-500/30 bg-amber-500/5' : 'border-border/70 bg-muted/20')"
        >
          <div class="flex items-center gap-2.5 sm:gap-3 min-w-0">
            <div
              class="w-8.5 h-8.5 sm:w-9 sm:h-9 rounded-lg flex items-center justify-center shrink-0"
              :class="savedEnabled
                ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20'
                : (form.enabled ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30' : 'bg-muted text-muted-foreground border border-border/60')"
            >
              <Zap v-if="form.type === 'oauth2'" class="w-4 h-4" />
              <Globe v-else class="w-4 h-4" />
            </div>
            <div class="min-w-0">
              <div class="flex items-center gap-1.5 sm:gap-2 flex-wrap">
                <span class="text-xs font-bold text-foreground truncate">
                  {{ savedEnabled
                    ? (form.type === 'oauth2' ? (form.oauth2_name || '通用 OAuth 2.0') : '标准 OpenID Connect')
                    : (form.enabled ? '单点登录配置待保存' : '单点登录未启用') }}
                </span>
                <Badge
                  variant="outline"
                  class="text-[9px] px-1 py-0 h-4 shrink-0"
                  :class="savedEnabled
                    ? 'border-emerald-500/30 text-emerald-600 dark:text-emerald-400'
                    : (form.enabled ? 'border-amber-500/40 text-amber-600 dark:text-amber-400' : 'border-muted-foreground/30 text-muted-foreground')"
                >
                  {{ form.type === 'oauth2' ? 'OAuth 2.0' : 'OIDC' }}
                </Badge>
              </div>
              <div class="text-[10px] text-muted-foreground pt-0.5 truncate">
                {{ savedEnabled
                  ? (form.auto_create ? '首次登录将根据角色策略自动建号' : '仅允许已绑定存量账号的用户登入')
                  : (form.enabled ? '已修改启用状态，请点击下方「保存单点登录配置」以生效' : (configExpanded ? '请在下方配置第三方认证平台端点参数' : '点击右侧按钮展开配置或开启快捷单点登录')) }}
              </div>
            </div>
          </div>

          <!-- 右侧：在线状态与展开/收起按钮 -->
          <div class="flex items-center gap-2 sm:gap-2.5 shrink-0 pl-1 sm:pl-2">
            <div
              v-if="savedEnabled"
              class="hidden sm:flex items-center gap-1.5 text-[11px] text-emerald-600 dark:text-emerald-400 font-medium"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              <span>通道在线</span>
            </div>
            <div
              v-else-if="form.enabled"
              class="hidden sm:flex items-center gap-1.5 text-[11px] text-amber-600 dark:text-amber-400 font-medium"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
              <span>待保存</span>
            </div>
            <div v-else class="hidden sm:block text-[11px] text-muted-foreground">
              未配置
            </div>

            <Button
              variant="outline"
              size="sm"
              class="text-xs h-7.5 px-2 sm:px-2.5 gap-1 shadow-xs bg-background hover:bg-muted/70 transition-all shrink-0"
              @click="configExpanded = !configExpanded"
            >
              <span class="whitespace-nowrap">{{ configExpanded ? '收起配置' : '配置参数' }}</span>
              <ChevronUp v-if="configExpanded" class="h-3.5 w-3.5" />
              <ChevronDown v-else class="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>

        <!-- 折叠展开的配置区域（像互联模块一样可自由展开/收起） -->
        <div v-if="configExpanded" class="space-y-3.5 pt-1 animate-in fade-in duration-200">
          <!-- 对接指引折叠面板 -->
          <div
            v-if="showGuide"
            class="p-3 rounded-xl border bg-gradient-to-br from-sky-500/5 via-background to-purple-500/5 border-sky-500/25 space-y-2 text-xs"
          >
            <div class="flex items-center gap-1.5 font-semibold text-foreground">
              <Sparkles class="w-3.5 h-3.5 text-sky-500" />
              <span>第三方单点登录对接速查</span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-[11px] text-muted-foreground">
              <div class="p-2 rounded-lg bg-muted/40 border border-border/40">
                <b class="text-foreground">标准 OIDC 模式:</b> Pocket ID、Keycloak、Authentik 等，只需填入 Issuer，发现文档自动嗅探。
              </div>
              <div class="p-2 rounded-lg bg-muted/40 border border-border/40">
                <b class="text-foreground">通用 OAuth 2.0 模式:</b> GitHub、Gitee、GitLab 或自建服务，点击预设按钮即可秒级填充端点。
              </div>
              <div class="p-2 rounded-lg bg-muted/40 border border-border/40 col-span-1 sm:col-span-2">
                <b class="text-foreground">回调重定向地址 (Callback):</b> 必须在第三方应用管理后台严格登记本卡片显示的回调地址。
              </div>
            </div>
          </div>

          <!-- 协议模式切换与快捷预设条（一体化紧凑控制栏） -->
          <div class="p-2.5 sm:p-3 sm:px-3.5 rounded-xl border border-border/70 bg-muted/20 flex flex-col md:flex-row md:items-center justify-between gap-2.5 sm:gap-3">
            <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-3">
              <div class="flex items-center gap-1.5 shrink-0">
                <Layers class="w-4 h-4 text-sky-500 shrink-0" />
                <span class="text-xs font-semibold text-foreground whitespace-nowrap">认证协议:</span>
              </div>
              <div class="grid grid-cols-2 sm:inline-flex rounded-lg border bg-background p-0.5 text-xs shadow-xs w-full sm:w-auto">
                <button
                  type="button"
                  @click="!demoMode && (form.type = 'oidc')"
                  class="px-2.5 py-1.5 sm:py-1 rounded-md transition-all flex items-center justify-center gap-1.5 text-xs select-none"
                  :class="form.type === 'oidc' ? 'bg-primary text-primary-foreground font-medium shadow-xs' : 'text-muted-foreground hover:text-foreground'"
                >
                  <Globe class="w-3.5 h-3.5 shrink-0" />
                  <span class="whitespace-nowrap">标准 OIDC</span>
                </button>
                <button
                  type="button"
                  @click="!demoMode && (form.type = 'oauth2')"
                  class="px-2.5 py-1.5 sm:py-1 rounded-md transition-all flex items-center justify-center gap-1.5 text-xs select-none"
                  :class="form.type === 'oauth2' ? 'bg-primary text-primary-foreground font-medium shadow-xs' : 'text-muted-foreground hover:text-foreground'"
                >
                  <Zap class="w-3.5 h-3.5 shrink-0" />
                  <span class="whitespace-nowrap">通用 OAuth 2.0</span>
                </button>
              </div>
            </div>

            <!-- OAuth2 模式下右侧直接无缝展现快捷预设 -->
            <div v-if="form.type === 'oauth2'" class="flex items-center gap-2 flex-wrap pt-1 sm:pt-0 border-t border-border/40 sm:border-0">
              <span class="text-[11px] text-muted-foreground flex items-center gap-1 shrink-0">
                <Sparkles class="w-3 h-3 text-amber-500" />
                <span>快速套用:</span>
              </span>
              <div class="flex items-center gap-1.5 flex-wrap">
                <button
                  v-for="p in oauth2Presets"
                  :key="p.name"
                  type="button"
                  class="h-7 text-xs px-2.5 rounded-md border border-border/80 bg-background hover:bg-muted/80 hover:border-primary/50 text-foreground transition-all flex items-center gap-1 font-medium disabled:opacity-50"
                  :disabled="demoMode || loading"
                  @click="applyPreset(p)"
                >
                  <span>{{ p.name }}</span>
                </button>
              </div>
            </div>

            <!-- OIDC 模式下右侧展现支持的身份源说明 -->
            <div v-else class="text-[11px] text-muted-foreground flex items-center gap-1 pt-0.5 sm:pt-0">
              <span>支持 Pocket ID / Keycloak / Authentik 等</span>
            </div>
          </div>

          <!-- 全局只读回调地址说明条（白虎面板内置标准接口，无需配置，提供给用户复制登记） -->
          <div class="p-3 rounded-xl border border-sky-500/20 bg-sky-500/5 flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
            <div class="space-y-0.5 min-w-0">
              <div class="flex items-center gap-1.5 text-xs font-semibold text-foreground">
                <ShieldCheck class="w-3.5 h-3.5 text-sky-500 shrink-0" />
                <span>第三方服务商登记专用回调地址 (Callback URI)</span>
              </div>
              <p class="text-[11px] text-muted-foreground truncate">
                白虎面板内置标准化回调路由，无需在本系统配置；请复制并在第三方后台登记为授权回调地址。
              </p>
            </div>
            <div class="flex items-center gap-1.5 shrink-0 w-full sm:w-auto">
              <code class="text-[11px] font-mono px-2 py-1 rounded bg-background border border-border/80 text-foreground truncate max-w-[240px] sm:max-w-[320px]">
                {{ defaultRedirectUrl }}
              </code>
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="h-7.5 px-2.5 text-xs gap-1 font-medium bg-background shadow-xs shrink-0"
                @click="copyRedirectUrl"
              >
                <Check v-if="copied" class="w-3.5 h-3.5 text-emerald-500" />
                <Copy v-else class="w-3.5 h-3.5" />
                <span>{{ copied ? '已复制' : '复制' }}</span>
              </Button>
            </div>
          </div>

          <!-- OIDC 专有模式表单 -->
          <div v-if="form.type === 'oidc'" class="space-y-1.5 pt-0.5">
            <div class="flex items-center justify-between gap-2 flex-wrap">
              <Label class="text-xs font-medium text-foreground flex items-center gap-1.5">
                <Globe class="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                <span>Issuer（提供方根地址）</span>
                <span class="text-destructive">*</span>
              </Label>
              <a
                v-if="discoveryUrl"
                :href="discoveryUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[11px] text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-0.5 shrink-0"
              >
                <span>检验发现文档</span>
                <ExternalLink class="w-2.5 h-2.5" />
              </a>
            </div>
            <Input
              v-model="form.issuer"
              placeholder="https://id.example.com"
              class="h-9 text-xs font-mono"
              :disabled="demoMode || loading"
              @blur="cleanIssuer"
            />
            <p class="text-[11px] text-muted-foreground pt-0.5">
              💡 仅需填写身份认证根域名（例如 <code>https://id.example.com</code>），无需拼接 <code>/.well-known/...</code> 或 <code>/api</code> 后缀。
            </p>
          </div>

          <!-- 通用 OAuth 2.0 专有模式表单 -->
          <div v-else class="space-y-3.5 pt-0.5">
            <!-- 端点三列并排 -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground flex items-center gap-1">
                  <span>授权端点 (Authorize URL)</span>
                  <span class="text-destructive">*</span>
                </Label>
                <Input
                  v-model="form.oauth2_auth_url"
                  placeholder="https://github.com/login/oauth/authorize"
                  class="h-9 text-xs font-mono"
                  :disabled="demoMode || loading"
                />
              </div>

              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground flex items-center gap-1">
                  <span>令牌端点 (Token URL)</span>
                  <span class="text-destructive">*</span>
                </Label>
                <Input
                  v-model="form.oauth2_token_url"
                  placeholder="https://github.com/login/oauth/access_token"
                  class="h-9 text-xs font-mono"
                  :disabled="demoMode || loading"
                />
              </div>

              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground flex items-center gap-1">
                  <span>用户信息端点 (Userinfo URL)</span>
                  <span class="text-destructive">*</span>
                </Label>
                <Input
                  v-model="form.oauth2_userinfo_url"
                  placeholder="https://api.github.com/user"
                  class="h-9 text-xs font-mono"
                  :disabled="demoMode || loading"
                />
              </div>
            </div>

            <!-- 平台标识与用户标识键双列并排 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground">
                  显示名称 <span class="text-[10px] text-muted-foreground font-normal">（登录页按钮回显）</span>
                </Label>
                <Input
                  v-model="form.oauth2_name"
                  placeholder="例如 GitHub 或 Gitee"
                  class="h-9 text-xs"
                  :disabled="demoMode || loading"
                />
              </div>

              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground">
                  唯一用户标识 Key <span class="text-[10px] text-muted-foreground font-normal">（留空智能提取）</span>
                </Label>
                <Input
                  v-model="form.oauth2_user_key"
                  placeholder="例如 id 或 sub"
                  class="h-9 text-xs font-mono"
                  :disabled="demoMode || loading"
                />
              </div>
            </div>
          </div>

          <!-- 全模式通用：客户端凭据（Client ID + Client Secret） -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-3.5 pt-1">
            <div class="space-y-1.5">
              <Label class="text-xs font-medium text-foreground flex items-center gap-1.5">
                <KeyRound class="w-3.5 h-3.5 text-muted-foreground" />
                <span>Client ID（客户端标识）</span>
                <span class="text-destructive">*</span>
              </Label>
              <Input
                v-model="form.client_id"
                placeholder="例如 baihu-panel 或 OAuth App Client ID"
                class="h-9 text-xs font-mono"
                :disabled="demoMode || loading"
              />
            </div>

            <div class="space-y-1.5">
              <Label class="text-xs font-medium text-foreground flex items-center gap-1.5">
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

          <!-- 全模式通用：附加 Scope -->
          <div class="space-y-1.5">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-1.5 sm:gap-2">
              <Label class="text-xs font-medium text-foreground flex items-center gap-1 flex-wrap">
                <span>授权范围 (Scope)</span>
                <span class="text-[10px] text-muted-foreground font-normal">（以空格或逗号分隔）</span>
              </Label>
              <div class="flex items-center gap-1 flex-wrap">
                <button
                  v-for="s in commonScopes"
                  :key="s"
                  type="button"
                  @click="addScope(s)"
                  :disabled="demoMode || loading"
                  class="text-[10px] px-1.5 py-0.5 rounded bg-muted hover:bg-muted/80 text-foreground border border-border/50 transition-colors disabled:opacity-50 shrink-0"
                >
                  +{{ s }}
                </button>
              </div>
            </div>
            <Input
              v-model="form.scopes"
              :placeholder="form.type === 'oauth2' ? '例如 read:user user:email' : '例如 groups 或 roles，默认已含 openid profile email'"
              class="h-9 text-xs font-mono"
              :disabled="demoMode || loading"
            />
          </div>

          <!-- 全模式通用：建号策略与默认角色 -->
          <div class="p-2.5 px-3.5 rounded-xl border border-border/70 bg-muted/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
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
            <div v-if="form.auto_create" class="flex items-center justify-between sm:justify-start gap-2 pt-1 sm:pt-0 border-t border-border/40 sm:border-0">
              <span class="text-[11px] text-muted-foreground shrink-0">分配角色:</span>
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

          <!-- 底部操作按钮栏（与上方 PasswordSettings / OtpSettings 高度对齐） -->
          <div class="pt-3 border-t border-border/50 flex flex-col sm:flex-row sm:items-center justify-between gap-3 mt-1">
            <span class="text-[11px] text-muted-foreground">
              单点登录配置保存后立即在登录页面生效
            </span>
            <div class="flex items-center gap-2 w-full sm:w-auto">
              <Button
                type="button"
                variant="outline"
                @click="handleTestConnection"
                :disabled="testing || saving || loading || demoMode"
                class="h-8.5 px-3 text-xs font-medium shadow-xs gap-1.5 flex-1 sm:flex-none"
              >
                <Loader2 v-if="testing" class="h-3.5 w-3.5 animate-spin" />
                <Zap v-else class="h-3.5 w-3.5 text-sky-500" />
                <span>{{ testing ? '测试中...' : '测试连接' }}</span>
              </Button>
              <Button
                @click="save"
                :disabled="saving || testing || loading || demoMode"
                class="h-8.5 px-4 text-xs font-medium shadow-sm gap-1.5 flex-1 sm:flex-none"
              >
                <Loader2 v-if="saving" class="h-3.5 w-3.5 animate-spin" />
                <Save v-else class="h-3.5 w-3.5" />
                <span>保存单点登录配置</span>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

