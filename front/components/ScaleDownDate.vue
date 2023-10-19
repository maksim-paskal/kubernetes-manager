<template>
  <div>branch will be paused <strong>{{ getScaleDownDelay() }}</strong> your local time</div>
</template>
<script>
export default {
  methods: {
    getScaleDownDelay() {
      const lang = navigator.language | window.navigator.language
      const scaleDownDelay = this.namespaceAnnotation('kubernetes-manager/scaleDownDelay')
      let scaleDownDelayDate = Date.parse(scaleDownDelay)

      if (isNaN(scaleDownDelayDate) || scaleDownDelayDate < Date.now()) {
        scaleDownDelayDate = Date.now()
      }

      const isToday = (new Date().toDateString() == new Date(scaleDownDelayDate).toDateString());

      if (isToday) {
        return `Today at ${new Intl.DateTimeFormat(lang, { timeStyle: 'medium' }).format(scaleDownDelayDate)}`
      }

      const options = { dateStyle: 'full', timeStyle: 'medium' };

      return new Intl.DateTimeFormat(lang, options).format(scaleDownDelayDate)
    },
  },
}
</script>