<template>
  <div class="status-pills">
    <button
      v-for="pill in pills"
      :key="pill.status"
      class="pill"
      :class="{ active: modelValue === pill.status, [pill.status || 'all']: true }"
      @click="emit('update:modelValue', pill.status)"
    >
      {{ pill.label }} {{ pill.count }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  items: Array<{ status: string }>
  modelValue: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const pills = computed(() => {
  const counts = { running: 0, error: 0, waiting: 0 }
  for (const item of props.items) {
    const s = item.status as keyof typeof counts
    if (s in counts) {
      counts[s]++
    }
  }
  return [
    { status: '', label: 'All', count: props.items.length },
    { status: 'running', label: 'Running', count: counts.running },
    { status: 'error', label: 'Error', count: counts.error },
    { status: 'waiting', label: 'Waiting', count: counts.waiting },
  ]
})
</script>

<style scoped lang="scss">
.status-pills {
  display: flex;
  gap: $spacing-sm;
}

.pill {
  border: 1px solid transparent;
  border-radius: 12px;
  padding: $spacing-xs $spacing-md;
  font-size: $font-size-xs;
  font-weight: $font-weight-medium;
  cursor: pointer;
  background: rgba(0, 212, 255, 0.04);
  color: $color-text-secondary;
  transition: all 0.2s ease;
  white-space: nowrap;

  &:hover {
    border-color: rgba(0, 212, 255, 0.15);
  }

  &.active {
    &.all {
      background: rgba(0, 212, 255, 0.08);
      color: $accent-cyan;
      border-color: rgba(0, 212, 255, 0.2);
      box-shadow: 0 0 8px rgba(0, 212, 255, 0.15);
    }

    &.running {
      background: rgba(0, 255, 136, 0.08);
      color: #00ff88;
      border-color: rgba(0, 255, 136, 0.2);
      box-shadow: 0 0 8px rgba(0, 255, 136, 0.15);
    }

    &.error {
      background: rgba(255, 68, 102, 0.08);
      color: #ff4466;
      border-color: rgba(255, 68, 102, 0.2);
      box-shadow: 0 0 8px rgba(255, 68, 102, 0.15);
    }

    &.waiting {
      background: rgba(255, 170, 0, 0.08);
      color: #ffaa00;
      border-color: rgba(255, 170, 0, 0.2);
      box-shadow: 0 0 8px rgba(255, 170, 0, 0.15);
    }
  }
}

@include mobile {
  .status-pills {
    overflow-x: auto;
    flex-wrap: nowrap;
    scrollbar-width: none;
    -ms-overflow-style: none;

    &::-webkit-scrollbar {
      display: none;
    }
  }
}
</style>
