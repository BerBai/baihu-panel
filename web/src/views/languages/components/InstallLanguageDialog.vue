<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Search, Loader2, Check, ChevronsUpDown, Globe } from 'lucide-vue-next'
import { api } from '@/api'
import { toast } from 'vue-sonner'
import { getLangIcon } from '@/utils/icons'
import { cn } from '@/lib/utils'

const emit = defineEmits<{
    (e: 'install', plugin: string, version: string, mirrorUrl: string, saveDefault: boolean): void
}>()

const isOpen = ref(false)

const newLangPlugin = ref('')
const newLangVersion = ref('')

const MIRROR_PRESETS = [
    { label: 'gh-proxy.com (推荐)', value: 'https://gh-proxy.com/' },
    { label: 'ghfast.top', value: 'https://ghfast.top/' },
    { label: 'ghproxy.net', value: 'https://ghproxy.net/' },
    { label: 'github.moeyy.xyz', value: 'https://github.moeyy.xyz/' },
    { label: '不使用加速 (官方直连)', value: 'none' },
    { label: '自定义镜像源', value: 'custom' },
]

const selectedMirror = ref('https://gh-proxy.com/')
const customMirror = ref('')
const saveAsDefaultMirror = ref(false)

const availablePlugins = ref<string[]>([])
const loadingPlugins = ref(false)
const pluginSearch = ref('')
const openPluginPopover = ref(false)

const availableVersions = ref<string[]>([])
const loadingVersions = ref(false)
const versionSearch = ref('')
const openVersionPopover = ref(false)

const filteredPlugins = computed(() => {
    if (!pluginSearch.value) return availablePlugins.value
    const s = pluginSearch.value.toLowerCase()
    return availablePlugins.value.filter(p => p.toLowerCase().includes(s))
})

const filteredVersions = computed(() => {
    const list = availableVersions.value
    if (!versionSearch.value) return list
    const s = versionSearch.value.toLowerCase()
    return list.filter(v => v.toLowerCase().includes(s))
})

async function fetchPlugins() {
    if (availablePlugins.value.length > 0) return
    loadingPlugins.value = true
    try {
        availablePlugins.value = await api.mise.plugins()
    } catch (e) {
        console.error('Fetch plugins failed', e)
    } finally {
        loadingPlugins.value = false
    }
}

async function fetchVersions(plugin: string) {
    if (!plugin) return
    loadingVersions.value = true
    availableVersions.value = []
    try {
        availableVersions.value = await api.mise.versions(plugin)
    } catch (e) {
        console.error('Fetch versions failed', e)
    } finally {
        loadingVersions.value = false
        if (!newLangVersion.value && availableVersions.value.length > 0) {
            newLangVersion.value = availableVersions.value[0] || ''
        }
    }
}

async function loadDefaultMirror() {
    try {
        const envs = await api.mise.getEnvs()
        const repl = envs['MISE_URL_REPLACEMENTS']
        if (repl) {
            try {
                const map = JSON.parse(repl)
                const target = (map['https://github.com/'] || '') as string
                if (target.endsWith('https://github.com/')) {
                    const prefix = target.slice(0, -'https://github.com/'.length)
                    const found = MIRROR_PRESETS.find(p => p.value !== 'none' && p.value !== 'custom' && (p.value === prefix || p.value.replace(/\/$/, '') === prefix.replace(/\/$/, '')))
                    if (found) {
                        selectedMirror.value = found.value
                    } else if (prefix) {
                        selectedMirror.value = 'custom'
                        customMirror.value = prefix
                    }
                    return
                }
            } catch {
                // ignore JSON parse error
            }
        }
    } catch (e) {
        console.error('Failed to load global envs for mirror', e)
    }
}

watch(newLangPlugin, (newVal) => {
    if (newVal) {
        fetchVersions(newVal)
    } else {
        availableVersions.value = []
    }
    newLangVersion.value = ''
})

function openDialog() {
    newLangPlugin.value = ''
    newLangVersion.value = ''
    saveAsDefaultMirror.value = false
    isOpen.value = true
    fetchPlugins()
    loadDefaultMirror()
}

function startInstall() {
    if (!newLangPlugin.value.trim()) {
        toast.error('请输入或选择语言名称')
        return
    }
    if (!newLangVersion.value.trim()) {
        toast.error('请选择版本')
        return
    }

    let finalMirror = ''
    if (selectedMirror.value === 'none') {
        finalMirror = ''
    } else if (selectedMirror.value === 'custom') {
        let custom = customMirror.value.trim()
        if (!custom) {
            toast.error('请输入自定义加速源地址')
            return
        }
        if (!custom.endsWith('/')) {
            custom += '/'
        }
        finalMirror = custom
    } else {
        finalMirror = selectedMirror.value
    }

    const version = newLangVersion.value.trim()
    isOpen.value = false
    emit('install', newLangPlugin.value.trim(), version, finalMirror, saveAsDefaultMirror.value)
}

defineExpose({ openDialog })
</script>

