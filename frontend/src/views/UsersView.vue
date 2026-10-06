<template>
  <div class="users-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">User Management</h1>
        <p class="page-subtitle">Review registrations, manage authorization roles, and set transcription allowances</p>
      </div>
      <div class="user-count-chip">
        <Users size="15" />
        <span>{{ users.length }} {{ users.length === 1 ? 'Account' : 'Accounts' }}</span>
      </div>
    </div>

    <!-- Apple Table Card -->
    <div class="users-card">
      <div class="apple-table-wrapper">
        <DataTable :value="users" class="apple-datatable" responsiveLayout="scroll">
          <!-- User / Email Column -->
          <Column header="USER">
            <template #body="slotProps">
              <div class="user-row-cell">
                <div class="user-avatar-pill">
                  {{ slotProps.data.email?.[0].toUpperCase() || 'U' }}
                </div>
                <div class="user-name-group">
                  <span class="user-email-text">{{ slotProps.data.email }}</span>
                  <span class="admin-badge" v-if="slotProps.data.email === 'admin'">System Master</span>
                </div>
              </div>
            </template>
          </Column>

          <!-- Role Column -->
          <Column header="ROLE">
            <template #body="slotProps">
              <Select 
                v-model="slotProps.data.role" 
                :options="roleOptions" 
                optionLabel="label" 
                optionValue="value" 
                class="apple-role-select" 
                :disabled="slotProps.data.email === 'admin'"
              />
            </template>
          </Column>

          <!-- Status Column -->
          <Column header="STATUS">
            <template #body="slotProps">
              <span class="apple-status-pill" :class="slotProps.data.is_approved ? 'status-approved' : 'status-pending'">
                <span class="dot-indicator"></span>
                <span>{{ slotProps.data.is_approved ? 'Approved' : 'Pending' }}</span>
              </span>
            </template>
          </Column>

          <!-- Quota Column -->
          <Column header="QUOTA (HRS)">
            <template #body="slotProps">
              <div class="limit-input-wrapper">
                <input 
                  type="number"
                  v-model.number="slotProps.data.transcription_limit_hours" 
                  min="0"
                  class="apple-num-input font-tabular"
                  :disabled="slotProps.data.email === 'admin'"
                />
                <span class="hrs-suffix">hrs</span>
              </div>
            </template>
          </Column>

          <!-- Actions Column -->
          <Column header="ACTIONS">
            <template #body="slotProps">
              <div class="actions-cell">
                <button 
                  v-if="!slotProps.data.is_approved" 
                  class="btn-action btn-approve"
                  @click="updateUser(slotProps.data, true)"
                >
                  <Check size="13" />
                  <span>Approve</span>
                </button>

                <button 
                  v-else 
                  class="btn-action btn-save"
                  @click="updateUser(slotProps.data, true)"
                  :disabled="slotProps.data.email === 'admin'"
                >
                  <span>Save</span>
                </button>

                <button 
                  class="btn-action btn-remove"
                  @click="deleteUser(slotProps.data.id)"
                  :disabled="slotProps.data.email === 'admin'"
                  title="Remove user"
                >
                  <Trash2 size="13" />
                </button>
              </div>
            </template>
          </Column>
        </DataTable>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useToast } from 'primevue/usetoast'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Select from 'primevue/select'
import { Users, Check, Trash2 } from '@lucide/vue'

const users = ref([])
const toast = useToast()

const roleOptions = ref([
  { label: 'User', value: 'user' },
  { label: 'Admin', value: 'admin' }
])

const fetchUsers = async () => {
  try {
    const res = await axios.get('/api/users')
    users.value = res.data
  } catch (e) {
    console.error("Failed to load users", e)
  }
}

const updateUser = async (user, isApproved) => {
  try {
    await axios.put(`/api/users/${user.id}`, {
      role: user.role,
      is_approved: isApproved,
      transcription_limit_hours: user.transcription_limit_hours
    })
    toast.add({ severity: 'success', summary: 'Success', detail: 'User updated successfully.', life: 3000 })
    fetchUsers()
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Update Failed', detail: e.response?.data?.detail || "Failed to update user", life: 5000 })
  }
}

