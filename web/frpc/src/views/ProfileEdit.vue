<template>
  <div class="profile-edit-page">
    <div class="page-header">
      <h2 class="page-title">{{ isEdit ? 'Edit Profile' : 'Create Profile' }}</h2>
    </div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="160px"
      label-position="top"
      class="profile-form"
    >
      <!-- Basic Section -->
      <ConfigSection title="Basic" :expanded="true">
        <el-form-item label="Profile Name" prop="name">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="e.g., office, home, production" />
        </el-form-item>
      </ConfigSection>

      <!-- Server Connection -->
      <ConfigSection title="Server Connection" :expanded="true">
        <el-form-item label="Server Address" prop="serverAddr">
          <el-input v-model="form.serverAddr" placeholder="e.g., 192.168.1.100" />
        </el-form-item>
        <el-form-item label="Server Port" prop="serverPort">
          <el-input-number v-model="form.serverPort" :min="1" :max="65535" placeholder="7000" />
        </el-form-item>
        <el-form-item label="Protocol">
          <el-select v-model="form.transport.protocol">
            <el-option label="tcp" value="tcp" />
            <el-option label="kcp" value="kcp" />
            <el-option label="quic" value="quic" />
            <el-option label="websocket" value="websocket" />
            <el-option label="wss" value="wss" />
          </el-select>
        </el-form-item>
      </ConfigSection>

      <!-- Authentication -->
      <ConfigSection title="Authentication" :expanded="false">
        <el-form-item label="Auth Method">
          <el-select v-model="form.auth.method">
            <el-option label="token" value="token" />
            <el-option label="oidc" value="oidc" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.auth.method === 'token'" label="Token">
          <el-input v-model="form.auth.token" type="password" show-password placeholder="Authentication token" />
        </el-form-item>
      </ConfigSection>

      <!-- Transport -->
      <ConfigSection title="Transport Options" :expanded="false">
        <el-form-item label="TCP Multiplexing">
          <el-switch v-model="form.transport.tcpMux" />
        </el-form-item>
        <el-form-item label="Connection Pool">
          <el-input-number v-model="form.transport.poolCount" :min="0" :max="100" />
        </el-form-item>
        <el-form-item label="Heartbeat Interval (s)">
          <el-input-number v-model="form.transport.heartbeatInterval" :min="-1" :max="3600" />
        </el-form-item>
      </ConfigSection>

      <!-- Identity -->
      <ConfigSection title="Identity" :expanded="false">
        <el-form-item label="User">
          <el-input v-model="form.user" placeholder="Optional user prefix for proxy names" />
        </el-form-item>
        <el-form-item label="Client ID">
          <el-input v-model="form.clientID" placeholder="Optional custom client ID" />
        </el-form-item>
      </ConfigSection>

      <!-- Auto Start -->
      <ConfigSection title="Startup" :expanded="false">
        <el-form-item label="Auto Start">
          <el-switch v-model="form.autoStart" />
          <span class="form-tip">Automatically start this profile when frpc launches</span>
        </el-form-item>
      </ConfigSection>

      <!-- Frps Dashboard -->
      <ConfigSection v-if="form.dashboard" title="Frps Dashboard (Optional)" :expanded="false">
        <el-form-item label="Dashboard Address">
          <el-input v-model="form.dashboard.addr" placeholder="e.g., http://192.168.1.100:7500" />
          <span class="form-tip">Used to query port usage from frps server</span>
        </el-form-item>
        <el-form-item label="Dashboard User">
          <el-input v-model="form.dashboard.user" placeholder="Dashboard username" />
        </el-form-item>
        <el-form-item label="Dashboard Password">
          <el-input v-model="form.dashboard.password" type="password" show-password placeholder="Dashboard password" />
        </el-form-item>
      </ConfigSection>

      <div class="form-actions">
        <el-button @click="handleCancel">Cancel</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ isEdit ? 'Save Changes' : 'Create Profile' }}
        </el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ConfigSection from '../components/ConfigSection.vue'
import { useProfileStore } from '../stores/profile'
import type { ProfileEntry, ProfileConfig } from '../types/profile'

const route = useRoute()
const router = useRouter()
const profileStore = useProfileStore()

const isEdit = computed(() => !!route.params.name)
const profileName = computed(() => route.params.name as string | undefined)
const saving = ref(false)
const formRef = ref()
const existingProxies = ref<any[]>([])
const existingVisitors = ref<any[]>([])

const defaultForm = (): ProfileConfig => ({
  name: '',
  serverAddr: '',
  serverPort: 7000,
  auth: { method: 'token', token: '' },
  transport: {
    protocol: 'tcp',
    poolCount: 1,
    tcpMux: true,
    heartbeatInterval: 30,
    heartbeatTimeout: 90,
  },
  user: '',
  clientID: '',
  autoStart: true,
  dashboard: { addr: '', user: '', password: '' },
})

const form = reactive<ProfileConfig>(defaultForm())

const rules = {
  name: [{ required: true, message: 'Profile name is required', trigger: 'blur' }],
  serverAddr: [{ required: true, message: 'Server address is required', trigger: 'blur' }],
  serverPort: [{ required: true, message: 'Server port is required', trigger: 'blur' }],
}

onMounted(async () => {
  if (isEdit.value && profileName.value) {
    try {
      const detail = await profileStore.getProfileDetail(profileName.value)
      Object.assign(form, detail.config)
      existingProxies.value = detail.proxies || []
      existingVisitors.value = detail.visitors || []
    } catch (err: any) {
      ElMessage.error('Failed to load profile: ' + (err.message || 'Unknown error'))
      router.replace('/profiles')
    }
  }
})

const handleSave = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const entry: ProfileEntry = {
      config: { ...form },
      proxies: isEdit.value && profileName.value ? existingProxies.value : [],
      visitors: isEdit.value && profileName.value ? existingVisitors.value : [],
    }

    if (isEdit.value && profileName.value) {
      await profileStore.updateProfileEntry(profileName.value, entry)
      ElMessage.success('Profile updated')
    } else {
      await profileStore.createProfileEntry(entry)
      ElMessage.success('Profile created')
    }
    router.replace('/profiles')
  } catch (err: any) {
    ElMessage.error('Failed: ' + (err.message || 'Unknown error'))
  } finally {
    saving.value = false
  }
}

const handleCancel = () => {
  router.back()
}
</script>

<style scoped lang="scss">
.profile-edit-page {
  height: 100%;
  overflow-y: auto;
  padding: $spacing-xl 40px;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: $spacing-xl;
}

.profile-form {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.form-tip {
  margin-left: $spacing-md;
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: $spacing-md;
  padding: $spacing-xl 0;
  border-top: 1px solid $color-border-light;
  margin-top: $spacing-lg;
}

@include mobile {
  .profile-edit-page {
    padding: $spacing-xl $spacing-lg;
  }
}
</style>