<template>
    <Dialog v-model:open="isOpen">
        <DialogContent class="sm:max-w-[420px]">
            <DialogHeader>
                <DialogTitle>管理语言运行时</DialogTitle>
                <DialogDescription>配置并安装新的编程语言环境</DialogDescription>
            </DialogHeader>
            <div class="grid gap-5 py-4">
                <div class="grid gap-2">
                    <Label>语言名称 (Mise Plugin)</Label>
                    <Popover v-model:open="openPluginPopover">
                        <PopoverTrigger asChild>
                            <Button variant="outline" role="combobox" :aria-expanded="openPluginPopover"
                                class="justify-between w-full font-normal">
                                {{ newLangPlugin || "选择或输入语言..." }}
                                <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent class="p-0 w-[var(--reka-popover-trigger-width)]" align="start">
                            <div class="p-2 border-b">
                                <div class="relative">
                                    <Search
                                        class="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                                    <Input v-model="pluginSearch" placeholder="搜索插件..." class="h-8 pl-8 text-xs"
                                        @keydown.enter="() => { if (pluginSearch) { newLangPlugin = pluginSearch; openPluginPopover = false } }" />
                                </div>
                            </div>
                            <ScrollArea class="h-64">
                                <div class="p-1">
                                    <div v-if="loadingPlugins" class="flex items-center justify-center py-6">
                                        <Loader2 class="h-4 w-4 animate-spin text-muted-foreground" />
                                    </div>
                                    <div v-else-if="filteredPlugins.length === 0"
                                        class="py-6 text-center text-xs text-muted-foreground">
                                        未找到匹配插件
                                    </div>
                                    <template v-else>
                                        <button v-for="p in filteredPlugins" :key="p"
                                            @click="() => { newLangPlugin = p; openPluginPopover = false }"
                                            class="w-full flex items-center px-2 py-1.5 text-sm rounded-sm hover:bg-muted text-left transition-colors group">
                                            <div
                                                class="mr-2 h-4 w-4 shrink-0 flex items-center justify-center relative">
                                                <div v-if="getLangIcon(p)"
                                                    class="w-full h-full rounded-sm bg-white/80 overflow-hidden p-0.5">
                                                    <img :src="getLangIcon(p)"
                                                        class="w-full h-full object-contain" />
                                                </div>
                                                <div v-else
                                                    class="w-full h-full flex items-center justify-center bg-primary/10 rounded-sm text-[8px] font-bold uppercase">
                                                    {{ p.substring(0, 2) }}
                                                </div>
                                                <Check v-if="newLangPlugin === p"
                                                    class="absolute -right-2 -top-1 h-3 w-3 text-primary bg-background rounded-full border shadow-sm" />
                                            </div>
                                            <span :class="{ 'font-bold text-primary': newLangPlugin === p }">{{ p
                                            }}</span>
                                        </button>
                                    </template>
                                </div>
                            </ScrollArea>
                        </PopoverContent>
                    </Popover>
                </div>

                <div class="grid gap-2">
                    <Label>版本</Label>
                    <Popover v-model:open="openVersionPopover">
                        <PopoverTrigger asChild :disabled="!newLangPlugin">
                            <Button variant="outline" role="combobox" :aria-expanded="openVersionPopover"
                                class="justify-between w-full font-normal" :disabled="!newLangPlugin">
                                {{ newLangVersion || "选择或输入版本..." }}
                                <div class="flex items-center">
                                    <Loader2 v-if="loadingVersions" class="mr-2 h-3 w-3 animate-spin opacity-50" />
                                    <ChevronsUpDown class="h-4 w-4 shrink-0 opacity-50" />
                                </div>
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent class="p-0 w-[var(--reka-popover-trigger-width)]" align="start">
                            <div class="p-2 border-b">
                                <div class="relative">
                                    <Search
                                        class="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                                    <Input v-model="versionSearch" placeholder="搜索版本..." class="h-8 pl-8 text-xs"
                                        @keydown.enter="() => { if (versionSearch) { newLangVersion = versionSearch; openVersionPopover = false } }" />
                                </div>
                            </div>
                            <ScrollArea class="h-64">
                                <div class="p-1">
                                    <div class="p-1">
                                        <div v-if="loadingVersions" class="flex items-center justify-center py-6">
                                            <Loader2 class="h-4 w-4 animate-spin text-muted-foreground" />
                                        </div>
                                        <div v-else-if="filteredVersions.length === 0"
                                            class="py-6 text-center text-xs text-muted-foreground">
                                            未找到匹配版本
                                        </div>
                                        <template v-else>
                                            <button v-for="v in filteredVersions" :key="v"
                                                @click="() => { newLangVersion = v; openVersionPopover = false }"
                                                class="w-full flex items-center px-2 py-1.5 text-sm rounded-sm hover:bg-muted text-left transition-colors">
                                                <Check
                                                    :class="cn('mr-2 h-3.5 w-3.5', newLangVersion === v ? 'opacity-100' : 'opacity-0')" />
                                                {{ v }}
                                            </button>
                                        </template>
                                    </div>
                                </div>
                            </ScrollArea>
                        </PopoverContent>
                    </Popover>
                </div>

                <div class="grid gap-2">
                    <Label class="flex items-center gap-1.5">
                        <Globe class="h-3.5 w-3.5 text-primary" />
                        <span>下载加速 (GitHub 镜像)</span>
                    </Label>
                    <Select v-model="selectedMirror">
                        <SelectTrigger class="w-full">
                            <SelectValue placeholder="选择加速源..." />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem v-for="item in MIRROR_PRESETS" :key="item.value" :value="item.value">
                                {{ item.label }}
                            </SelectItem>
                        </SelectContent>
                    </Select>
                    <div v-if="selectedMirror === 'custom'" class="pt-1">
                        <Input v-model="customMirror" placeholder="输入代理前缀，如 https://ghfast.top/" class="h-8 text-xs" />
                    </div>
                    <div class="flex items-center space-x-2 pt-1">
                        <Checkbox id="save-default-mirror" :checked="saveAsDefaultMirror"
                            @update:checked="(val: boolean | 'indeterminate') => saveAsDefaultMirror = !!val" />
                        <label for="save-default-mirror" class="text-xs text-muted-foreground cursor-pointer select-none">
                            设为全局默认加速源 (保存到 Mise 配置)
                        </label>
                    </div>
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="isOpen = false">取消</Button>
                <Button @click="startInstall">开始安装</Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
