import { ref, onMounted, onUnmounted, isRef, unref } from 'vue'
import axios from 'axios'

export function useDisScroll() {
  onMounted(() => {
    document.body.style.overflow = "hidden"
    document.body.style.pointerEvents = "none"
    document.addEventListener("touchmove", (e) => { e.preventDefault() }, false)
  })
  onUnmounted(() => {
    document.body.style.removeProperty("overflow")
    document.body.style.removeProperty("pointer-events")
    document.removeEventListener("touchmove", (e) => { e.preventDefault() }, false)
  })
}

export function useScrollOff(enable) {
  function setScroll(flag) {
    if (unref(flag) == true) {
      document.body.style.overflow = "hidden"
      document.body.style.pointerEvents = "none"
      document.addEventListener("touchmove", (e) => { e.preventDefault() }, false)
    } else {
      document.body.style.removeProperty("overflow")
      document.body.style.removeProperty("pointer-events")
      document.removeEventListener("touchmove", (e) => { e.preventDefault() }, false)
    }
  }

  onMounted(() => setScroll(false))
  onUnmounted(() => setScroll(false))

  if (isRef(enable)) {
    watchEffect(setScroll)
  } else {
    setScroll(enable)
  }
}


export function useGetURLQuery(key) {
  let currentURL = location.href
  var queryPart = currentURL.split("?")[1]
  if (queryPart==undefined) return "/admin#/"
  var querys = queryPart.split("&")
  var result="/admin#/"
  querys.forEach((item, index, querys) => {
    if (item.split("=")[0] == key) {
      result=item.split("=")[1]
    }
  });
  return result
}