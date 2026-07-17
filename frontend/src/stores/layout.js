import { ref } from 'vue'

export const isSidebarOpen = ref(false)

export function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

export function closeSidebar() {
  isSidebarOpen.value = false
}
