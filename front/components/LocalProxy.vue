<style scoped>
.form-field {
  margin-bottom: 10px;
}
</style>
<template>
  <b-alert v-if="$fetchState.error" variant="danger" show>{{
    $fetchState.error.message
  }}</b-alert>
  <b-spinner v-else-if="$fetchState.pending" variant="primary" />
  <div v-else>
    <b-form-select v-model="selectedHost" class="form-select form-field" :options="selectedHostOptions" />
    <b-form-select v-model="selectedOS" class="form-select form-field" :options="selectedOSOptions" />
    <b-form-select v-model="selectedTemplate" class="form-select form-field" :options="selectedTemplateOptions" />
    <div v-if="templatedText" style="margin-top:20px">
      <CopyTextbox :text="templatedText" height="200px" />
      <a href="https://github.com/maksim-paskal/developer-proxy"
        target="_blank">https://github.com/maksim-paskal/developer-proxy</a>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  mounted() {
    if (this.environment?.Hosts) {
      this.selectedHostOptions = this.environment.Hosts
      this.selectedHost = this.environment.Hosts[0]
    }
  },
  watch: {
    selectedOS() {
      this.template();
    },
    selectedHost() {
      this.template();
    },
    selectedTemplate() {
      this.template();
    },
  },
  data() {
    return {
      data: {},
      selectedHost: "",
      selectedHostOptions: [],
      selectedOS: "Linux",
      selectedOSOptions: ['Linux', 'MacOS'],
      templatedText: "",
      selectedTemplate: "",
      selectedTemplateOptions: [],
      selectedTemplateOptionsData: []
    }
  },
  methods: {
    template() {
      if (!this.selectedHost) return
      if (!this.selectedOS) return
      if (!this.selectedTemplate) return

      const data = this.selectedTemplateOptionsData.find((template) => template.Title === this.selectedTemplate)

      this.templatedText = ""

      // https://github.com/maksim-paskal/developer-proxy
      const developerProxyCmd = this.selectedOS == 'MacOS' ? 'developer-proxy' : "docker run --pull=always --rm -it --net=host paskalmaksim/developer-proxy:latest"

      let text = data.Template.trim()
      text = text.replace(/{{cmd}}/g, developerProxyCmd)
      text = text.replace(/{{host}}/g, this.selectedHost)
      text = text.replace(/{{os}}/g, data.selectedOS)
      text = text.replace(/{{template}}/g, data.selectedTemplate)

      this.templatedText = text

      console.log(this.templatedText)
    }
  },
  async fetch() {
    const result = await fetch(`/api/${this.$route.params.environmentID}/local-proxy-templates`);
    if (result.ok) {
      this.data = await result.json();

      this.templatedText = ""
      this.selectedTemplate = ""
      this.selectedTemplateOptions = []
      this.selectedTemplateOptionsData = []

      this.data.Result.forEach((template) => {
        this.selectedTemplateOptions.push(template.Title)
        this.selectedTemplateOptionsData.push(template)
      })

      if (this.selectedTemplateOptions.length > 0) {
        this.selectedTemplate = this.selectedTemplateOptions[0]
      }

      this.template()
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
  head() {
    return {
      title: this.pageTitle('Local proxy', true)
    }
  },
}
</script>