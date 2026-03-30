import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)

/* Layout */
import Layout from '@/layout'

/**
 * Note: sub-menu only appear when route children.length >= 1
 * Detail see: https://panjiachen.github.io/vue-element-admin-site/guide/essentials/router-and-nav.html
 *
 * hidden: true                   if set true, item will not show in the sidebar(default is false)
 * alwaysShow: true               if set true, will always show the root menu
 *                                if not set alwaysShow, when item has more than one children route,
 *                                it will becomes nested mode, otherwise not show the root menu
 * redirect: noRedirect           if set noRedirect will no redirect in the breadcrumb
 * name:'router-name'             the name is used by <keep-alive> (must set!!!)
 * meta : {
    roles: ['admin','editor']    control the page roles (you can set multiple roles)
    title: 'title'               the name show in sidebar and breadcrumb (recommend set)
    icon: 'svg-name'             the icon show in the sidebar
    breadcrumb: false            if set false, the item will hidden in breadcrumb(default is true)
    activeMenu: '/example/list'  if set path, the sidebar will highlight the path you set
  }
 */

/**
 * constantRoutes
 * a base page that does not have permission requirements
 * all roles can be accessed
 */
export const constantRoutes = [{
  path: '/login',
  component: () => import('@/views/login/index'),
  hidden: true
},

{
  path: '/404',
  component: () => import('@/views/404'),
  hidden: true
},

{
  path: '/',
  component: Layout,
  redirect: '/main',
  children: [{
    path: 'main',
    name: 'main',
    component: () => import('@/views/mainPage/list/index'),
    meta: {
      title: '主页',
      icon: 'realestate'
    }
  }]
}
]

/**
 * asyncRoutes 异步路由
 * the routes that need to be dynamically loaded based on user roles   异步路由是根据用户角色动态加载的路由。
 * 每个路由对象包括路径（path）、组件（component）、重定向（redirect）、路由名称（name）、子路由（children）等元素
 */
