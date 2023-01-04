<template>
  <div>
    <b-button @click="copyToClipboard()" variant="outline-primary" style="margin-bottom: 10px"><em
        :class=iconClass />&nbsp;Copy to clipboard
    </b-button>
    <textarea class="form-control" onclick="this.focus();this.select()"
      style="background-color:#eeeeee;border:0px;padding:10px;outline:none;width:100%" v-model="textareaValue"
      readonly />
  </div>
</template>
<script>
export default {
  props: ['text'],
  mounted() {
    this.textareaValue = this.text;
  },
  data() {
    return {
      textareaValue: '',
      iconClass: 'bi bi-clipboard',
    };
  },
  methods: {
    copyToClipboard() {
      navigator.clipboard.writeText(this.textareaValue);
      this.iconClass = 'bi bi-clipboard-check-fill';
    },
  },
}
</script>
