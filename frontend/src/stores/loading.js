import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useLoadingStore = defineStore('loading', () => {
  const activeTasks = ref(0)
  const isNavigating = ref(false)
  const visible = ref(false)
  const progress = ref(0)

  let trickleTimer = null
  let hideTimer = null
  let navFinishTimer = null

  const isBusy = computed(() => isNavigating.value || activeTasks.value > 0)

  const clearTimers = () => {
    if (trickleTimer) {
      clearInterval(trickleTimer)
      trickleTimer = null
    }
    if (hideTimer) {
      clearTimeout(hideTimer)
      hideTimer = null
    }
  }

  const startBar = () => {
    if (hideTimer) {
      clearTimeout(hideTimer)
      hideTimer = null
    }

    if (!visible.value) {
      visible.value = true
      progress.value = 14
    } else if (progress.value < 25) {
      progress.value = 25
    }

    if (!trickleTimer) {
      trickleTimer = setInterval(() => {
        if (progress.value < 55) {
          progress.value += 8
        } else if (progress.value < 80) {
          progress.value += 3.5
        } else if (progress.value < 91) {
          progress.value += 0.8
        }
      }, 180)
    }
  }

  const completeBarIfIdle = () => {
    if (isBusy.value) return
    if (trickleTimer) {
      clearInterval(trickleTimer)
      trickleTimer = null
    }
    if (!visible.value) return

    progress.value = 100
    hideTimer = setTimeout(() => {
      visible.value = false
      progress.value = 0
      hideTimer = null
    }, 260)
  }

  const startNavigation = () => {
    if (navFinishTimer) {
      clearTimeout(navFinishTimer)
      navFinishTimer = null
    }
    isNavigating.value = true
    startBar()
  }

  const finishNavigation = () => {
    // Brief grace window so synchronous view mounts that immediately call startTask()
    // merge seamlessly into a single smooth progress sweep instead of double-flashing.
    if (navFinishTimer) clearTimeout(navFinishTimer)
    navFinishTimer = setTimeout(() => {
      isNavigating.value = false
      navFinishTimer = null
      completeBarIfIdle()
    }, 70)
  }

  const startTask = () => {
    activeTasks.value += 1
    startBar()
  }

  const finishTask = () => {
    activeTasks.value = Math.max(0, activeTasks.value - 1)
    completeBarIfIdle()
  }

  const reset = () => {
    clearTimers()
    if (navFinishTimer) {
      clearTimeout(navFinishTimer)
      navFinishTimer = null
    }
    activeTasks.value = 0
    isNavigating.value = false
    visible.value = false
    progress.value = 0
  }

  return {
    activeTasks,
    isNavigating,
    isBusy,
    visible,
    progress,
    startNavigation,
    finishNavigation,
    startTask,
    finishTask,
    reset,
  }
})
