<!--
<template>
  <v-list v-if="annotations">
    <v-list-tile v-for="(value, key, id) in annotations" :key="id">
      <v-list-tile-content>
        <v-list-tile-title>{{key}}</v-list-tile-title>
        <v-list-tile-sub-title v-if="isURL(value)"><a :href="value">{{value}}</a></v-list-tile-sub-title>
        <v-list-tile-sub-title v-else>{{value}}</v-list-tile-sub-title>
      </v-list-tile-content>
    </v-list-tile>
  </v-list>
</template>
//-->
<template>
  <ul v-if="annotations">
    <li v-for="(value, key, id) in annotations" :key="id">
        <b>{{key}}</b>:
        <span v-if="isURL(value)"><a :href="value">{{value}}</a></span>
        <span v-else>{{value}}</span>
    </li>
  </ul>
</template>

<script>
  export default {
    props: ['annotations'],
    methods: {
      //transform: function(obj) {
      //  return obj == null ? null : Object.keys(obj).map((key) => ({key: key, value: obj[key]}))
      //},
      isURL: function (url) {
        return (url.startsWith('https://') || url.startsWith('http://'))
      },
    },
  }
</script>
