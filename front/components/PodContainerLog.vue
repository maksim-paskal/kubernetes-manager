<template>
  <div>
    <div style="padding: 10px">
      <b-button @click="reload()">reload</b-button>
      <b-button @click="timestamp()">{{ timestamps ? "hide" : "show" }} timestamps</b-button>
    </div>
    <div style="overflow: scroll; padding: 10px; height: 500px; background-color: black;color:white">
      <b-spinner v-if="$fetchState.pending" variant="primary" />
      <pre v-else>{{ data }}</pre>
      <div ref="bottomEl"></div>
    </div>
  </div>
</template>
<script>
export default {
  props: ["pod", "container"],
  watch: {
    timestamps() {
      this.$fetch();
    }
  },
  data() {
    return {
      timestamps: false,
      data: {},
    };
  },
  async fetch() {
    this.data = {}

    const result = await fetch(`/api/${this.$route.params.environmentID}/pod-container-logs?pod=${this.pod}&container=${this.container}&timestamps=${this.timestamps}`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result

      setTimeout(() => {
        this.scrollToBottom()
      }, 100)
    }
  },
  methods: {
    reload() {
      this.$fetch()
    },
    scrollToBottom() {
      this.$refs.bottomEl?.scrollIntoView({ behavior: 'smooth' });
    },
    timestamp() {
      this.timestamps = !this.timestamps
    }
  }
}
</script>
