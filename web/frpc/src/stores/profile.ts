import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ProfileStatus, ProfileEntry } from '../types/profile'
import {
  listProfiles,
  getProfile,
  createProfile,
  updateProfile,
  deleteProfile,
  startProfile,
  stopProfile,
} from '../api/frpc'

export const useProfileStore = defineStore('profile', () => {
  const profiles = ref<ProfileStatus[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchProfiles = async () => {
    loading.value = true
    error.value = null
    try {
      const res = await listProfiles()
      profiles.value = res.profiles || []
    } catch (err: any) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  const createProfileEntry = async (data: ProfileEntry) => {
    await createProfile(data)
    await fetchProfiles()
  }

  const updateProfileEntry = async (name: string, data: ProfileEntry) => {
    await updateProfile(name, data)
    await fetchProfiles()
  }

  const deleteProfileEntry = async (name: string) => {
    await deleteProfile(name)
    await fetchProfiles()
  }

  const startProfileByName = async (name: string) => {
    await startProfile(name)
    await fetchProfiles()
  }

  const stopProfileByName = async (name: string) => {
    await stopProfile(name)
    await fetchProfiles()
  }

  const getProfileDetail = async (name: string) => {
    return getProfile(name)
  }

  return {
    profiles,
    loading,
    error,
    fetchProfiles,
    createProfileEntry,
    updateProfileEntry,
    deleteProfileEntry,
    startProfileByName,
    stopProfileByName,
    getProfileDetail,
  }
})
