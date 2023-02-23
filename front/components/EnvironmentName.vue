<template>
  <div>
    <b-alert v-if="errorText" variant="danger" show>{{ errorText }}</b-alert>
    <b-spinner size="sm" variant="primary" v-if="!environmentName"></b-spinner>
    <b-form v-else-if="saveMode" @submit="save" style="display:flex;width:600px;">
      <b-form-input v-model="newNamespaceName" required />
      &nbsp;<b-button type="submit" variant="primary">save</b-button>
      &nbsp;<b-button @click="cancel()">cancel</b-button>
    </b-form>
    <div v-else style="display:flex;align-items: center;">
      <h4>{{ environmentName }}</h4>&nbsp;
      <CopyIcon :text="environmentName" />
      <b-button title="edit" variant="link" size="sm" @click="saveMode = true; newNamespaceName = environmentName"><em
          class="bi bi-pencil" />
      </b-button>
      <Liked :environmentID="this.$route.params.environmentID" />
    </div>
    <div style="font-size:10pt" v-if="!saveMode && environmentName != this.environment.Namespace">{{
      this.environment.Namespace }}</div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      saveMode: false,
      newNamespaceName: "",
    }
  },
  computed: {
    environmentName() {
      const name = this.namespaceAnnotation(this.const().LabelEnvironmentName)
      if (name) {
        return name
      }

      return this.environment.Namespace
    }
  },
  methods: {
    cancel() {
      this.saveMode = false
      this.errorText = "";
      this.infoText = "";
    },
    async save(event) {
      event.preventDefault()
      this.errorText = "";
      this.infoText = "";
      try {
        if (this.newNamespaceName === this.environmentName) {
          this.saveMode = false
          this.errorText = "unchanged name";
          return
        }

        let meta = { Annotations: {} }
        meta.Annotations[this.const().LabelEnvironmentName] = this.newNamespaceName

        await this.call("make-save-namespace-name", { Name: this.newNamespaceName })
        if (!this.errorText) {
          this.loadEnvironment(this.$route.params.environmentID)
          this.cancel();
        }
      } catch (e) {
        console.log(e)
      }
    }
  }
}
</script>