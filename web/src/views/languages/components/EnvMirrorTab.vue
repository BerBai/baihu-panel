<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Check, ArrowRight, Trash2, Cpu, Package } from 'lucide-vue-next'
import { api } from '@/api'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'

const globalEnvs = ref<Record<string, string>>({})
const loadingEnvs = ref(false)
const showAddEnvDialog = ref(false)
const newEnvKey = ref('')
const newEnvValue = ref('')

const RUNTIME_PRESETS = [
    {
        label: 'Mise GitHub 运行时加速',
        key: 'MISE_URL_REPLACEMENTS',
        value: '{"https://github.com/":"https://gh-proxy.com/https://github.com/"}',
        desc: '加速 Python、Node 等预编译包及 GitHub 插件下载'
    },
    {
        label: 'Node.js 运行时镜像',
        key: 'MISE_NODE_MIRROR_URL',
        value: 'https://npmmirror.com/mirrors/node/',
        desc: 'Mise 安装 Node.js 时走 npmmirror 国内镜像'
    },
    {
        label: 'Go 运行时官方镜像',
        key: 'MISE_GO_DOWNLOAD_MIRROR',
        value: 'https://golang.google.cn/dl',
        desc: 'Mise 安装 Go 时走国内官方镜像'
    },
    {
        label: 'Python 编译源码镜像',
        key: 'PYTHON_BUILD_MIRROR_URL',
        value: 'https://registry.npmmirror.com/-/binary/python/',
        desc: 'python-build 编译模式下载 Python 源码包镜像'
    },
]

const PACKAGE_PRESETS = [
    { label: 'npm (淘宝镜像)', key: 'NPM_CONFIG_REGISTRY', value: 'https://registry.npmmirror.com', desc: 'npm 包下载镜像' },
    { label: 'Python (清华镜像)', key: 'PIP_INDEX_URL', value: 'https://pypi.tuna.tsinghua.edu.cn/simple', desc: 'pip 依赖包镜像' },
    { label: 'Go (goproxy.cn)', key: 'GOPROXY', value: 'https://goproxy.cn,direct', desc: 'Go 模块依赖代理' },
    { label: 'Rust (Sparse protocol)', key: 'CARGO_REGISTRY_CRATES_IO_PROTOCOL', value: 'sparse', desc: 'Cargo 稀疏索引协议' },
    { label: 'Rust (中科大镜像)', key: 'CARGO_REGISTRIES_CRATES_IO_INDEX', value: 'https://mirrors.ustc.edu.cn/crates.io-index', desc: 'crates.io 国内镜像' },
]

async function loadGlobalEnvs() {
    loadingEnvs.value = true
    try {
        globalEnvs.value = await api.mise.getEnvs()
    } catch (e) {
        toast.error('获取全局环境变量失败: ' + e)
    } finally {
        loadingEnvs.value = false
    }
}

async function handleSetEnv(key: string, value: string) {
    try {
        await api.mise.setEnv(key, value)
        toast.success(`环境变量 ${key} 设置成功`)
        loadGlobalEnvs()
    } catch (e) {
        toast.error('设置失败: ' + e)
    }
}

async function handleUnsetEnv(key: string) {
    try {
        await api.mise.unsetEnv(key)
        toast.success(`环境变量 ${key} 已移除`)
        loadGlobalEnvs()
    } catch (e) {
        toast.error('移除失败: ' + e)
    }
}

function applyPreset(preset: { key: string, value: string }) {
    newEnvKey.value = preset.key
    newEnvValue.value = globalEnvs.value[preset.key] || preset.value
    showAddEnvDialog.value = true
}

function fillReplacement(mirrorPrefix: string) {
    newEnvValue.value = JSON.stringify({ "https://github.com/": `${mirrorPrefix}https://github.com/` })
}

