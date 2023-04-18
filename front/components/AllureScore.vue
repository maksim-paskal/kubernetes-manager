<template>
  <div>
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <div v-else-if="$fetchState.pending" class="text-center">
      <b-spinner variant="primary" />
    </div>
    <div v-else>
      <div v-if="this.results > 0" :style="this.testStyle">{{ this.results.toFixed(2) }}%</div>
      <div v-if="this.showFailedTests">
        <b-button v-if="failedTestsFormated.length > 0" size="sm" variant="warning"
          @click="buttonFailedTests = !buttonFailedTests">{{ buttonFailedTests ? "Hide" : "Show" }} failed
          tests</b-button>
        <div v-if="buttonFailedTests">
          <div v-for="item in failedTestsFormated" :key="item">{{ item }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  props: ["allureResults", "variant", "showFailedTests"],
  data() {
    return {
      data: [],
      total: 0,
      success: 0,
      results: 0,
      testStyle: "",
      failedTests: [],
      buttonFailedTests: false
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
  async fetch() {
    const packageUrl = this.allureResults.replace(/\/index.html$/, '') + '/data/packages.json';

    const getTestResults = (parent, item) => {
      if (!item || !item.children) {
        return;
      }

      if (Array.isArray(item.children)) {
        item.children.forEach((item) => {
          if (Array.isArray(item.children)) {
            getTestResults(parent + ";" + item.name, item);
          } else {
            this.total++;
            if (item.status === 'passed') {
              this.success++;
            } else {
              this.failedTests.push(parent + ";" + item.name);
            }
          }
        });
      } else {
        this.total++;
        if (item.status === 'passed') {
          this.success++;
        } else {
          this.failedTests.push(parent + ";" + item.name);
        }
      }
    }

    const result = await fetch(packageUrl);
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
