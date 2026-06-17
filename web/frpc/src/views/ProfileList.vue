<template>
  <div class="profiles-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">Server Profiles</h2>
        <p class="page-subtitle">Manage connections to frps servers</p>
      </div>
      <ActionButton @click="handleCreate">+ Add Profile</ActionButton>
    </div>

    <div v-loading="profileStore.loading" class="profile-list">
      <div v-if="profileStore.profiles.length === 0 && !profileStore.loading" class="empty-state">
        <p class="empty-text">No profiles configured</p>
        <p class="empty-hint">Create a server profile to start managing proxies.</p>
      </div>

      <div
        v-for="p in profileStore.profiles"
        :key="p.name"
        class="profile-card"
      >
        <div class="card-main" @click="goToProfileProxies(p.name)">
          <div class="card-info">
            <span class="profile-name">{{ p.name }}</span>
            <span class="profile-server" v-if="p.serverAddr">
              {{ p.serverAddr }}
            </span>
          </div>
          <ProfileStatusBadge :status="p.status" />
        </div>

        <div class="card-actions" @click.stop>
          <el-button
            v-if="p.status !== 'running'"
            size="small"
            type="primary"
            @click="handleStart(p.name)"
          >
            Start
          </el-button>
          <el-button
            v-else
            size="small"
            type="warning"
            @click="handleStop(p.name)"
          >
            Stop
          </el-button>
          <el-button size="small" @click="goToEdit(p.name)">Edit</el-button>
          <el-popconfirm
            title="Delete this profile?"
            @confirm="handleDelete(p.name)"
          >
            <template #reference>
              <el-button size="small" type="danger" text>Delete</el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useProfileStore } from '../stores/profile'
import ProfileStatusBadge from '../components/ProfileStatusBadge.vue'
import ActionButton from '@shared/components/ActionButton.vue'

const router = useRouter()
const profileStore = useProfileStore()

const handleCreate = () => router.push('/profiles/create')
const goToEdit = (name: string) => router.push(`/profiles/${encodeURIComponent(name)}/edit`)
const goToProfileProxies = (name: string) => router.push(`/profiles/${encodeURIComponent(name)}/proxies`)

const handleStart = async (name: string) => {
  try {
    await profileStore.startProfileByName(name)
    ElMessage.success('Profile started')
  } catch (err: any) {
    ElMessage.error('Start failed: ' + (err.message || 'Unknown error'))
  }
}

const handleStop = async (name: string) => {
  try {
    await profileStore.stopProfileByName(name)
    ElMessage.success('Profile stopped')
  } catch (err: any) {
    ElMessage.error('Stop failed: ' + (err.message || 'Unknown error'))
  }
}

const handleDelete = async (name: string) => {
  try {
    await profileStore.deleteProfileEntry(name)
    ElMessage.success('Profile deleted')
  } catch (err: any) {
    ElMessage.error('Delete failed: ' + (err.message || 'Unknown error'))
  }
}

onMounted(() => {
  profileStore.fetchProfiles().catch(() => {})
})
</script>

<style scoped lang="scss">
.profiles-page {
  height: 100%;
  overflow-y: auto;
  padding: $spacing-xl 40px;
  max-width: 960px;
  margin: 0 auto;
  @include page-transition;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: $spacing-xl;

  .page-title {
    @include gradient-text;
  }
}

.profile-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.profile-card {
  @include glass-panel;
  border-radius: $radius-md;
  overflow: hidden;
  transition: all 0.3s ease;

  &:hover {
    transform: translateY(-2px);
    @include glow-border;
  }
}

.card-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-lg;
  cursor: pointer;
}

.card-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.profile-name {
  font-weight: $font-weight-semibold;
  font-size: $font-size-md;
  color: $color-text-primary;
  transition: color 0.2s ease;

  .profile-card:hover & {
    color: $accent-cyan;
  }
}

.profile-server {
  font-size: $font-size-sm;
  color: $color-text-muted;
  font-family: ui-monospace, monospace;
}

.card-actions {
  display: flex;
  gap: $spacing-sm;
  padding: 0 $spacing-lg $spacing-lg;
}

.empty-state {
  @include glass-panel;
  text-align: center;
  padding: 60px $spacing-xl;
  border-radius: $radius-md;
}

.empty-text {
  font-size: $font-size-lg;
  font-weight: $font-weight-medium;
  color: $color-text-secondary;
  margin: 0 0 $spacing-sm;
}

.empty-hint {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin: 0;
}

@include mobile {
  .profiles-page {
    padding: $spacing-xl $spacing-lg;
  }
}
</style>