const deleteUser = async (id) => {
  if (confirm("Are you sure you want to remove this user?")) {
    try {
      await axios.delete(`/api/users/${id}`)
      toast.add({ severity: 'success', summary: 'Removed', detail: 'User removed successfully.', life: 3000 })
      fetchUsers()
    } catch (e) {
      toast.add({ severity: 'error', summary: 'Error', detail: e.response?.data?.detail || "Failed to delete user", life: 5000 })
    }
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.users-page {
  max-width: 1080px;
  margin: 0 auto;
  padding: 2.5rem 2rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
  margin-bottom: 0.35rem;
}

.page-subtitle {
  font-size: 0.95rem;
  color: var(--text-secondary);
}

.user-count-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.4rem 0.85rem;
  border-radius: var(--radius-full);
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.08);
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.users-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 4px 20px -2px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
  overflow: hidden;
}

.apple-table-wrapper {
  overflow-x: auto;
}

.user-row-cell {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.25rem 0;
}

.user-avatar-pill {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, #0071E3, #AF52DE);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.825rem;
  flex-shrink: 0;
}

.user-name-group {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.user-email-text {
  font-weight: 600;
  font-size: 0.88rem;
  color: var(--text-primary);
}

.admin-badge {
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--apple-blue);
  background: var(--apple-blue-light);
  padding: 0.05rem 0.4rem;
  border-radius: var(--radius-full);
  display: inline-block;
  width: fit-content;
}

.apple-role-select {
  width: 110px;
  font-size: 0.825rem;
}

.apple-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.6rem;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
}

.dot-indicator {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
}

.status-approved {
  background: var(--apple-green-light);
  color: #1E8E3E;
}

.status-approved .dot-indicator {
  background: #34C759;
}

.status-pending {
  background: var(--apple-orange-light);
  color: #D97706;
}

.status-pending .dot-indicator {
  background: #FF9500;
}

.limit-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.apple-num-input {
  width: 64px;
  padding: 0.3rem 0.5rem;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: var(--radius-sm);
  font-size: 0.825rem;
  font-family: inherit;
  outline: none;
  background: var(--bg-primary);
}

.apple-num-input:focus {
  border-color: var(--apple-blue);
  background: var(--white);
}

.hrs-suffix {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.font-tabular {
  font-variant-numeric: tabular-nums;
}

.actions-cell {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.65rem;
  border-radius: var(--radius-sm);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: var(--transition-fast);
}

.btn-approve {
  background: var(--apple-green-light);
  color: #1E8E3E;
}

.btn-approve:hover {
  background: #D1F2D9;
}

.btn-save {
  background: rgba(0, 0, 0, 0.05);
  color: var(--text-secondary);
}

.btn-save:hover {
  background: rgba(0, 0, 0, 0.08);
  color: var(--text-primary);
}

.btn-remove {
  background: transparent;
  color: var(--apple-red);
  padding: 0.35rem;
}

.btn-remove:hover {
  background: var(--apple-red-light);
}

.btn-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* PrimeVue DataTable custom styling */
:deep(.p-datatable-header) {
  background: transparent;
  border: none;
}

:deep(.p-datatable-thead > tr > th) {
  background: rgba(0, 0, 0, 0.02) !important;
  color: var(--text-muted) !important;
  font-size: 0.7rem !important;
  font-weight: 700 !important;
  letter-spacing: 0.05em !important;
  padding: 0.85rem 1.25rem !important;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06) !important;
}

:deep(.p-datatable-tbody > tr > td) {
  padding: 0.85rem 1.25rem !important;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04) !important;
  color: var(--text-primary) !important;
}

:deep(.p-datatable-tbody > tr:hover) {
  background: rgba(0, 113, 227, 0.02) !important;
}
</style>
