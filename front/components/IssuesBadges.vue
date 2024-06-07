<template>
  <div>
    <span>issues</span>
    <b-spinner v-if="$fetchState.pending" small variant="primary" />
    <div v-else style="display: inline-block;">
      <span v-if="newErrors > 0" class="badge-margin badge rounded-pill bg-danger">{{ newErrors }}</span>
      <span v-if="warnings > 0" class="badge-margin badge rounded-pill bg-warning">{{ warnings }}</span>
    </div>
  </div>
</template>
<script>
export default {
  props: ["environmentID"],
  watch: {
    environmentID() {
      this.$fetch();
    }
  },
  data() {
    return {
      data: {},
      warnings: 0,
      newErrors: 0,
    }
  },
  async fetch() {
    const result = await fetch(`/api/${this.environmentID}/issues`);
    if (result.ok) {
      this.data = await result.json();
      this.data.Result.Items.forEach((item) => {
        if (item.Level === "warning") {
          this.warnings += 1;
        } else if (item.Level === "error") {
          this.newErrors += 1;
        }
      });
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
}
</script>