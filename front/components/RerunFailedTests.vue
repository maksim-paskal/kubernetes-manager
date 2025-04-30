<template>
  <div>
    <b-alert v-if="failedTestError" style="margin-bottom:10px" variant="danger" show>{{ failedTestError }}</b-alert>
    <div v-if="this.customAction">
      <b-form-input placeholder="Description" v-model="failedTestDescription" />
      <DropDown ref="failedTestBranch" style="margin-top:10px;" id="failedTestBranch" text="Select branch"
        :endpoint="`/api/project-refs?id=${this.customAction.ProjectID}`" :value="failedTestRef" />
      <b-form-select v-model="failedTestType" style="margin-top:10px;" class="form-select"
        :options="this.customAction.Tests" />
    </div>
    <textarea style="margin-top:10px;width:100%;height:300px" readonly v-model="failedTestsXML" />
  </div>
</template>
<script>
export default {
  props: ["customAction", "params"],
  data() {
    return {
      failedTestError: "",
      failedTestRef: "",
      failedTestType: "",
      failedTestsXML: "",
      failedTestDescription: "",
    }
  },
  mounted() {
    this.failedTestRef = this.params.Ref;
    this.failedTestType = this.params.Type;
    this.failedTestDescription = this.params.Description;

    // group the object by class name
    const groupedByProject = this.params.failedTests.reduce((acc, line) => {
      const parts = line.split("->");
      const methodName = parts[parts.length - 1];
      const className = parts.slice(0, parts.length - 1).join(".");

      if (!acc[className]) {
        acc[className] = [];
      }

      acc[className].push(methodName);

      return acc;
    }, {});

    // format the object to XML
    const doc = document.implementation.createDocument("", "", null);

    const suite = doc.createElement("suite");
    suite.setAttribute("name", "Rerun Failed Tests");
    suite.appendChild(doc.createElement("listeners"));

    const test = doc.createElement("test");
    test.setAttribute("name", "Tests");
    test.setAttribute("parallel", "classes");
    test.setAttribute("thread-count", "6");
    suite.appendChild(test);

    const classes = doc.createElement("classes");
    test.appendChild(classes);

    Object.keys(groupedByProject).map(key => {
      const classNode = doc.createElement("class");
      classNode.setAttribute("name", key);
      const methods = doc.createElement("methods");

      groupedByProject[key].forEach(methodName => {
        const include = doc.createElement("include");
        include.setAttribute("name", methodName);
        methods.appendChild(include);
      });

      classNode.appendChild(methods);
      classes.appendChild(classNode);
    });

    doc.appendChild(suite);

    const serializer = new XMLSerializer();
    this.failedTestsXML = serializer.serializeToString(doc);

    this.failedTestsXML = '<!DOCTYPE suite SYSTEM "https://testng.org/testng-1.0.dtd" >\n'
      + this.failedTestsXML.replaceAll(">", ">\n")
  },
  methods: {
    getArgs() {
      if (!this.$refs.failedTestBranch.selected) {
        this.failedTestError = "Please select a branch";
        return null;
      }
      if (!this.failedTestType) {
        this.failedTestError = "Please select a test";
        return null;
      }

      return {
        Ref: this.$refs.failedTestBranch.selected,
        Test: this.failedTestType,
        ExtraEnv: {
          "CUSTOM_ACTION": "true",
          "AUTOTEST_DESCRIPTION": this.failedTestDescription,
          "AUTOTEST_XML_FILE@FILE": this.failedTestsXML,
        },
      }
    },
  },
}
</script>