import { http } from './http'
import type {
  StatusResponse,
  ProxyListResp,
  ProxyDefinition,
  VisitorListResp,
  VisitorDefinition,
} from '../types'
import type {
  ProfileListResponse,
  ProfileDetailResponse,
  ProfileEntry,
} from '../types/profile'

export const getStatus = () => {
  return http.get<StatusResponse>('/api/status')
}

export const getConfig = () => {
  return http.get<string>('/api/config')
}

export const putConfig = (content: string) => {
  return http.put<void>('/api/config', content)
}

export const reloadConfig = () => {
  return http.get<void>('/api/reload')
}

// Config lookup API (any source)
export const getProxyConfig = (name: string) => {
  return http.get<ProxyDefinition>(
    `/api/proxy/${encodeURIComponent(name)}/config`,
  )
}

export const getVisitorConfig = (name: string) => {
  return http.get<VisitorDefinition>(
    `/api/visitor/${encodeURIComponent(name)}/config`,
  )
}

// Store API - Proxies
export const listStoreProxies = () => {
  return http.get<ProxyListResp>('/api/store/proxies')
}

export const getStoreProxy = (name: string) => {
  return http.get<ProxyDefinition>(
    `/api/store/proxies/${encodeURIComponent(name)}`,
  )
}

export const createStoreProxy = (config: ProxyDefinition) => {
  return http.post<ProxyDefinition>('/api/store/proxies', config)
}

export const updateStoreProxy = (name: string, config: ProxyDefinition) => {
  return http.put<ProxyDefinition>(
    `/api/store/proxies/${encodeURIComponent(name)}`,
    config,
  )
}

export const deleteStoreProxy = (name: string) => {
  return http.delete<void>(`/api/store/proxies/${encodeURIComponent(name)}`)
}

// Store API - Visitors
export const listStoreVisitors = () => {
  return http.get<VisitorListResp>('/api/store/visitors')
}

export const getStoreVisitor = (name: string) => {
  return http.get<VisitorDefinition>(
    `/api/store/visitors/${encodeURIComponent(name)}`,
  )
}

export const createStoreVisitor = (config: VisitorDefinition) => {
  return http.post<VisitorDefinition>('/api/store/visitors', config)
}

export const updateStoreVisitor = (
  name: string,
  config: VisitorDefinition,
) => {
  return http.put<VisitorDefinition>(
    `/api/store/visitors/${encodeURIComponent(name)}`,
    config,
  )
}

export const deleteStoreVisitor = (name: string) => {
  return http.delete<void>(`/api/store/visitors/${encodeURIComponent(name)}`)
}

// Profile API
export const listProfiles = () => {
  return http.get<ProfileListResponse>('/api/profiles')
}

export const getProfile = (name: string) => {
  return http.get<ProfileDetailResponse>(
    `/api/profiles/${encodeURIComponent(name)}`,
  )
}

export const createProfile = (data: ProfileEntry) => {
  return http.post<ProfileEntry>('/api/profiles', data)
}

export const updateProfile = (name: string, data: ProfileEntry) => {
  return http.put<ProfileEntry>(
    `/api/profiles/${encodeURIComponent(name)}`,
    data,
  )
}

export const deleteProfile = (name: string) => {
  return http.delete<void>(`/api/profiles/${encodeURIComponent(name)}`)
}

export const startProfile = (name: string) => {
  return http.post<void>(`/api/profiles/${encodeURIComponent(name)}/start`)
}

export const stopProfile = (name: string) => {
  return http.post<void>(`/api/profiles/${encodeURIComponent(name)}/stop`)
}

// Per-profile proxy running status
export const getProfileProxyStatus = (profile: string) => {
  return http.get<StatusResponse>(
    `/api/profiles/${encodeURIComponent(profile)}/status`,
  )
}

// Profile-scoped proxy CRUD
export const listProfileProxies = (profile: string) => {
  return http.get<ProxyListResp>(
    `/api/profiles/${encodeURIComponent(profile)}/proxies`,
  )
}

export const getProfileProxy = (profile: string, name: string) => {
  return http.get<ProxyDefinition>(
    `/api/profiles/${encodeURIComponent(profile)}/proxies/${encodeURIComponent(name)}`,
  )
}

export const createProfileProxy = (profile: string, config: ProxyDefinition) => {
  return http.post<ProxyDefinition>(
    `/api/profiles/${encodeURIComponent(profile)}/proxies`,
    config,
  )
}

export const updateProfileProxy = (
  profile: string,
  name: string,
  config: ProxyDefinition,
) => {
  return http.put<ProxyDefinition>(
    `/api/profiles/${encodeURIComponent(profile)}/proxies/${encodeURIComponent(name)}`,
    config,
  )
}

export const deleteProfileProxy = (profile: string, name: string) => {
  return http.delete<void>(
    `/api/profiles/${encodeURIComponent(profile)}/proxies/${encodeURIComponent(name)}`,
  )
}

// Profile-scoped visitor CRUD
export const listProfileVisitors = (profile: string) => {
  return http.get<VisitorListResp>(
    `/api/profiles/${encodeURIComponent(profile)}/visitors`,
  )
}

export const getProfileVisitor = (profile: string, name: string) => {
  return http.get<VisitorDefinition>(
    `/api/profiles/${encodeURIComponent(profile)}/visitors/${encodeURIComponent(name)}`,
  )
}

export const createProfileVisitor = (
  profile: string,
  config: VisitorDefinition,
) => {
  return http.post<VisitorDefinition>(
    `/api/profiles/${encodeURIComponent(profile)}/visitors`,
    config,
  )
}

export const updateProfileVisitor = (
  profile: string,
  name: string,
  config: VisitorDefinition,
) => {
  return http.put<VisitorDefinition>(
    `/api/profiles/${encodeURIComponent(profile)}/visitors/${encodeURIComponent(name)}`,
    config,
  )
}

export const deleteProfileVisitor = (profile: string, name: string) => {
  return http.delete<void>(
    `/api/profiles/${encodeURIComponent(profile)}/visitors/${encodeURIComponent(name)}`,
  )
}
