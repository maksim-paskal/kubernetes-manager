<template>
  <div>
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-alert v-else-if="reportNotFound" variant="warning" show>
      <b-button class="bi bi-wrench-adjustable" variant="outline-primary"
        @click="showReportNotFound = !showReportNotFound">&nbsp;report not found</b-button>
      <ul v-if="showReportNotFound" style="padding-top: 10px">
        <li>Some errors occurred while running autotests. Open the pipeline for details.</li>
        <li>If the autotest was run a week ago, it has been deleted.</li>
      </ul>
    </b-alert>
    <div v-else-if="$fetchState.pending" class="text-center">
      <b-spinner variant="primary" />
    </div>
    <div v-else>
      <div v-if="this.results > 0" :style="this.testStyle">{{ this.results.toFixed(2) }}%</div>
      <div v-if="this.showFailedTests && failedTestsFormated.length > 0">
        <b-button size="sm" variant="warning" @click="buttonFailedTests = !buttonFailedTests">{{ buttonFailedTests ?
          "Hide" : "Show" }} failed
          tests</b-button>
        <b-button size="sm" @click="showFailedTestDialog">Rerun failed tests</b-button>
        <div v-if="buttonFailedTests">
          <div v-for="item in failedTestsFormated" :key="item">{{ item }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  props: ["item", "variant", "showFailedTests"],
  data() {
    return {
      data: [],
      total: 0,
      success: 0,
      results: 0,
      testStyle: "",
      failedTests: [],
      buttonFailedTests: false,
      reportNotFound: false,
      showReportNotFound: false
    }
  },
  computed: {
    failedTestsFormated() {
      let formated = [];

      this.failedTests.forEach((item) => {
        let text = item.replace(/^;/, '');
        text = text.replaceAll(';', '->');

        formated.push(text);
      });

      formated.sort();

      return formated;
    }
  },
  methods: {
    showFailedTestDialog() {
      this.$parent.showFailedTestDialog({
        Ref: this.item.PipelineRef,
        Type: this.item.Test,
        Description: "Rerun failed tests for report " + this.item.ResultURL,
        failedTests: this.failedTestsFormated,
      });
    },
  },
  async fetch() {
    const packageUrl = this.item.ResultURL.replace(/\/index.html$/, '') + '/data/packages.json';

    const getTestResults = (parent, item) => {
      if (!item || !item.children) {
        return;
      }

      const calc = (item) => {
        if (item.status === 'unknown') {
          return;
        }
        this.total++;
        if (item.status === 'passed') {
          this.success++;
        } else {
          this.failedTests.push(parent + ";" + item.name);
        }
      }

      if (Array.isArray(item.children)) {
        item.children.forEach((item) => {
          if (Array.isArray(item.children)) {
            getTestResults(parent + ";" + item.name, item);
          } else {
            calc(item);
          }
        });
      } else {
        calc(item);
      }
    }

    const result = await fetch(packageUrl);
    if (result.status === 404) {
      this.reportNotFound = true;
      return;
    }

    if (result.ok) {
      this.data = await result.json();

      getTestResults("", this.data);

      this.results = 100 * this.success / this.total;

      if (this.variant == "large") {
        this.testStyle = "font-size: 50pt";
      }
    } else {
      const text = await result.text();
      throw Error(text);
    }
  }
}
</script>