export const asyncRoutes = [
  {
    path: '/account',
    component: Layout,
    redirect: '/account/all',
    name: 'Account',
    alwaysShow: true,
    meta: {
      roles: ['admin'],
      title: '账户',
      icon: 'donating'
    },
    children: [
      {
        path: 'all',
        name: 'AccountAll',
        component: () => import('@/views/account/list/index'),
        meta: {
          roles: ['admin'],
          title: '所有账户',
          icon: 'donatingAll'
        }
      },

      {
        path: 'add',
        name: 'Add',
        component: () => import('@/views/account/add/index'),
        meta: {
          roles: ['admin'],
          title: '新增账户',
          icon: 'addRealestate'
        }
      }
    ]
  },

  {
    path: '/prescription',
    component: Layout,
    redirect: '/prescription/all',
    name: 'Prescription',
    alwaysShow: true,
    meta: {
      roles: ['doctor','patient'],
      title: '病历',
      icon: 'donating'
    },
    children: [
      {
        path: 'all',
        name: 'PrescriptionAll',
        component: () => import('@/views/prescription/list/index'),
        meta: {
          roles: ['admin','doctor'],
          title: '所有病历',
          icon: 'donatingAll'
        }
      },
      {
        path: 'mine',
        name: 'PrescriptionOfMine',
        component: () => import('@/views/prescription/mine/index'),
        meta: {
          roles: ['admin','patient'],
          title: '我的病历',
          icon: 'donatingDonor'
        }
      },
      {
        path: 'accessible',
        name: 'PrescriptionAccessible',
        component: () => import('@/views/prescription/accessible/index'),
        meta: {
          roles: ['doctor'],
          title: '可访问病历',
          icon: 'donatingAll'
        }
      },
      {
        path: 'authorization',
        name: 'PrescriptionAuthorization',
        component: () => import('@/views/prescription/authorization/index'),
        meta: {
          roles: ['patient'],
          title: '病历授权管理',
          icon: 'donatingDonor'
        }
      },
      {
        path: 'add',
        name: 'Add',
        component: () => import('@/views/prescription/add/index'),
        meta: {
          roles: ['admin','doctor'],
          title: '新增病历',
          icon: 'addRealestate'
        }
      }
    ]
  },

  {
    path: '/insurance',
    component: Layout,
    redirect: '/insurance/all',
    name: 'Insurance',
    alwaysShow: true,
    meta: {
      roles: ['insurance'],
      title: '保险报销',
      icon: 'donating'
    },
    children: [{
      path: 'all',
      name: 'InsuranceAll',
      component: () => import('@/views/insuranceCover/list/index'),
      meta: {
        roles: ['admin','insurance'],
        title: '所有的报销记录',
        icon: 'donatingAll'
      }
    },
      {
        path: 'mine',
        name: 'InsuranceMine',
        component: () => import('@/views/insuranceCover/mine/index'),
        meta: {
          roles: ['admin','patient'],
          title: '已发起的报销',
          icon: 'donatingDonor'
        }
      },
      {
        path: 'add',
        name: 'Add',
        component: () => import('@/views/insuranceCover/add/index'),
        meta: {
          roles: ['admin','insurance'],
          title: '新增报销',
          icon: 'addRealestate'
        }
      },
      {
        path: 'patientAdd',
        name: 'PatientAdd',
        component: () => import('@/views/insuranceCover/patientAdd/index'),
        meta: {
          roles: ['admin','patient'],
          title: '申请报销',
          icon: 'addRealestate'
        }
      }
    ]
  },

  {
    path: '/drug',
    component: Layout,
    redirect: '/drug/all',
    name: 'Drug',
    alwaysShow: true,
    meta: {
      roles: ['drugstore'],
      title: '药品订单',
      icon: 'donating'
    },
    children: [{
      path: 'all',
      name: 'DrugAll',
      component: () => import('@/views/drugOrder/list/index'),
      meta: {
        roles: ['admin','drugstore'],
        title: '所有订单',
        icon: 'donatingAll'
      }
    },
      {
        path: 'mine',
        name: 'DrugMine',
        component: () => import('@/views/drugOrder/mine/index'),
        meta: {
          roles: ['admin', 'patient'],
          title: '我的订单',
          icon: 'donatingAll'
        }
      },
      {
        path: '/addDrug',
        name: 'AddDrug',
        component: () => import('@/views/drugOrder/add/index'),
        meta: {
          roles: ['admin','drugstore'],
          title: '新增订单',
          icon: 'addRealestate'
        }
      }]
  },

  {
    path: '/outpatient',
    component: Layout,
    redirect: '/outpatient/register',
    name: 'Outpatient',
    alwaysShow: true,
    meta: {
      roles: ['admin', 'doctor', 'patient'],
      title: '就诊',
      icon: 'donating'
    },
    children: [
      {
        path: 'register',
        name: 'OutpatientRegister',
        component: () => import('@/views/outpatient/register/index'),
        meta: { roles: ['admin', 'patient'], title: '挂号预约', icon: 'addRealestate' }
      },
      {
        path: 'my-registration',
        name: 'OutpatientMyRegistration',
        component: () => import('@/views/outpatient/my-registration/index'),
        meta: { roles: ['admin', 'patient'], title: '我的预约', icon: 'donatingAll' }
      },
      {
        path: 'payment',
        name: 'OutpatientPayment',
        component: () => import('@/views/outpatient/payment/index'),
        meta: { roles: ['admin', 'patient'], title: '缴费管理', icon: 'donatingDonor' }
      },
      {
        path: 'queue',
        name: 'OutpatientQueue',
        component: () => import('@/views/outpatient/queue/index'),
        meta: { roles: ['admin', 'patient'], title: '排队叫号', icon: 'donatingAll' }
      },
      {
        path: 'slot',
        name: 'OutpatientSlot',
        component: () => import('@/views/outpatient/slot/index'),
        meta: { roles: ['admin', 'doctor'], title: '号源管理', icon: 'addRealestate' }
      },
      {
        path: 'doctor-queue',
        name: 'OutpatientDoctorQueue',
        component: () => import('@/views/outpatient/doctor-queue/index'),
        meta: { roles: ['admin', 'doctor'], title: '就诊队列', icon: 'donatingAll' }
      },
      {
        path: 'statistics',
        name: 'OutpatientStatistics',
        component: () => import('@/views/outpatient/statistics/index'),
        meta: { roles: ['admin'], title: '门诊数据统计', icon: 'donatingAll' }
      }
    ]
  },

  {
    path: '/ai-health-assistant',
    component: Layout,
    children: [
      {
        path: '',
        name: 'AIHealthAssistant',
        component: () => import('@/views/ai-health-assistant/index'),
        meta: { roles: ['admin', 'doctor', 'patient'], title: 'AI健康助手', icon: 'donatingAll' }
      }
    ]
  },

  {
    path: '/audit',
    component: Layout,
    redirect: '/audit/search',
    name: 'Audit',
    alwaysShow: true,
    meta: {
      roles: ['admin'],
      title: '审计监控',
      icon: 'donating'
    },
    children: [
      {
        path: 'search',
        name: 'AuditSearch',
        component: () => import('@/views/audit/search/index'),
        meta: {
          roles: ['admin'],
          title: '审计检索',
          icon: 'donatingAll'
        }
      }
    ]
  },

  // 404 page must be placed at the end !!!
  {
    path: '*',
    redirect: '/404',
    hidden: true
  }
]

const createRouter = () => new Router({
  base: '/',
  // mode: 'history', // require service support
  scrollBehavior: () => ({
    y: 0
  }),
  routes: constantRoutes
})

const router = createRouter()

// Detail see: https://github.com/vuejs/vue-router/issues/1234#issuecomment-357941465
export function resetRouter() {
  const newRouter = createRouter()
  router.matcher = newRouter.matcher // reset router
}

export default router
