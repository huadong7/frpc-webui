<template>
  <div class="server-proxies-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">Server Proxies</h2>
        <p class="page-subtitle">Proxies registered on frps servers across all profiles</p>
      </div>
      <el-button
        type="primary"
        :loading="loading"
        @click="fetchAll"
      >
        <el-icon v-if="!loading" class="el-icon--left"><Refresh /></el-icon>
        Refresh
      </el-button>
    </div>

    <div v-if="!profileStore.profiles.length && !profileStore.loading" class="empty-state">
      <p class="empty-text">No profiles configured</p>
      <p class="empty-hint">Add a server profile first to view its proxies.</p>
    </div>

    <div v-else v-loading="loading" class="profile-sections">
      <div
        v-for="p in profileStore.profiles"
        :key="p.name"
        class="profile-section"
      >
        <div class="section-header">
          <div class="section-info">
            <span class="section-name">{{ p.name }}</span>
            <span class="section-server" v-if="p.serverAddr">
              {{ p.serverAddr }}:{{ p.serverPort }}
            </span>
          </div>
          <div class="section-meta">
            <el-radio-group
              v-if="serverProxies[p.name] && serverProxies[p.name].length > 0"
              v-model="statusFilter[p.name]"
              size="small"
            >
              <el-radio-button value="all">All</el-radio-button>
              <el-radio-button value="online">Online</el-radio-button>
              <el-radio-button value="offline">Offline</el-radio-button>
            </el-radio-group>
            <span v-if="serverProxies[p.name]" class="proxy-count">
              {{ sortedProxies(p.name).length }} proxies
            </span>
            <span v-if="errors[p.name]" class="section-error">
              {{ errors[p.name] }}
            </span>
          </div>
        </div>

        <div v-if="serverProxies[p.name] && serverProxies[p.name].length > 0" class="proxy-table-wrapper">
          <table class="proxy-table">
            <thead>
              <tr>
                <th
                  v-for="col in columns"
                  :key="col.key"
                  :class="{ sortable: col.sortable, active: sortState[p.name]?.key === col.key }"
                  @click="col.sortable && toggleSort(p.name, col.key)"
                >
                  <span class="th-content">
                    {{ col.label }}
                    <span v-if="col.sortable" class="sort-icon">
                      <span v-if="sortState[p.name]?.key === col.key">
                        {{ sortState[p.name].order === 'asc' ? '▲' : '▼' }}
                      </span>
                      <span v-else class="sort-inactive">⇅</span>
                    </span>
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="proxy in sortedProxies(p.name)" :key="proxy.name">
                <td class="proxy-name">{{ proxy.name }}</td>
                <td>
                  <el-tag size="small" :type="typeTagColor(proxy.type)">{{ proxy.type }}</el-tag>
                </td>
                <td class="proxy-addr">{{ formatAddress(proxy) }}</td>
                <td>
                  <span class="status-pill" :class="proxy.status === 'online' ? 'running' : 'disabled'">
                    <span class="status-dot"></span>
                    {{ proxy.status }}
                  </span>
                </td>
                <td>{{ proxy.curConns }}</td>
                <td>{{ formatTraffic(proxy.todayTrafficIn) }}</td>
                <td>{{ formatTraffic(proxy.todayTrafficOut) }}</td>
                <td class="client-id">{{ proxy.clientID || '-' }}</td>
                <td>{{ proxy.user || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else-if="!errors[p.name] && !loading" class="no-proxies">
          No proxies found on this server
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { useProfileStore } from '../stores/profile'
import { getProfileServerProxies } from '../api/frpc'
import type { ServerProxyInfo } from '../api/frpc'

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'type', label: 'Type', sortable: true },
  { key: 'address', label: 'Address', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'curConns', label: 'Conns', sortable: true },
  { key: 'todayTrafficIn', label: 'Traffic In', sortable: true },
  { key: 'todayTrafficOut', label: 'Traffic Out', sortable: true },
  { key: 'clientID', label: 'Client ID', sortable: false },
  { key: 'user', label: 'User', sortable: false },
]

const profileStore = useProfileStore()
const loading = ref(false)
const serverProxies = ref<Record<string, ServerProxyInfo[]>>({})
const errors = ref<Record<string, string>>({})
const statusFilter = reactive<Record<string, string>>({})
const sortState = reactive<Record<string, { key: string; order: 'asc' | 'desc' }>>({})

const fetchAll = async () => {
  loading.value = true
  serverProxies.value = {}
  errors.value = {}

  if (!profileStore.profiles.length) {
    await profileStore.fetchProfiles()
  }

  const tasks = profileStore.profiles.map(async (p) => {
    statusFilter[p.name] = 'all'
    sortState[p.name] = { key: 'address', order: 'asc' }
    try {
      const resp = await getProfileServerProxies(p.name)
      serverProxies.value[p.name] = resp.proxies || []
    } catch (err: any) {
      errors.value[p.name] = err.message || 'Failed to load'
      serverProxies.value[p.name] = []
    }
  })

  await Promise.allSettled(tasks)
  loading.value = false
}

