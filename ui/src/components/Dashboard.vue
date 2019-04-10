<template>
  <div>
    <div>
      <v-alert :value="error" type="error">{{error.message}}</v-alert>
      <v-switch
        class="text-lg-right"
        :label="`Production namespaces only: ${prodOnly.toString()}`"
        v-model="prodOnly"
        v-on:change="fetchData"
      ></v-switch>
    </div>

      <v-tabs
        slot="extension"
        v-model="active"
        v-on:change="loadNamespace"
      >
        <v-tab
          v-for="(namespace, n) in namespaces"
          :key="n"
          :to="namespace"
          ripple
        >
          {{ namespace }}
        </v-tab>
      </v-tabs>

    <v-tabs-items v-model="tabs">
      <v-tab-item
        v-for="(namespace, n) in namespaces"
        :key="n"
      >
        <v-card>
          <v-card-text>
            <v-data-table
              :headers="namespacesHeaders"
              :items="namespaceData"
              class="elevation-1"
              :loading="loadingTable"
              hide-actions
            >
              <template slot="items" slot-scope="props">
                <td>{{ props.item.name}}</td>
                <td><DeploymentAnnotations :annotations="props.item.deployment_annotations"></DeploymentAnnotations></td>
                <td>
                  <ul v-if="props.item.ingresses">
                    <li v-for="(ingress,n) in props.item.ingresses" :key="n">
                      <a :href="ingress | formatIngressURL">
                        <v-chip v-if="ingress.looks_like_grpc" small>GRPC</v-chip>
                        {{ingress | formatIngressURL }}
                      </a>
                    </li>
                  </ul>
                </td>
                <td>
                  <ul>
                    <li v-for="image in props.item.images" :key="image">
                      <v-tooltip bottom>
                      <span slot="activator">{{image | truncateBegin(25)}}</span>
                      <span>{{image}}</span>
                      </v-tooltip>
                    </li>
                  </ul>
                </td>
                <td>{{ props.item.deployment_status.available_replicas }}/{{ props.item.deployment_status.replicas }}</td>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
  import DeploymentAnnotations from "./DeploymentAnnotations"

  export default {
    components: {
      DeploymentAnnotations
    },
    data: () => ({
        namespaces: [],
        namespacesHeaders: [
            {text: 'Name', value: 'name'},
            {text: 'Deployment annotations', value: 'deployment_annotations'},
            {text: 'Ingresses', value: 'ingresses'},
            {text: 'Images', value: 'images'},
            {text: 'Deployment status', value: 'deployment_status'},
        ],
        namespaceData: [],
        active: null,
        loadingTable: true,
        prodOnly: true,
        error: ''
    }),
    methods: {
      fetchData: function () {
          this.error = ''
          return fetch(location.protocol+"//"+location.hostname+"/api/namespaces")
            .then((res) => {return res.json()})
            .then((data) => data.filter(n => this.prodOnly ? n.endsWith('-prod') : true))
            .then((data) => {this.namespaces = data.sort()})
            .catch((error) => {this.error = error.toString()})
      },
      loadNamespace: function (namespace) {
          this.error = ''
          this.loadingTable=true
          fetch(location.protocol+"//"+location.hostname+"/api/namespace/"+namespace)
            .then((res) => {return res.json()})
            .then((data) => {
                this.namespaceData = data
                this.loadingTable=false
            })
            .catch((error) => {this.error = error.toString()})
      },
    },
    mounted: function () {
      this.fetchData()
        .then(() => {
          const a = this.namespaces.findIndex((namespace) => {this.$route.params.namespace == namespace})
          this.active = a == -1 ? 0 : a
      })
    },
    filters: {
      formatIngressURL: function (ingress) {
        return 'http://' + ingress.host + (ingress.path == "" ? "/" : ingress.path)
      },
      truncateBegin: function (s, len) {
        if (s.length > len) {
          return "..." + s.substring(s.length-len);
        }
        return s
      },
    }
  }
</script>

<style>

</style>
