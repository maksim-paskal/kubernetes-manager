<template>
  <div class="detail-tab">
    <b-alert v-if="$fetchState.error" variant="danger" show>{{
      $fetchState.error.message
    }}</b-alert>
    <b-spinner v-if="$fetchState.pending" variant="primary" />
    <div v-else>
      <b-form-input v-model="dataFilter" autocomplete="off" placeholder="Type to Search" />
      <b-table style="margin-top:5px" striped hover :items="data.Result" :fields="dataFields" :filter="dataFilter">
        <template v-slot:cell(Created)="row">
          <div :title="row.item.Created">{{ row.item.CreatedShort }}&nbsp;ago</div>
        </template>
      </b-table>
    </div>
  </div>
</template>
<script>
export default {
  layout: "details",
  head() {
    return {
      title: this.pageTitle('Events', true)
    }
  },
  data() {
    return {
      data: {},
      dataFilter: null,
      dataFields: [
        { key: "Created" },
        { key: "Type" },
        { key: "Reason" },
        { key: "Object" },
        { key: "Message" },
      ],
    };
  },
  async fetch() {
    const result = await fetch(`/api/${this.$route.params.environmentID}/events`);
    if (result.ok) {
      this.data = await result.json();
    } else {
      const text = await result.text();
      throw Error(text);
    }
  },
}
</script>