<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h1 class="page-title">Dashboard</h1>
      <p class="page-subtitle">Manage your frp client connections and proxies</p>
    </div>

    <!-- Getting Started: no profiles -->
    <div v-if="!profileStore.profiles.length && !profileStore.loading" class="getting-started">
      <div class="start-card">
        <h3>Welcome to frpc Manager</h3>
        <p>Get started by adding your first server profile. A profile defines the connection to a frps server.</p>
        <el-button type="primary" @click="goToCreateProfile">
          + Add Server Profile
        </el-button>
      </div>
    </div>

    <!-- Profile Overview -->
    <div v-else class="profile-overview">
      <div class="section-header">
        <h3>Server Profiles</h3>
        <el-button size="small" @click="goToCreateProfile">+ Add Profile</el-button>
      </div>

      <div v-loading="profileStore.loading" class="profile-cards">
        <div
          v-for="p in profileStore.profiles"
          :key="p.name"
          class="profile-card"
          @click="goToProfile(p.name)"
        >
          <div class="profile-card-header">
            <span class="profile-name">{{ p.name }}</span>
            <ProfileStatusBadge :status="p.status" />
          </div>
          <div class="profile-card-body">
            <div v-if="p.status === 'error' && p.error" class="profile-error">
              {{ p.error }}
            </div>
            <div v-if="p.runID" class="profile-runid">
              Run ID: {{ p.runID }}
            </div>
            <el-button
              v-if="p.status === 'running'"
              size="small"
              type="primary"
              plain
              @click.stop="goToCreateProxy(p.name)"
            >
              + New Proxy
            </el-button>
          </div>
          <div class="profile-card-actions" @click.stop>
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
              @click="handleStop(p.name)"
            >
              Stop
            </el-button>
            <el-button size="small" @click="goToProfileEdit(p.name)">
              Edit
            </el-button>
          </div>
        </div>
      </div>

      <div class="quick-links">
        <h3>Quick Links</h3>
        <div class="link-cards">
          <router-link to="/proxies" class="link-card">
            <span class="link-title">Proxies</span>
            <span class="link-desc">Manage proxy tunnels</span>
          </router-link>
          <router-link to="/visitors" class="link-card">
            <span class="link-title">Visitors</span>
            <span class="link-desc">Manage visitor connections</span>
          </router-link>
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

const router = useRouter()
const profileStore = useProfileStore()

const goToCreateProfile = () => router.push('/profiles/create')
const goToProfile = (name: string) => router.push(`/profiles/${encodeURIComponent(name)}/proxies`)
const goToProfileEdit = (name: string) => router.push(`/profiles/${encodeURIComponent(name)}/edit`)
const goToCreateProxy = (name: string) => router.push(`/profiles/${encodeURIComponent(name)}/proxies/create`)

const handleStart = async (name: string) => {
  try {
    await profileStore.startProfileByName(name)
    ElMessage.success(`Profile "${name}" started`)
  } catch (err: any) {
    ElMessage.error('Start failed: ' + (err.message || 'Unknown error'))
  }
}

const handleStop = async (name: string) => {
  try {
    await profileStore.stopProfileByName(name)
    ElMessage.success(`Profile "${name}" stopped`)
  } catch (err: any) {
    ElMessage.error('Stop failed: ' + (err.message || 'Unknown error'))
  }
}

onMounted(() => {
  profileStore.fetchProfiles().catch(() => {})
})
</script>

<style scoped lang="scss">
.dashboard-page {
  height: 100%;
  overflow-y: auto;
  padding: $spacing-xl 40px;
  max-width: 960px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 32px;
}

.getting-started {
  display: flex;
  justify-content: center;
  padding-top: 80px;
}

.start-card {
  text-align: center;
  padding: 48px;
  border: 1px solid $color-border-light;
  border-radius: $radius-md;
  max-width: 480px;

  h3 {
    margin: 0 0 $spacing-md;
    font-size: $font-size-xl;
    color: $color-text-primary;
  }

  p {
    margin: 0 0 $spacing-xl;
    color: $color-text-muted;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $spacing-lg;

  h3 {
    margin: 0;
    font-size: $font-size-lg;
    color: $color-text-primary;
  }
}

.profile-cards {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.profile-card {
  border: 1px solid $color-border-light;
  border-radius: $radius-md;
  padding: $spacing-lg;
  cursor: pointer;
  transition: border-color $transition-fast, box-shadow $transition-fast;

  &:hover {
    border-color: $color-border;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }
}

.profile-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $spacing-sm;
}

.profile-name {
  font-weight: $font-weight-semibold;
  font-size: $font-size-md;
  color: $color-text-primary;
}

.profile-card-body {
  margin-bottom: $spacing-md;
}

.profile-error {
  color: #f56c6c;
  font-size: $font-size-sm;
}

.profile-runid {
  color: $color-text-muted;
  font-size: $font-size-xs;
}

.profile-card-actions {
  display: flex;
  gap: $spacing-sm;
}

.quick-links {
  margin-top: 32px;

  h3 {
    margin: 0 0 $spacing-lg;
    font-size: $font-size-lg;
    color: $color-text-primary;
  }
}

.link-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-md;
}

.link-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: $spacing-lg;
  border: 1px solid $color-border-light;
  border-radius: $radius-md;
  text-decoration: none;
  transition: border-color $transition-fast;

  &:hover {
    border-color: $color-border;
  }

  .link-title {
    font-weight: $font-weight-medium;
    color: $color-text-primary;
  }

  .link-desc {
    font-size: $font-size-sm;
    color: $color-text-muted;
  }
}

@include mobile {
  .dashboard-page {
    padding: $spacing-xl $spacing-lg;
  }

  .link-cards {
    grid-template-columns: 1fr;
  }
}
</style>