const toggleSort = (profileName: string, key: string) => {
  const current = sortState[profileName]
  if (current && current.key === key) {
    sortState[profileName] = { key, order: current.order === 'asc' ? 'desc' : 'asc' }
  } else {
    sortState[profileName] = { key, order: 'asc' }
  }
}

const getSortValue = (proxy: ServerProxyInfo, key: string): [number, string] => {
  switch (key) {
    case 'name': return [0, proxy.name]
    case 'type': return [0, proxy.type]
    case 'address': {
      if (proxy.remotePort) return [1, String(proxy.remotePort)]
      const sub = proxy.subDomain || proxy.customDomains || ''
      return [2, sub]
    }
    case 'status': return [0, proxy.status]
    case 'curConns': return [proxy.curConns, '']
    case 'todayTrafficIn': return [proxy.todayTrafficIn, '']
    case 'todayTrafficOut': return [proxy.todayTrafficOut, '']
    default: return [0, '']
  }
}

const sortedProxies = (profileName: string): ServerProxyInfo[] => {
  const list = serverProxies.value[profileName] || []
  const filter = statusFilter[profileName] || 'all'
  const sort = sortState[profileName] || { key: 'address', order: 'asc' }

  let filtered = list
  if (filter !== 'all') {
    filtered = list.filter((p) => p.status === filter)
  }

  return [...filtered].sort((a, b) => {
    if (a.status === 'offline' && b.status !== 'offline') return 1
    if (a.status !== 'offline' && b.status === 'offline') return -1

    const [na, sa] = getSortValue(a, sort.key)
    const [nb, sb] = getSortValue(b, sort.key)

    let cmp = na - nb
    if (cmp === 0) cmp = sa.localeCompare(sb)

    return sort.order === 'asc' ? cmp : -cmp
  })
}

const formatAddress = (proxy: ServerProxyInfo): string => {
  if (proxy.remotePort) {
    return `:${proxy.remotePort}`
  }
  if (proxy.type === 'http' || proxy.type === 'https' || proxy.type === 'tcpmux') {
    const parts: string[] = []
    if (proxy.subDomain) parts.push(`${proxy.subDomain}.*`)
    if (proxy.customDomains) parts.push(proxy.customDomains)
    return parts.join(', ') || '-'
  }
  return '-'
}

const formatTraffic = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

const typeTagColor = (type: string): string => {
  const colors: Record<string, string> = {
    tcp: '',
    udp: 'warning',
    http: 'success',
    https: 'success',
    tcpmux: 'info',
    stcp: 'info',
    xtcp: 'danger',
    sudp: 'warning',
  }
  return colors[type] || ''
}

onMounted(() => {
  fetchAll()
})
</script>

<style scoped lang="scss">
.server-proxies-page {
  height: 100%;
  overflow-y: auto;
  padding: $spacing-xl 40px;
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: $spacing-xl;
}

.profile-sections {
  display: flex;
  flex-direction: column;
  gap: $spacing-xl;
}

.profile-section {
  border: 1px solid $color-border-light;
  border-radius: $radius-md;
  overflow: hidden;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-md $spacing-lg;
  background: $color-bg-secondary;
  border-bottom: 1px solid $color-border-light;
}

.section-info {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.section-name {
  font-weight: $font-weight-semibold;
  font-size: $font-size-md;
  color: $color-text-primary;
}

.section-server {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.section-meta {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.proxy-count {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.section-error {
  font-size: $font-size-sm;
  color: #e6a23c;
}

.proxy-table-wrapper {
  overflow-x: auto;
}

.proxy-table {
  width: 100%;
  border-collapse: collapse;

  th, td {
    padding: $spacing-sm $spacing-md;
    text-align: left;
    border-bottom: 1px solid $color-border-light;
    font-size: $font-size-sm;
    white-space: nowrap;
  }

  th {
    font-weight: $font-weight-medium;
    color: $color-text-muted;
    background: $color-bg-secondary;

    &.sortable {
      cursor: pointer;
      user-select: none;
      transition: color $transition-fast;

      &:hover {
        color: $color-text-primary;
      }

      &.active {
        color: $color-text-primary;
      }
    }
  }

  td {
    color: $color-text-primary;
  }

  tr:last-child td {
    border-bottom: none;
  }

  tr:hover td {
    background: $color-bg-hover;
  }
}

.th-content {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.sort-icon {
  font-size: 10px;
  line-height: 1;
}

.sort-inactive {
  opacity: 0.3;
}

.proxy-name {
  font-weight: $font-weight-medium;
}

.proxy-addr {
  color: $color-text-primary;
}

.client-id {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  color: $color-text-muted;
}

.no-proxies {
  padding: $spacing-xl;
  text-align: center;
  color: $color-text-muted;
  font-size: $font-size-sm;
}

.empty-state {
  text-align: center;
  padding: 60px $spacing-xl;
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
  .server-proxies-page {
    padding: $spacing-xl $spacing-lg;
  }

  .proxy-table {
    font-size: $font-size-xs;

    th, td {
      padding: $spacing-xs $spacing-sm;
    }
  }
}
</style>
