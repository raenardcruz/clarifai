<template>
  <iframe
    v-if="isHtml"
    ref="iframeRef"
    class="summary-html-frame"
    :srcdoc="renderedHtml"
    sandbox="allow-same-origin allow-modals"
    title="Session Summary Document"
    @load="onIframeLoad"
  ></iframe>
  <div v-else class="markdown-content" v-html="parsedMarkdown"></div>
</template>

<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import { isHtmlSummary, renderMarkdownSummary } from '../utils/summary'

const props = defineProps({
  content: {
    type: String,
    default: ''
  }
})

const iframeRef = ref(null)
let resizeObserver = null

const isHtml = computed(() => isHtmlSummary(props.content))
const parsedMarkdown = computed(() => (isHtml.value ? '' : renderMarkdownSummary(props.content)))

// Strip outer card/canvas chrome so the template blends seamlessly inside the host card
const embedOverrideStyle = `<style>
  body {
    background: transparent !important;
    padding: 0 !important;
    margin: 0 !important;
  }
  .document-sheet {
    max-width: 100% !important;
    width: 100% !important;
    margin: 0 !important;
    padding: 0.25rem 0 !important;
    background: transparent !important;
    border: none !important;
    border-radius: 0 !important;
    box-shadow: none !important;
  }
</style>`

const renderedHtml = computed(() => {
  if (!isHtml.value || !props.content) return ''
  if (/<\/head>/i.test(props.content)) {
    return props.content.replace(/<\/head>/i, `${embedOverrideStyle}</head>`)
  }
  return `${embedOverrideStyle}${props.content}`
})

const cleanupObserver = () => {
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }
}

const updateHeight = () => {
  const iframe = iframeRef.value
  if (!iframe) return
  try {
    const doc = iframe.contentDocument || iframe.contentWindow?.document
    if (!doc || !doc.body) return
    const height = Math.max(doc.documentElement.scrollHeight, doc.body.scrollHeight)
    if (height > 0) {
      iframe.style.height = `${height + 16}px`
    }
  } catch {
    // Ignore cross-origin access errors if any
  }
}

const onIframeLoad = () => {
  cleanupObserver()
  updateHeight()

  const iframe = iframeRef.value
  try {
    const doc = iframe?.contentDocument || iframe?.contentWindow?.document
    if (doc?.body && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => updateHeight())
      resizeObserver.observe(doc.body)
    }
    if (doc?.fonts?.ready) {
      doc.fonts.ready.then(() => updateHeight()).catch(() => {})
    }
  } catch {
    // Ignore
  }
}

watch(
  () => props.content,
  () => {
    cleanupObserver()
  }
)

onUnmounted(() => {
  cleanupObserver()
})

const print = () => {
  if (isHtml.value && iframeRef.value?.contentWindow) {
    iframeRef.value.contentWindow.print()
  } else {
    window.print()
  }
}

defineExpose({ print })
</script>

<style scoped>
.summary-html-frame {
  width: 100%;
  min-height: 420px;
  border: none;
  display: block;
  overflow: hidden;
  border-radius: 12px;
  background: transparent;
}
</style>