function startAddEnv() {
    if (!newEnvKey.value.trim()) return
    handleSetEnv(newEnvKey.value.trim(), newEnvValue.value.trim())
    showAddEnvDialog.value = false
    newEnvKey.value = ''
    newEnvValue.value = ''
}

onMounted(() => {
    loadGlobalEnvs()
})
</script>

<template>
    <div class="space-y-6">
        <!-- 运行时下载加速区域 -->
        <div class="rounded-lg border bg-card p-6">
            <div class="flex items-start justify-between mb-6">
                <div>
                    <h3 class="text-base font-bold flex items-center gap-2">
                        <Cpu class="h-4 w-4 text-primary" />
                        运行时下载加速 (Mise Runtime Mirror)
                    </h3>
                    <p class="text-xs text-muted-foreground mt-1 max-w-2xl">
                        加速安装编程语言本身（Python、Node、Go 等）及下载 GitHub Releases 预编译包。解决国内服务器连接 GitHub 超时慢的问题。
                    </p>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-4">
                <div v-for="preset in RUNTIME_PRESETS" :key="preset.key" :class="cn(
                    'relative p-4 rounded-xl border transition-all group overflow-hidden',
                    globalEnvs[preset.key] ? 'bg-primary/5 border-primary shadow-sm' : 'bg-muted/30 hover:border-primary/50 cursor-pointer'
                )"
                    @click="applyPreset(preset)">

                    <div class="flex items-start justify-between mb-2">
                        <div>
                            <div class="font-semibold text-sm">{{ preset.label }}</div>
                            <div class="text-[11px] text-muted-foreground mt-0.5">{{ preset.desc }}</div>
                        </div>
                        <div v-if="globalEnvs[preset.key]"
                            class="h-5 w-5 shrink-0 rounded-full bg-primary text-primary-foreground flex items-center justify-center">
                            <Check class="h-3 w-3" />
                        </div>
                    </div>

                    <div class="text-[10px] font-mono mb-3">
                        <div class="text-muted-foreground truncate">{{ preset.key }}</div>
                        <div v-if="globalEnvs[preset.key]"
                            class="mt-1 text-primary/80 font-medium truncate bg-primary/5 px-1.5 py-0.5 rounded-sm border border-primary/10"
                            :title="globalEnvs[preset.key]">
                            {{ globalEnvs[preset.key] }}
                        </div>
                    </div>

                    <div class="flex items-center justify-between mt-auto">
                        <span v-if="globalEnvs[preset.key]" class="text-[10px] font-medium text-primary">
                            已启用 (点击修改)
                        </span>
                        <span v-else
                            class="text-[10px] font-medium text-muted-foreground group-hover:text-primary transition-colors flex items-center">
                            立即启用
                            <ArrowRight class="h-3 w-3 ml-1 group-hover:translate-x-1 transition-transform" />
                        </span>

                        <Button v-if="globalEnvs[preset.key]" variant="ghost" size="icon"
                            class="h-6 w-6 text-destructive hover:bg-destructive/10"
                            @click.stop="handleUnsetEnv(preset.key)">
                            <Trash2 class="h-3.5 w-3.5" />
                        </Button>
                    </div>
                </div>
            </div>
        </div>

        <!-- 包依赖镜像加速区域 -->
        <div class="rounded-lg border bg-card p-6">
            <div class="flex items-start justify-between mb-6">
                <div>
                    <h3 class="text-base font-bold flex items-center gap-2">
                        <Package class="h-4 w-4 text-primary" />
                        包管理器依赖源 (Package Registry)
                    </h3>
                    <p class="text-xs text-muted-foreground mt-1 max-w-2xl">
                        配置 pip、npm、go get、cargo 等包管理器在安装第三方依赖库时使用的国内镜像源。
                    </p>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <div v-for="preset in PACKAGE_PRESETS" :key="preset.key" :class="cn(
                    'relative p-4 rounded-xl border transition-all group overflow-hidden',
                    globalEnvs[preset.key] ? 'bg-primary/5 border-primary shadow-sm' : 'bg-muted/30 hover:border-primary/50 cursor-pointer'
                )"
                    @click="applyPreset(preset)">

                    <div class="flex items-start justify-between mb-2">
                        <div>
                            <div class="font-semibold text-sm">{{ preset.label }}</div>
                            <div class="text-[11px] text-muted-foreground mt-0.5">{{ preset.desc }}</div>
                        </div>
                        <div v-if="globalEnvs[preset.key]"
                            class="h-5 w-5 shrink-0 rounded-full bg-primary text-primary-foreground flex items-center justify-center">
                            <Check class="h-3 w-3" />
                        </div>
                    </div>

                    <div class="text-[10px] font-mono mb-3">
                        <div class="text-muted-foreground truncate">{{ preset.key }}</div>
                        <div v-if="globalEnvs[preset.key]"
                            class="mt-1 text-primary/80 font-medium truncate bg-primary/5 px-1.5 py-0.5 rounded-sm border border-primary/10"
                            :title="globalEnvs[preset.key]">
                            {{ globalEnvs[preset.key] }}
                        </div>
                    </div>

                    <div class="flex items-center justify-between mt-auto">
                        <span v-if="globalEnvs[preset.key]" class="text-[10px] font-medium text-primary">
                            已启用 (点击修改)
                        </span>
                        <span v-else
                            class="text-[10px] font-medium text-muted-foreground group-hover:text-primary transition-colors flex items-center">
                            立即启用
                            <ArrowRight class="h-3 w-3 ml-1 group-hover:translate-x-1 transition-transform" />
                        </span>

                        <Button v-if="globalEnvs[preset.key]" variant="ghost" size="icon"
                            class="h-6 w-6 text-destructive hover:bg-destructive/10"
                            @click.stop="handleUnsetEnv(preset.key)">
                            <Trash2 class="h-3.5 w-3.5" />
                        </Button>
                    </div>
                </div>
            </div>
        </div>

        <!-- 启用/编辑镜像加速对话框 -->
        <Dialog v-model:open="showAddEnvDialog">
            <DialogContent class="sm:max-w-[440px]">
                <DialogHeader>
                    <DialogTitle>设置镜像加速</DialogTitle>
                    <DialogDescription>
                        正在配置 <span class="font-bold text-primary">{{ newEnvKey }}</span>
                    </DialogDescription>
                </DialogHeader>
                <div class="grid gap-4 py-4">
                    <div class="grid gap-2">
                        <Label>环境变量名</Label>
                        <Input v-model="newEnvKey" readonly class="bg-muted cursor-default" />
                    </div>
                    <div class="grid gap-2">
                        <div class="flex items-center justify-between">
                            <Label>镜像源配置 (Value)</Label>
                            <div v-if="newEnvKey === 'MISE_URL_REPLACEMENTS'" class="flex gap-1.5 text-[11px]">
                                <button type="button" @click="fillReplacement('https://gh-proxy.com/')"
                                    class="text-primary hover:underline">gh-proxy</button>
                                <span class="text-muted-foreground">|</span>
                                <button type="button" @click="fillReplacement('https://ghfast.top/')"
                                    class="text-primary hover:underline">ghfast</button>
                                <span class="text-muted-foreground">|</span>
                                <button type="button" @click="fillReplacement('https://ghproxy.net/')"
                                    class="text-primary hover:underline">ghproxy</button>
                                <span class="text-muted-foreground">|</span>
                                <button type="button" @click="fillReplacement('https://github.moeyy.xyz/')"
                                    class="text-primary hover:underline">moeyy</button>
                            </div>
                        </div>
                        <Input v-model="newEnvValue" placeholder="请输入镜像地址或配置" @keydown.enter="startAddEnv" />
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" @click="showAddEnvDialog = false">取消</Button>
                    <Button @click="startAddEnv">确认启用</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    </div>
</template>
