import type { ProxyDefinition } from './proxy-store'

export interface DashboardConfig {
  addr?: string
  user?: string
  password?: string
}

export interface ProfileConfig {
  name: string
  serverAddr: string
  serverPort: number
  auth: {
    method: string
    token?: string
    oidc?: Record<string, unknown>
    additionalScopes?: string[]
  }
  transport: {
    protocol?: string
    poolCount?: number
    tcpMux?: boolean
    heartbeatInterval?: number
    heartbeatTimeout?: number
    dialServerTimeout?: number
    tls?: {
      enable?: boolean
      certFile?: string
      keyFile?: string
    }
  }
  user?: string
  clientID?: string
  loginFailExit?: boolean
  start?: string[]
  metadatas?: Record<string, string>
  dnsServer?: string
  udpPacketSize?: number
  autoStart: boolean
  dashboard?: DashboardConfig
}

export interface ProfileEntry {
  config: ProfileConfig
  proxies?: ProxyDefinition[]
  visitors?: VisitorDefRef[]
}

export interface ProfileStatus {
  name: string
  status: 'running' | 'stopped' | 'error'
  runID?: string
  error?: string
  serverAddr?: string
  serverPort?: number
}

export interface ProfileListResponse {
  profiles: ProfileStatus[]
}

export interface ProfileDetailResponse {
  config: ProfileConfig
  status: string
  runID?: string
  error?: string
  proxies: ProxyDefinition[]
  visitors: VisitorDefRef[]
}

// Lightweight visitor reference in profile responses
interface VisitorDefRef {
  name: string
  type: string
  [key: string]: unknown
}
