<template>
  <img
    v-if="shouldRender && meta.logoUrl"
    class="source-logo-img"
    :class="[`source-${meta.key}`, `size-${size}`]"
    :src="meta.logoUrl"
    :alt="meta.label"
    :title="`Imported from ${meta.label}`"
  />
</template>

<script setup>
import { computed } from 'vue'
import { getSourceMeta } from '../utils/sources'

const props = defineProps({
  source: {
    type: String,
    default: 'local'
  },
  size: {
    type: String,
    default: 'sm' // 'sm' | 'md'
  },
  showLocal: {
    type: Boolean,
    default: false
  }
})

const meta = computed(() => getSourceMeta(props.source))
const shouldRender = computed(() => meta.value.isImported || props.showLocal)
</script>

<style scoped>
.source-logo-img {
  display: inline-block;
  width: auto;
  object-fit: contain;
  vertical-align: middle;
  flex-shrink: 0;
}

.size-sm {
  height: 15px;
  max-width: 108px;
}

.size-sm.source-fireflies {
  height: 18px;
}

.size-md {
  height: 19px;
  max-width: 136px;
}

.size-md.source-fireflies {
  height: 23px;
}
</style>
