import Vue from 'vue'
import './plugins/vuetify'
import App from './App.vue'
import VueRouter from 'vue-router'
import Dashboard from './components/Dashboard'

Vue.config.productionTip = false

Vue.use(VueRouter)

const router = new VueRouter({
  mode: 'hash',
  base: __dirname,
  routes: [
    { path: '/',           component: Dashboard },
    { path: '/:namespace', component: Dashboard },
  ]
})

new Vue({
  render: h => h(App),
  router,
}).$mount('#app')
