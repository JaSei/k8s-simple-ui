<template>
  <v-app>
    <v-toolbar app>
      <v-toolbar-title class="headline text-uppercase">
        <span>Kubernetes</span>
        <span class="font-weight-light">DASHBOARD</span>
      </v-toolbar-title>
    </v-toolbar>

    <v-content>
      <Dashboard/>
    </v-content>
  </v-app>
</template>

<script>
import Dashboard from './components/Dashboard'

export default {
  name: 'App',
  components: {
    Dashboard
  },
  data () {
    return {
      //
    }
  }
}
</script>
