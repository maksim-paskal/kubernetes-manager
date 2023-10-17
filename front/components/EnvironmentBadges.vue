<template>
  <div v-if="this.badges">
    <span v-bind:key="i" v-for="(item, i) in this.badges" :title="item.Description"
      :class="`badge rounded-pill bg-${getBadgeClass(item.Key)}`" style="margin-right: 2px">
      {{ item.Key == 'staled' || item.Key == 'scaledown' ? item.Key : `${item.Key}: ${item.Value}` }}
    </span>
  </div>
</template>
<script>
export default {
  props: ["badges"],
  methods: {
    getBadgeClass(key) {
      if (key == 'staled') return "danger"
      if (key == 'scaledown') return "warning text-dark"
      return 'primary';
    }
  }
}
</script>