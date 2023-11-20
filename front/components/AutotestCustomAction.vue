<template>
  <div>
    <DropDown ref="autotestBranch" id="autotestBranch" text="Select branch"
      :endpoint="`/api/project-refs?id=${this.customAction.ProjectID}`" />
    <ul style="list-style-type: none;padding-left: 0px;">
      <li style="margin-top: 10px">
        <pre style="margin-bottom: 0px">Autotest type</pre>
        <b-form-select v-model="test" class="form-select" :options="this.customAction.Tests" />
      </li>
      <li style="margin-top: 10px" v-bind:key="index" v-for="(env, index) in this.customAction.Env">
        <pre style="margin-bottom: 0px">{{ env.Description }}</pre>
        <b-form-input class="autotest-env" :name="env.Name" :value="env.Default" />
      </li>
    </ul>
  </div>
</template>
<script>
export default {
  props: ["customAction"],
  data() {
    return {
      test: "",
    }
  },
  mounted() {
    if (this.customAction.Tests.length == 1) {
      this.test = this.customAction.Tests[0];
    }
  },

  methods: {
    getCustomActionInput() {
      const inputCollection = document.getElementsByClassName("autotest-env");

      let selectedEnv = {
        CUSTOM_ACTION: "true",
      };

      for (let i = 0; i < inputCollection.length; i++) {
        selectedEnv[inputCollection[i].name] = inputCollection[i].value;
      }

      return {
        Ref: this.$refs.autotestBranch.selected,
        Test: this.test,
        Force: true,
        ExtraEnv: selectedEnv,
      }
    }
  }
}
</script>