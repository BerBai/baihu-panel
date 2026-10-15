import { createApp } from 'vue'
import { install as VueMonacoEditorPlugin } from '@guolao/vue-monaco-editor'
import './assets/index.css'
import 'vue-sonner/style.css'
import App from './App.vue'
import router from './router'
import { isWindowsPlatform } from './windows/platform'

if (typeof document !== 'undefined' && isWindowsPlatform()) {
  document.documentElement.classList.add('is-windows')
}

window.addEventListener('vite:preloadError', () => {
  console.log('Detected vite preload error. Reloading page...')
  window.location.reload()
})

// 过滤 Monaco Editor 剪贴板 workaround 在页面点击时抛出的内部流程取消伪错误
const isCanceledError = (a: any) =>
  a === 'Canceled' ||
  a?.name === 'Canceled' ||
  a?.message === 'Canceled' ||
  (typeof a?.toString === 'function' && a.toString().includes('Canceled'))

const origLog = console.log
console.log = function (...args: any[]) {
  if (args[0] === '%c  ERR' && args.some(isCanceledError)) {
    return
  }
  origLog.apply(console, args)
}

const origError = console.error
console.error = function (...args: any[]) {
  if (args.some(isCanceledError)) {
    return
  }
  origError.apply(console, args)
}

const BASE_URL = (window as any).__BASE_URL__ || ''
const origin = window.location.origin

createApp(App)
  .use(router)
  .use(VueMonacoEditorPlugin, {
    paths: {
      vs: __MONACO_CDN__ || `${origin}${BASE_URL}/assets/vs`
    }
  })
  .mount('#app')
