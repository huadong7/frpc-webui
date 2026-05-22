<template>
  <span class="profile-status-badge" :class="status">
    <span class="status-dot" />
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: 'running' | 'stopped' | 'error'
}>()

const label = computed(() => {
  switch (props.status) {
    case 'running': return 'Running'
    case 'stopped': return 'Stopped'
    case 'error': return 'Error'
    default: return props.status
  }
})
</script>

<style scoped lang="scss">
.profile-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  padding: 3px 10px;
  border-radius: 10px;
  text-transform: capitalize;

  &.running {
    background: rgba(103, 194, 58, 0.1);
    color: #67c23a;
  }

  &.error {
    background: rgba(245, 108, 108, 0.1);
    color: #f56c6c;
  }

  &.stopped {
    background: $color-bg-muted;
    color: $color-text-light;
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
  }
}
</style>
