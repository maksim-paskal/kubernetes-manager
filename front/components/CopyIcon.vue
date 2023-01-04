<template>
  <b-button @click="copyToClipboard()" variant="link" size="sm" title="copy"><em :class=iconClass /></b-button>
</template>
<script>
export default {
  props: ['text', "variant"],
  data() {
    return {
      iconClass: `bi bi-clipboard ${this.variant}`,
    };
  },
  methods: {
    copyToClipboard() {
      navigator.clipboard.writeText(this.text);
      this.iconClass = `bi bi-clipboard-check-fill ${this.variant}`;
      setTimeout(() => {
        this.iconClass = `bi bi-clipboard ${this.variant}`;
      }, 3000);
    },
  },
}
</script>