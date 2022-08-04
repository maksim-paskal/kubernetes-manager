<template>
  <b-spinner v-if="typeof data.PodsReady === 'undefined'" variant="primary" />
  <div v-else style="margin-left:10px">
    <table>
      <tr>
        <td><em class="bi bi-check-circle-fill text-success" />&nbsp;Pod running: {{ data.PodsReady }}</td>
        <td style="cursor: pointer;" @click="$bvModal.show('bv-modal-failed-pods')" v-if="data.PodsFailed > 0"><em
            class="bi bi-x-circle-fill text-danger" />&nbsp;Pod failed: {{ data.PodsFailed
            }}</td>
      </tr>
      <tr>
        <td colspan="2">
          <em title="requested CPU" class="bi bi-cpu" />&nbsp;{{ data.CPURequests }}
          <em title="requested memory" class="bi bi-memory" />&nbsp;{{ data.MemoryRequests }}
          <em title="requested storage" class="bi bi-hdd" />&nbsp;{{ data.StorageRequests }}
        </td>
      </tr>
    </table>
    <b-modal size="xl" id="bv-modal-failed-pods" title="Failed pods" ok-only>
      <ul>
        <li v-bind:key="index" v-for="(item, index) in data.PodsFailedName">
          {{ item }}
        </li>
      </ul>
    </b-modal>
  </div>
</template>
<style scoped>
em {
  margin-left: 5px
}
</style>
<script>
export default {
  data() {
    return {
      data: {}
    }
  },
  mounted() {
    setInterval(() => {
      this.$fetch()
    }, 5000);
  },
  async fetch() {
    const result = await fetch(`/api/${this.$route.params.environmentID}/pods`);
    if (result.ok) {
      const data = await result.json();
      this.data = data.Result
    } else {
      const text = await result.text();
      throw Error(text);
    }
  }
}
</script>
