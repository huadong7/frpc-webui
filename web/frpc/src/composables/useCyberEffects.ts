import { ref, onMounted, onUnmounted, watch, type Ref } from 'vue'
import { usePreferredReducedMotion } from '@vueuse/core'

export function useCyberEffects(_isDark: Ref<boolean>) {
  const mouseGlowRef = ref<HTMLElement | null>(null)
  const reducedMotion = usePreferredReducedMotion()

  let rafId = 0
  let mouseX = 0
  let mouseY = 0
  let currentX = 0
  let currentY = 0

  const onMouseMove = (e: MouseEvent) => {
    mouseX = e.clientX
    mouseY = e.clientY
  }

  const onMouseEnter = () => {
    mouseGlowRef.value?.classList.add('active')
  }

  const onMouseLeave = () => {
    mouseGlowRef.value?.classList.remove('active')
  }

  const animate = () => {
    if (reducedMotion.value) return

    currentX += (mouseX - currentX) * 0.08
    currentY += (mouseY - currentY) * 0.08

    if (mouseGlowRef.value) {
      mouseGlowRef.value.style.left = `${currentX}px`
      mouseGlowRef.value.style.top = `${currentY}px`
    }

    rafId = requestAnimationFrame(animate)
  }

  onMounted(() => {
    if (reducedMotion.value) return

    window.addEventListener('mousemove', onMouseMove, { passive: true })
    document.addEventListener('mouseenter', onMouseEnter)
    document.addEventListener('mouseleave', onMouseLeave)
    rafId = requestAnimationFrame(animate)
  })

  onUnmounted(() => {
    window.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseenter', onMouseEnter)
    document.removeEventListener('mouseleave', onMouseLeave)
    cancelAnimationFrame(rafId)
  })

  watch(reducedMotion, (val) => {
    if (val) {
      cancelAnimationFrame(rafId)
      if (mouseGlowRef.value) {
        mouseGlowRef.value.classList.remove('active')
      }
    } else {
      rafId = requestAnimationFrame(animate)
    }
  })

  return { mouseGlowRef }
}
