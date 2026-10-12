<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Plus, X, SlidersHorizontal } from 'lucide-vue-next'

interface Env {
  plugin: string
  version: string
}

const props = defineProps<{
  open: boolean
  selectedEnvs: Env[]
  langGroups: Record<string, string[]>
  getLangIcon: (plugin: string) => string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'update:selectedEnvs': [value: Env[]]
  'confirm': []
}>()

function addEnv() {
  const newEnvs = [...props.selectedEnvs, { plugin: '', version: '' }]
  emit('update:selectedEnvs', newEnvs)
}

function removeEnv(index: number) {
  const newEnvs = [...props.selectedEnvs]
  newEnvs.splice(index, 1)
  emit('update:selectedEnvs', newEnvs)
}

function updateEnvPlugin(index: number, plugin: string) {
  const newEnvs = [...props.selectedEnvs]
  const group = props.langGroups ? props.langGroups[plugin] : null
  newEnvs[index] = {
    plugin,
    version: (group && group.length > 0) ? (group[0] as string) : ''
  }
  emit('update:selectedEnvs', newEnvs)
}

function updateEnvVersion(index: number, version: string) {
  const newEnvs = [...props.selectedEnvs]
  if (newEnvs[index]) {
    newEnvs[index].version = version
    emit('update:selectedEnvs', newEnvs)
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="val => emit('update:open', val)">
    <DialogContent class="sm:max-w-[420px] font-inter" @pointerDownOutside.prevent>
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <SlidersHorizontal class="h-4 w-4 text-primary" />
          <span>运行环境配置</span>
        </DialogTitle>
        <DialogDescription class="text-xs">
          选择脚本执行所依赖的 Mise 语言环境与对应版本。
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4 py-2">
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <Label class="text-xs font-medium text-foreground/80">语言与版本 (Mise)</Label>
            <span v-if="selectedEnvs.length === 0" class="text-[11px] text-muted-foreground">默认系统环境</span>
          </div>

          <div class="space-y-2">
            <div
              v-for="(env, index) in selectedEnvs"
              :key="index"
              class="flex items-center gap-2"
            >
              <!-- 语言下拉 -->
              <div class="flex-1 min-w-0">
                <Select :modelValue="env.plugin" @update:model-value="newVal => updateEnvPlugin(index, newVal as string)">
                  <SelectTrigger class="w-full h-8.5 text-xs font-inter">
                    <div class="flex items-center gap-2 overflow-hidden">
                      <div class="h-4 w-4 rounded overflow-hidden shrink-0 bg-white border border-black/5 flex items-center justify-center p-0.5" v-if="getLangIcon(env.plugin)">
                        <img :src="getLangIcon(env.plugin)" :alt="env.plugin" class="w-full h-full object-contain" />
                      </div>
                      <SelectValue placeholder="选择语言..." class="truncate" />
                    </div>
                  </SelectTrigger>
                  <SelectContent class="max-h-56 font-inter">
                    <template v-for="(_, lang) in langGroups" :key="lang">
                      <SelectItem v-if="lang !== 'none'" :value="lang as string" class="text-xs font-inter">
                        <div class="flex items-center gap-2">
                          <img v-if="getLangIcon(lang as string)" :src="getLangIcon(lang as string)" class="h-3.5 w-3.5" />
                          <span>{{ lang }}</span>
                        </div>
                      </SelectItem>
                    </template>
                  </SelectContent>
                </Select>
              </div>

              <!-- 版本下拉 -->
              <div class="w-28 shrink-0">
                <Select :modelValue="env.version" @update:model-value="newVal => updateEnvVersion(index, newVal as string)" :disabled="!env.plugin">
                  <SelectTrigger class="w-full h-8.5 text-xs font-inter font-medium tracking-tight">
                    <SelectValue placeholder="版本" />
                  </SelectTrigger>
                  <SelectContent v-if="env.plugin && langGroups && langGroups[env.plugin]" class="max-h-56 font-inter">
                    <SelectItem v-for="v in langGroups[env.plugin]" :key="v" :value="v" class="text-xs font-inter font-medium tracking-tight">
                      {{ v }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <!-- 删除按钮 -->
              <Button
                variant="ghost"
                size="icon"
                class="h-8.5 w-8.5 text-muted-foreground hover:text-destructive shrink-0"
                title="删除"
                @click="removeEnv(index)"
              >
                <X class="h-4 w-4" />
              </Button>
            </div>

            <!-- 添加语言按钮 -->
            <Button
              variant="outline"
              size="sm"
              class="w-full border-dashed h-8 text-xs font-normal text-muted-foreground hover:text-foreground font-inter"
              @click="addEnv"
            >
              <Plus class="h-3.5 w-3.5 mr-1.5" /> 添加语言环境
            </Button>
          </div>
        </div>

        <p class="text-[11px] text-muted-foreground leading-relaxed">
          未配置时将使用系统默认环境运行；如需其他语言运行时，请先在「环境依赖 - 语言依赖」中完成安装。
        </p>
      </div>

      <DialogFooter class="gap-2 sm:gap-0 font-inter">
        <Button variant="outline" size="sm" class="h-8 text-xs" @click="emit('update:open', false)">
          取消
        </Button>
        <Button size="sm" class="h-8 text-xs" @click="emit('confirm')">
          确定
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.font-inter {
  font-family: 'Inter Variable', 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif !important;
  font-feature-settings: "cv05", "cv08", "cv11", "ss01", "ss03", "tnum" !important;
  font-optical-sizing: auto;
}
</style>
