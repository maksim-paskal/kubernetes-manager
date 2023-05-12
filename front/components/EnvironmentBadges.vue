<template>
  <div v-if="this.badges">
    <span v-bind:key="i" v-for="(item, i) in this.badges" :title="item.Description"
      :class="`badge rounded-pill bg-${item.Key == 'staled' ? 'danger' : 'primary'}`" style="margin-right: 2px">
      {{ item.Key == 'staled' ? item.Key : `${item.Key}: ${item.Value}` }}
    </span>
  </div>
</template>
<script>
export default {
  props: ["badges"]
}
</script>