<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder"
    :disabled="!profiles.length"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <el-option
      v-for="p in profiles"
      :key="p.name"
      :label="p.name"
      :value="p.name"
    >
      <span style="float: left">{{ p.name }}</span>
      <span style="float: right; font-size: 12px; color: var(--el-text-color-secondary)">
        {{ p.serverAddr || '' }}
      </span>
    </el-option>
  </el-select>
</template>

<script setup lang="ts">
import type { ProfileStatus } from '../types/profile'

defineProps<{
  modelValue: string
  profiles: ProfileStatus[]
  placeholder?: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
