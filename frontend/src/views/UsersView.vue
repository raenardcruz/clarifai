<template>
  <div class="users-page">
    <div class="page-header mb-8 flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold mb-1">User Management</h1>
        <p class="text-sm text-gray-500">Approve registrations and manage roles</p>
      </div>
    </div>

    <div class="users-card card">
      <DataTable :value="users" class="users-table-pv">
        <Column field="email" header="Username" class="font-medium"></Column>
        <Column header="Role">
          <template #body="slotProps">
            <Select 
              v-model="slotProps.data.role" 
              :options="roleOptions" 
              optionLabel="label" 
              optionValue="value" 
              class="w-32 text-sm" 
              :disabled="slotProps.data.email === 'admin'"
            />
          </template>
        </Column>
        <Column header="Status">
          <template #body="slotProps">
            <div class="flex items-center gap-2">
              <span class="status-dot" :class="slotProps.data.is_approved ? 'bg-green-500' : 'bg-yellow-500'"></span>
              {{ slotProps.data.is_approved ? 'Approved' : 'Pending' }}
            </div>
          </template>
        </Column>
        <Column header="Limit (hrs)">
          <template #body="slotProps">
            <InputText 
              v-model.number="slotProps.data.transcription_limit_hours" 
              type="number"
              min="0"
              class="w-20 text-sm"
              :disabled="slotProps.data.email === 'admin'"
            />
          </template>
        </Column>
        <Column header="Actions">
          <template #body="slotProps">
            <div class="flex gap-2">
              <Button 
                v-if="!slotProps.data.is_approved" 
                label="Approve" 
                severity="success" 
                size="small" 
                @click="updateUser(slotProps.data, true)"
              />
              <Button 
                v-else 
                label="Save" 
                severity="secondary" 
                size="small" 
                @click="updateUser(slotProps.data, true)"
                :disabled="slotProps.data.email === 'admin'"
              />
              <Button 
                label="Remove" 
                severity="danger" 
                size="small" 
                @click="deleteUser(slotProps.data.id)"
                :disabled="slotProps.data.email === 'admin'"
              />
            </div>
          </template>
        </Column>
      </DataTable>
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
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'

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
    toast.add({ severity: 'success', summary: 'Success', detail: 'User updated successfully', life: 3000 })
    fetchUsers()
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Update Failed', detail: e.response?.data?.detail || "Failed to update user", life: 5000 })
  }
}

const deleteUser = async (id) => {
  if (confirm("Are you sure you want to remove this user?")) {
    try {
      await axios.delete(`/api/users/${id}`)
      toast.add({ severity: 'success', summary: 'Removed', detail: 'User removed successfully', life: 3000 })
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
  max-width: 1000px;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
}

.users-table th {
  text-align: left;
  padding: 1rem;
  border-bottom: 2px solid var(--neutral-200);
  color: var(--text-muted);
  font-weight: 600;
  font-size: 0.875rem;
}

.users-table td {
  padding: 1rem;
  border-bottom: 1px solid var(--neutral-100);
  vertical-align: middle;
}

.status-dot {
  width: 8px; height: 8px; border-radius: 50%;
}

.bg-green-500 { background-color: #22c55e; }
.bg-yellow-500 { background-color: #eab308; }
.bg-green-100 { background-color: #dcfce7; }
.text-green-700 { color: #15803d; }
.hover\:bg-green-200:hover { background-color: #bbf7d0; }
.bg-red-100 { background-color: #fee2e2; }
.text-red-700 { color: #b91c1c; }
.hover\:bg-red-200:hover { background-color: #fecaca; }
.py-1 { padding-top: 0.25rem; padding-bottom: 0.25rem; }
.px-2 { padding-left: 0.5rem; padding-right: 0.5rem; }
.w-32 { width: 8rem; }
.w-20 { width: 5rem; }
.btn-sm { padding: 0.375rem 0.75rem; font-size: 0.75rem; }
</style>
