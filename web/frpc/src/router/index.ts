import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import ProfileList from '../views/ProfileList.vue'
import ProfileEdit from '../views/ProfileEdit.vue'
import ProxyDetail from '../views/ProxyDetail.vue'
import ProxyEdit from '../views/ProxyEdit.vue'
import ProxyList from '../views/ProxyList.vue'
import VisitorDetail from '../views/VisitorDetail.vue'
import VisitorEdit from '../views/VisitorEdit.vue'
import VisitorList from '../views/VisitorList.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: Dashboard,
    },
    // Profiles
    {
      path: '/profiles',
      name: 'ProfileList',
      component: ProfileList,
    },
    {
      path: '/profiles/create',
      name: 'ProfileCreate',
      component: ProfileEdit,
    },
    {
      path: '/profiles/:name/edit',
      name: 'ProfileEdit',
      component: ProfileEdit,
    },
    // Profile-scoped proxies
    {
      path: '/profiles/:name/proxies',
      name: 'ProfileProxyList',
      component: ProxyList,
    },
    {
      path: '/profiles/:profile/proxies/detail/:name',
      name: 'ProfileProxyDetail',
      component: ProxyDetail,
    },
    {
      path: '/profiles/:profile/proxies/create',
      name: 'ProfileProxyCreate',
      component: ProxyEdit,
    },
    {
      path: '/profiles/:profile/proxies/:name/edit',
      name: 'ProfileProxyEdit',
      component: ProxyEdit,
    },
    // Profile-scoped visitors
    {
      path: '/profiles/:name/visitors',
      name: 'ProfileVisitorList',
      component: VisitorList,
    },
    {
      path: '/profiles/:profile/visitors/detail/:name',
      name: 'ProfileVisitorDetail',
      component: VisitorDetail,
    },
    {
      path: '/profiles/:profile/visitors/create',
      name: 'ProfileVisitorCreate',
      component: VisitorEdit,
    },
    {
      path: '/profiles/:profile/visitors/:name/edit',
      name: 'ProfileVisitorEdit',
      component: VisitorEdit,
    },
    // Legacy (flat) routes for backward compat
    {
      path: '/proxies',
      name: 'ProxyList',
      component: ProxyList,
    },
    {
      path: '/proxies/detail/:name',
      name: 'ProxyDetail',
      component: ProxyDetail,
    },
    {
      path: '/proxies/create',
      name: 'ProxyCreate',
      component: ProxyEdit,
    },
    {
      path: '/proxies/:name/edit',
      name: 'ProxyEdit',
      component: ProxyEdit,
    },
    {
      path: '/visitors',
      name: 'VisitorList',
      component: VisitorList,
    },
    {
      path: '/visitors/detail/:name',
      name: 'VisitorDetail',
      component: VisitorDetail,
    },
    {
      path: '/visitors/create',
      name: 'VisitorCreate',
      component: VisitorEdit,
    },
    {
      path: '/visitors/:name/edit',
      name: 'VisitorEdit',
      component: VisitorEdit,
    },
    {
      path: '/config',
      redirect: '/profiles',
    },
  ],
})

router.beforeEach(async (to) => {
  // Skip guards for profile management pages (always accessible)
  if (to.path.startsWith('/profiles') && !to.path.includes('/proxies') && !to.path.includes('/visitors')) {
    return true
  }
  return true
})

export default router
