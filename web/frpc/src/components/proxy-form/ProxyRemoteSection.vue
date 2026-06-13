<template>
  <template v-if="['tcp', 'udp'].includes(form.type)">
    <div class="field-row two-col">
      <ConfigField label="Remote Port" type="number" v-model="form.remotePort"
        :min="0" :max="65535" prop="remotePort" tip="Use 0 for random port assignment" :readonly="readonly" />
      <div></div>
    </div>
    <div v-if="!isEdit && usedPortsError" class="used-ports-hint">
      <el-alert title="Unable to load used ports from frps dashboard" type="warning" :closable="false" show-icon>
        <template #default>
          <p>Make sure the frps dashboard address, username, and password are correctly configured in the profile settings.</p>
        </template>
      </el-alert>
    </div>
    <div v-if="usedPorts && usedPorts.length > 0" class="used-ports-section">
      <div class="used-ports-label">Used Ports on Server:</div>
      <div class="used-ports-list">
        <el-tag v-for="port in usedPorts" :key="port" size="small" type="info" class="port-tag">
          {{ port }}
        </el-tag>
      </div>
    </div>
    <div v-if="portConflict" class="port-conflict-warning">
      <el-alert title="Port already in use" type="error" :closable="false" show-icon />
    </div>
  </template>
  <template v-if="['http', 'https', 'tcpmux'].includes(form.type)">
    <div class="field-row two-col">
      <ConfigField label="Custom Domains" type="tags" v-model="form.customDomains"
        prop="customDomains" placeholder="example.com" :readonly="readonly" />
      <ConfigField v-if="form.type !== 'tcpmux'" label="Subdomain" type="text"
        v-model="form.subdomain" placeholder="test" :readonly="readonly" />
      <ConfigField v-if="form.type === 'tcpmux'" label="Multiplexer" type="select"
        v-model="form.multiplexer" :options="[{ label: 'HTTP CONNECT', value: 'httpconnect' }]" :readonly="readonly" />
    </div>
  </template>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProxyFormData } from '../../types'
import ConfigField from '../ConfigField.vue'

const props = withDefaults(defineProps<{
  modelValue: ProxyFormData
  readonly?: boolean
  usedPorts?: number[]
  usedPortsError?: boolean
}>(), { readonly: false, usedPorts: () => [], usedPortsError: false })

const isEdit = computed(() => !!props.modelValue.name)

const emit = defineEmits<{ 'update:modelValue': [value: ProxyFormData] }>()

const form = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const portConflict = computed(() => {
  const port = form.value.remotePort
  if (!port || port === 0) return false
  return props.usedPorts.includes(port)
})
</script>

<style scoped lang="scss">
@use '@/assets/css/form-layout';

.used-ports-section {
  margin-top: 8px;
  padding: 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 4px;
}

.used-ports-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.used-ports-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-height: 120px;
  overflow-y: auto;
}

.port-tag {
  font-family: monospace;
}

.port-conflict-warning {
  margin-top: 8px;
}

.used-ports-hint {
  margin-top: 8px;
}
</style>
